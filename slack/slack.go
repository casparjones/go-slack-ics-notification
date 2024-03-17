package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/apognu/gocal"
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

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
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

func (s *Slack) SendImageToSlack(fileBytes []byte, fileName, message, channel string) error {
	token := os.Getenv("SLACK_TOKEN")

	// Erstelle einen Buffer und einen multipart writer
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Füge die Datei hinzu. Ersetze 'file' durch einen geeigneten Dateinamen.
	filePart, err := multipartWriter.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}

	// Schreibe die Byte-Daten in das filePart
	if _, err := filePart.Write(fileBytes); err != nil {
		return err
	}

	// Füge die anderen Felder hinzu
	if err := multipartWriter.WriteField("initial_comment", message); err != nil {
		return err
	}
	if err := multipartWriter.WriteField("channels", channel); err != nil {
		return err
	}

	// Wichtig: Schließe den multipart writer, um das Ende des Formulars zu signalisieren
	if err := multipartWriter.Close(); err != nil {
		return err
	}

	// Erstelle die Anfrage
	req, err := http.NewRequest("POST", "https://slack.com/api/files.upload", &requestBody)
	if err != nil {
		return err
	}

	// Setze den Authorization Header und Content-Type
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Sende die Anfrage
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Überprüfe den Status der Antwort
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fehler beim Hochladen der Datei: %s", resp.Status)
	}

	// Optional: Antwort von Slack verarbeiten

	return nil
}

func ReturnSlackMessage(inputString Input) Message {
	// Implement your logic to transform the inputString into a SlackMessage.
	answer := inputString.Content
	text := Text{
		Type: "mrkdwn",
		Text: answer,
	}

	block := Block{
		Type: "section",
		Text: &text,
	}

	return Message{
		Color:  "#f2c744",
		Blocks: []Block{block},
	}
}

func ReturnSlackImage(inputString Input) Message {
	message := ReturnSlackMessage(inputString)
	message.Attachments = []Attachment{
		{
			Fallback: inputString.Content,
			ImageURL: inputString.ImageUrl,
		},
	}

	return message
}

var Instance Slack
