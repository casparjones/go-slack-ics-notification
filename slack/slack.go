package slack

import (
	"bytes"
	"encoding/json"
	"github.com/apognu/gocal"
	"io"
	"log"
	"net/http"
	"os"
)

type Slack struct {
	Message string
	User    string
}

func (s *Slack) toJSON(v interface{}) string {
	marshal, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Fehler beim Umwandeln in JSON: %v", err)
	}
	return string(marshal)
}

func (s *Slack) SendCalenderEvent(e gocal.Event, user string) {
	msg := SlackMessage{
		Channel: user,
		Blocks: []Block{
			{
				Type: "header",
				Text: &Text{
					Type: "plain_text",
					Text: e.Start.Format("02.01.2006") + " » " + e.Summary,
				},
			},
			{
				Type: "section",
				Text: &Text{
					Type: "mrkdwn",
					Text: e.Description,
				},
			},
		},
	}

	payload := s.toJSON(msg)
	s.sendPayloadToSlack([]byte(payload))
}

func (s *Slack) sendPayloadToSlack(payload []byte) string {
	url := "https://slack.com/api/chat.postMessage"
	token := os.Getenv("SLACK_TOKEN")

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("Fehler beim Erstellen der Anforderung: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Fehler beim Senden der Anforderung: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Fehler beim Lesen der Antwort: %v", err)
	}

	return string(body)
}

func (s *Slack) Send() {
	o := make(map[string]string)

	o["text"] = s.Message
	o["channel"] = s.User
	payload := s.toJSON(o)
	s.sendPayloadToSlack([]byte(payload))
}

func (s *Slack) SendMessage(channel string, user string, message SlackMessage) string {
	message.User = user
	message.Channel = channel
	payload := s.toJSON(message)
	s.sendPayloadToSlack([]byte(payload))

	return "send"
}

var Instance Slack
