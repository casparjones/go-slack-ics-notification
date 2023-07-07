package clipdrop

import (
	"bytes"
	"fmt"
	"go-slack-ics/s3"
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

func (tti *TextToImage) Prompt(event slack.Event) (*TtiResponse, error) {
	message := slack.GetSimpleMessage(event.User, event.Channel, fmt.Sprintf("imagine .o0(%s); please wait...", event.Text))
	initMessageResponse := slack.Instance.SendMessage(event.Channel, event.User, message)
	event.Timestamp = initMessageResponse.Ts

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	err := w.WriteField("prompt", event.Text)
	if err != nil {
		panic(err)
	}

	w.Close()

	req, err := http.NewRequest("POST", "https://clipdrop-api.co/text-to-image/v1", &b)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("x-api-key", tti.apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)

	currentTime := time.Now()
	filename := tti.hash(event.Text+"-"+currentTime.Format("2006-01-02 15:04:05")) + ".png"
	objectServer := s3.NewClient("clipdrop")
	url := objectServer.Save(filename, bodyBytes)
	result := TtiResponse{event.Text, url}

	slackMessage := tti.returnSlackMessage(result)
	slack.Instance.ChangeMessage(event.Timestamp, event.Channel, event.User, slackMessage)

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
