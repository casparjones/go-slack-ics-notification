package leonardo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type Leonardo struct {
	Endpoint string
	apiKey   string
}

type TtiResponse struct {
	Text     string
	ImageUrl string
}

type SDGenerationJob struct {
	GenerationId  string `json:"generationId"`
	ApiCreditCost int    `json:"apiCreditCost"`
}

type Payload struct {
	SDGenerationJob SDGenerationJob `json:"sdGenerationJob"`
}

// Funktion zum Extrahieren des Dateinamens aus einer URL
func (leonardo *Leonardo) ExtractFileNameFromURL(fileURL string) (string, error) {
	// Parsen der URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	// Extrahieren des Pfads und des Dateinamens
	fileName := path.Base(parsedURL.Path)
	return fileName, nil
}

func (leonardo *Leonardo) LoadImage(url string) ([]byte, string, error) {
	// Erstelle eine neue HTTP-Anfrage
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %v", err)
	}

	// Führe die Anfrage aus
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Überprüfe den HTTP-Statuscode
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to load image, status code: %d", resp.StatusCode)
	}

	// Lese die Antwort (das Bild) in ein []byte
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %v", err)
	}

	filename, err := leonardo.ExtractFileNameFromURL(url)
	log.Println("Load Image:", filename)

	if err != nil {
		return nil, "", fmt.Errorf("failed to extract filename: %v", err)
	}

	return imageData, filename, nil
}

func (leonardo *Leonardo) GetImage(id string) ([]byte, string, error) {
	generateUrl := "https://cloud.leonardo.ai/api/rest/v1/generations/" + id

	req, _ := http.NewRequest("GET", generateUrl, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", "Bearer "+leonardo.apiKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	jsonData, err := io.ReadAll(res.Body)

	// Erstelle eine Instanz von Payload
	var responseData ImageResponse

	// Unmarshale das JSON-[]byte-Objekt in die Payload-Struktur
	err = json.Unmarshal(jsonData, &responseData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil, "", err
	}

	log.Println("Get ID:", responseData.GenerationsByPK.Status)
	if responseData.GenerationsByPK.Status != "COMPLETE" {
		time.Sleep(5 * time.Second)
		return leonardo.GetImage(id)
	}

	return leonardo.LoadImage(responseData.GenerationsByPK.GeneratedImages[0].URL)
}

type GeneratedImage struct {
	URL                             string        `json:"url"`
	NSFW                            bool          `json:"nsfw"`
	ID                              string        `json:"id"`
	LikeCount                       int           `json:"likeCount"`
	MotionMP4URL                    *string       `json:"motionMP4URL"`
	GeneratedImageVariationGenerics []interface{} `json:"generated_image_variation_generics"`
}

type GenerationsByPK struct {
	GeneratedImages []GeneratedImage `json:"generated_images"`
	Status          string           `json:"status"`
	ID              string           `json:"id"`
}

type ImageResponse struct {
	GenerationsByPK GenerationsByPK `json:"generations_by_pk"`
}

func (leonardo *Leonardo) Generate(prompt string) ([]byte, string, error) {
	// Definiere den String als Map
	data := map[string]interface{}{
		"alchemy":     true,
		"height":      1024,
		"modelId":     "b24e16ff-06e3-43eb-8d33-4416c2d75876",
		"num_images":  1,
		"presetStyle": "DYNAMIC",
		"prompt":      prompt,
		"width":       1024,
	}

	// Konvertiere die Map zu JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return nil, "", err
	}

	// Erstelle den Payload
	payload := strings.NewReader(string(jsonData))

	req, _ := http.NewRequest("POST", leonardo.Endpoint, payload)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Bearer "+leonardo.apiKey)

	res, _ := http.DefaultClient.Do(req)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(res.Body)
	jsonData, err = io.ReadAll(res.Body)

	// Erstelle eine Instanz von Payload
	var responseData Payload

	// Unmarshale das JSON-[]byte-Objekt in die Payload-Struktur
	err = json.Unmarshal(jsonData, &responseData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil, "", err
	}

	log.Println("Generation ID:", responseData.SDGenerationJob.GenerationId)

	return leonardo.GetImage(responseData.SDGenerationJob.GenerationId)
}

func NewTextToImage(prompt string) ([]byte, string, error) {
	leonardo := Leonardo{}
	leonardo.Endpoint = "https://cloud.leonardo.ai/api/rest/v1/generations"
	leonardo.apiKey = os.Getenv("LEONARDO_API_KEY")

	return leonardo.Generate(prompt)
}
