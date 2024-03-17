package clipdrop

import (
	"bytes"
	"fmt"
	"go-slack-ics/slack"
	"hash/fnv"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TextToImage struct {
	Endpoint string
	apiKey   string
}

type TtiResponse struct {
	Text     string
	ImageUrl string
}

func (tti *TextToImage) hash(s string) string {
	h := fnv.New64a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return ""
	}
	return strconv.FormatUint(h.Sum64(), 10)
}

func (tti *TextToImage) Prompt(event slack.Command) (*TtiResponse, error) {
	// message := slack.GetSimpleMessage(event.UserID, event.ChannelID, fmt.Sprintf("imagine .o0(%s); please wait...", event.Text))
	// initMessageResponse := slack.Instance.SendMessage(event.ChannelID, event.UserID, message)
	if event.Text == "" {
		return nil, fmt.Errorf("text is empty")
	}

	// Die URL der API
	url := "https://clipdrop-api.co/text-to-image/v1"

	// Deinen API-Key hier einsetzen
	apiKey := tti.apiKey

	// Erstelle einen Buffer, um den multipart-Formularinhalt zu schreiben
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Füge das Feld 'prompt' zum Formular hinzu
	if err := multipartWriter.WriteField("prompt", event.Text); err != nil {
		panic(err)
	}

	// Wichtig: Schließe den multipart writer, um das Ende des Formulars zu signalisieren
	if err := multipartWriter.Close(); err != nil {
		panic(err)
	}

	// Erstelle die Anfrage
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		panic(err)
	}

	// Füge den API-Key und den Content-Type Header hinzu
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Führe die Anfrage aus
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)

	currentTime := time.Now()
	filename := tti.hash(event.Text+"-"+currentTime.Format("2006-01-02 15:04:05")) + ".png"
	// objectServer := s3.NewClient("clipdrop")
	// s3Url := objectServer.Save(filename, bodyBytes)
	result := TtiResponse{event.Text, ""}

	err = slack.Instance.SendImageToSlack(bodyBytes, filename, event.Text, event.ChannelID)
	if err != nil {
		return nil, err
	}
	// slack.Instance.ChangeMessage(initMessageResponse.Ts, event.ChannelID, event.UserID, slackMessage)

	return &result, nil
}

func (tti *TextToImage) returnSlackMessage(result TtiResponse) slack.Message {
	input := slack.Input{Content: result.Text, ImageUrl: result.ImageUrl}
	return slack.ReturnSlackImage(input)
}

func NewTextToImage() (tti TextToImage) {
	tti.Endpoint = "https://clipdrop-api.co/text-to-image/v1"
	tti.apiKey = os.Getenv("CLIPDROP_API_KEY")
	return tti
}
