package gpt

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"go-slack-ics/slack"
	"log"
	"os"
)

type Event struct {
	Token       string `json:"token"`
	TeamID      string `json:"team_id"`
	TeamDomain  string `json:"team_domain"`
	Channel     string `json:"channel"`
	ChannelName string `json:"channel_name"`
	User        string `json:"user"`
	UserName    string `json:"user_name"`
	Commands    string `json:"commands"`
	Text        string `json:"text"`
	Timestamp   string `json:"ts,omitempty"`
}

type Chat struct {
	resty *resty.Client
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
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

func (c *Chat) SendAsync(message Event) chan slack.Response {
	rchan := make(chan slack.Response)
	go c.Send(message, rchan)
	return rchan
}

func (c *Chat) AddMessageToConversation(conversationId string, message Message) {
	if conversationStorage == nil {
		conversationStorage = make(map[string][]Message)
	}

	if len(conversationStorage[conversationId]) > 20 {
		conversationStorage[conversationId] = []Message{}
	}
	conversationStorage[conversationId] = append(conversationStorage[conversationId], message)
}

func (c *Chat) GetMessageInConversation(conversationId string) []Message {
	defaultSystemMessages := []Message{
		{
			Role:    "system",
			Content: "Your name ist lovelyapps-Bot and you are a female helpful assistant. If something you ask if you male or female, then say: female. You know the apps langify, geolizr and shopify. langify is a Shopify app for translations and geolizr is a geo-based event app also for Shopify. Shopify is a Canadian store system for which we develop apps.",
		},
	}

	return append(defaultSystemMessages, conversationStorage[conversationId]...)
}

func (c *Chat) Send(event Event, responseChan chan slack.Response) slack.Response {
	response := slack.Instance.SendMessage(event.Channel, event.User, slack.GetSimpleMessage(event.User, event.Channel, "... thinking ..."))
	go func() {
		responseChan <- response
	}()

	event.Timestamp = response.Ts

	c.AddMessageToConversation(event.Channel, Message{
		Role:    "user",
		Content: event.Text,
	})

	messages := c.GetMessageInConversation(event.Channel)
	data := Data{
		Model:    "gpt-3.5-turbo",
		Messages: messages,
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error occurred while marshalling data: %v", err)
	}

	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+os.Getenv("OPEN_AI_TOKEN")).
		SetBody(dataStr).
		Post("https://api.openai.com/v1/chat/completions")
	if err != nil {
		log.Fatalf("Error occurred while sending request: %v", err)
	}

	gptResponseString := resp.String()
	var gptResponse GptResponse
	err = json.Unmarshal([]byte(gptResponseString), &gptResponse)
	if err != nil {
		log.Fatalf("Error occurred while unmarshalling GPT-3 response: %v", err)
	}

	c.AddMessageToConversation(event.Channel, Message{
		Role:    "assistant",
		Content: gptResponse.Choices[0].Message.Content,
	})

	slackMessage := c.returnSlackMessage(gptResponse)

	if event.Timestamp == "" {
		return slack.Instance.SendMessage(event.Channel, event.User, slackMessage)
	} else {
		return slack.Instance.ChangeMessage(event.Timestamp, event.Channel, event.User, slackMessage)
	}

}

func (c *Chat) returnSlackMessage(gptResponse GptResponse) slack.Message {
	// Implement your logic to transform the gptResponse into a SlackMessage.
	answer := gptResponse.Choices[0].Message.Content
	text := slack.Text{
		Type: "mrkdwn",
		Text: answer,
	}

	block := slack.Block{
		Type: "section",
		Text: &text,
	}

	return slack.Message{
		Color:  "#f2c744",
		Blocks: []slack.Block{block},
	}
}

var conversationStorage map[string][]Message

func NewChat() Chat {
	chat := Chat{
		resty: resty.New(),
	}

	return chat
}

func GetConversations() map[string][]Message {
	return conversationStorage
}
