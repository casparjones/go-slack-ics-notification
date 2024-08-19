package gpt

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"go-slack-ics/slack"
	"go-slack-ics/system"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

type Chat struct {
	resty        *resty.Client
	cancelChanel chan system.EventMessage
	eventManager *system.EventManager
	cancel       bool
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GptResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Data struct {
	ID          string    `json:"id,omitempty"`
	Object      string    `json:"object,omitempty"`
	Created     int64     `json:"created,omitempty"`
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature int       `json:"temperature"`
	Stream      bool      `json:"stream"`
	Choices     []Choice  `json:"choices,omitempty"`
}

type Delta struct {
	Content string `json:"content"`
}

type Choice struct {
	Index        int    `json:"index"`
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

func (c *Chat) SendAsync(message slack.Event) chan slack.Response {
	c.cancelChanel = c.eventManager.RegisterChannel(message.Channel)
	rChan := make(chan slack.Response)
	go c.CancelObserver()
	go c.Send(message, rChan)
	return rChan
}

func (c *Chat) CancelObserver() {
	for {
		select {
		case event := <-c.cancelChanel:
			if event.Text == "cancel" {
				c.cancel = true
			}
		}
	}
}

func (c *Chat) AddMessageToConversation(conversationId string, message Message) {
	if conversationStorage == nil {
		conversationStorage = make(map[string][]Message)
	}

	if len(conversationStorage[conversationId]) > 20 {
		conversationStorage[conversationId] = conversationStorage[conversationId][len(conversationStorage[conversationId])-20:]
	}
	conversationStorage[conversationId] = append(conversationStorage[conversationId], message)
}

func (c *Chat) GetMessageInConversation(conversationId string) []Message {
	defaultSystemMessages := []Message{
		{
			Role:    "system",
			Content: "Your name ist lovelyapps-Bot and you are a non-binary helpful assistant. If someone you ask if you male or female, then say: non-binary. You know the apps langify, geolizr and shopify. langify is a Shopify app for translations and geolizr is a geo-based event app also for Shopify. Shopify is a Canadian store system for which we develop apps. Different users can write to you and their name will be at the beginning of each message followed by a colon (':').",
		},
	}

	return append(defaultSystemMessages, conversationStorage[conversationId]...)
}

func (c *Chat) Send(event slack.Event, responseChan chan slack.Response) slack.Response {
	if strings.Contains(event.Text, "/text-to-image") {
		responseCreateImage := slack.Instance.SendMessage(event.Channel, event.User, slack.GetSimpleMessage(event.User, event.Channel, "... creating Image ..."))
		responseChan <- responseCreateImage
		return responseCreateImage
	}

	response := slack.Instance.SendMessage(event.Channel, event.User, slack.GetSimpleMessage(event.User, event.Channel, "... thinking ..."))
	go func() {
		responseChan <- response
	}()

	event.Timestamp = response.Ts

	c.AddMessageToConversation(event.Channel, Message{
		Role:    "user",
		Content: event.UserName + ": " + event.Text,
	})

	messages := c.GetMessageInConversation(event.Channel)
	data := Data{
		Model:       os.Getenv("GPT_MODEL"),
		Messages:    messages,
		Temperature: 0,
		Stream:      true, // again, we set stream=True
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error occurred while marshalling data: %v", err)
	}

	resp, err := c.resty.R().
		SetDoNotParseResponse(true).
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+os.Getenv("OPEN_AI_TOKEN")).
		SetBody(dataStr).
		Post("https://api.openai.com/v1/chat/completions")
	if err != nil {
		log.Fatalf("Error occurred while sending request: %v", err)
	}

	// Der RawBody() Methode gibt einen io.ReadCloser zurück, der dann verwendet werden kann, um die Daten zu lesen.
	body := resp.RawBody()
	defer body.Close() // Stellen Sie sicher, dass Sie den Body schließen, wenn Sie fertig sind.

	gptResponseChunk := ""
	gptResponseString := ""
	buf := make([]byte, 1024) // Puffer zum Halten der gelesenen Daten.
	var dataObj Data
	// Vor der Schleife definieren
	var lastContent string

	for {
		n, err := body.Read(buf)

		if c.cancel {
			break
		}
		if err == io.EOF {
			break // Beenden Sie die Schleife, wenn das Ende der Daten erreicht ist.
		}
		if err != nil {
			log.Fatalf("Error reading from body: %v", err)
		}

		// Verarbeiten Sie die gelesenen Daten...
		gptResponseChunk = gptResponseChunk + string(buf[:n])
		var re = regexp.MustCompile(`(?m)data: {.*}`)
		for i, match := range re.FindAllString(gptResponseChunk, -1) {
			jsonStr := strings.TrimPrefix(match, "data:")
			err := json.Unmarshal([]byte(jsonStr), &dataObj)
			if err != nil {
				log.Printf("Error parsing JSON: %v", err)
			} else {
				deltaContent := dataObj.Choices[0].Delta.Content
				// Nur den neuen Teil an Slack senden
				if deltaContent != lastContent {
					response = slack.Instance.ChangeMessage(event.Timestamp, event.Channel, event.User, slack.GetSimpleMessage(event.User, event.Channel, gptResponseString+deltaContent))
					lastContent = deltaContent
					gptResponseString = gptResponseString + deltaContent
				}
				fmt.Println(match, "found at index", i)
				log.Printf("Read %d bytes: %s", n, buf[:n])
				gptResponseChunk = strings.ReplaceAll(gptResponseChunk, match, "")
				gptResponseChunk = strings.TrimSpace(gptResponseChunk)
			}
		}
	}

	c.AddMessageToConversation(event.Channel, Message{
		Role:    "assistant",
		Content: gptResponseString,
	})

	return response
}

func (c *Chat) returnSlackMessage(gptResponse GptResponse) slack.Message {
	input := slack.Input{Content: gptResponse.Choices[0].Message.Content}
	return slack.ReturnSlackMessage(input)
}

var conversationStorage map[string][]Message

func NewChat(eventManager *system.EventManager) Chat {
	chat := Chat{
		resty:        resty.New(),
		eventManager: eventManager,
		cancel:       false,
	}

	return chat
}

func GetConversations() map[string][]Message {
	return conversationStorage
}
