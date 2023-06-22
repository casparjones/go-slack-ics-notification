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

func (c *Chat) SendAsync(message Event) bool {
	go c.Send(message)
	return true
}

func (c *Chat) Send(event Event) string {
	data := Data{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: event.Text,
			},
		},
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
	slackMessage := c.returnSlackMessage(gptResponseString)

	return slack.Instance.SendMessage(event.Channel, event.User, slackMessage)
}

func (c *Chat) returnSlackMessage(gptResponse string) slack.SlackMessage {
	var response GptResponse
	err := json.Unmarshal([]byte(gptResponse), &response)
	if err != nil {
		log.Fatalf("Error occurred while unmarshalling GPT-3 response: %v", err)
	}

	// Implement your logic to transform the gptResponse into a SlackMessage.
	answer := response.Choices[0].Message.Content
	text := slack.Text{
		Type: "mrkdwn",
		Text: answer,
	}

	block := slack.Block{
		Type: "section",
		Text: &text,
	}

	return slack.SlackMessage{
		Color:  "#f2c744",
		Blocks: []slack.Block{block},
	}
}

func NewChat() Chat {
	chat := Chat{
		resty: resty.New(),
	}

	return chat
}
