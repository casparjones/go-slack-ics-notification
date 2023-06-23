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
	msg := Message{
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
	s.PostMessage([]byte(payload))
}

func (s *Slack) PostMessage(payload []byte) Response {
	url := "https://slack.com/api/chat.postMessage"
	return s.sendPayload(url, payload)
}

func (s *Slack) changeMessage(payload []byte) Response {
	url := "https://slack.com/api/chat.update"
	return s.sendPayload(url, payload)
}

func (s *Slack) sendPayload(url string, payload []byte) Response {
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
		log.Printf("Fehler beim Senden der Anforderung: %v", err)
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

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("response error: %s", err.Error())
	}

	return response
}

func (s *Slack) Send() {
	o := make(map[string]string)

	o["text"] = s.Message
	o["channel"] = s.User
	payload := s.toJSON(o)
	s.PostMessage([]byte(payload))
}

func (s *Slack) SendMessage(channel string, user string, message Message) Response {
	message.User = user
	message.Channel = channel
	payload := s.toJSON(message)
	response := s.PostMessage([]byte(payload))

	return response
}

func (s *Slack) ChangeMessage(ts string, channel string, user string, message Message) Response {
	message.User = user
	message.Channel = channel
	message.TimeStamp = ts
	payload := s.toJSON(message)
	response := s.changeMessage([]byte(payload))

	return response
}

var Instance Slack
