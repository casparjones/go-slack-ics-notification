package web

import (
	"github.com/gin-gonic/gin"
	"go-slack-ics/clipdrop"
	"go-slack-ics/gpt"
	"go-slack-ics/leonardo"
	"go-slack-ics/mock"
	"go-slack-ics/slack"
	"go-slack-ics/system"
	"io"
	"log"
	"net/url"
	"os"
)

type App struct{}

func (App) ServeHTTP() {
	r := gin.Default()

	templatePath := "templates/*" // Setze den Template-Pfad
	r.LoadHTMLGlob(templatePath)

	eventManager := system.NewEventManager()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World!")
	})

	r.GET("/go", func(c *gin.Context) {
		c.String(200, "Du bist im Go-Pfad!")
	})

	r.POST("/gpt-conversations", func(c *gin.Context) {
		response := gpt.GetConversations()
		c.JSON(200, response)
	})

	r.POST("/text-to-image", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		values, err := url.ParseQuery(string(body))

		var event slack.Command
		event.ChannelID = values.Get("channel_id")
		event.Text = values.Get("text")
		event.Command = values.Get("command")

		if event.Text == "" {
			c.JSON(200, gin.H{
				"response_type": "in_channel",
				"error":         "text is empty",
			})
			return
		}

		go func() {
			bodyBytes, filename, err := leonardo.NewTextToImage(event.Text)
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			err = slack.Instance.SendImageToSlack(bodyBytes, filename, event.Text, event.ChannelID)
			if err != nil {
				log.Fatal(err.Error())
			}
		}()

		c.JSON(200, gin.H{
			"response_type": "in_channel",
		})
	})

	r.POST("/clipdrop/tti", func(c *gin.Context) {
		tti := clipdrop.NewTextToImage()
		var event slack.Command
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		values, err := url.ParseQuery(string(body))

		event.ChannelID = values.Get("channel_id")
		event.Text = values.Get("text")
		event.Command = values.Get("command")

		if event.Text == "" {
			c.JSON(200, gin.H{
				"response_type": "in_channel",
				"error":         "text is empty",
			})
			return
		}

		go func() {
			_, err := tti.Prompt(event)
			if err != nil {
				log.Fatal(err.Error())
			}
		}()

		c.JSON(200, gin.H{
			"response_type": "in_channel",
		})
	})

	r.POST("/gpt-event", func(c *gin.Context) {
		chat := gpt.NewChat(eventManager)
		var payload slack.Payload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		chat.SendAsync(payload.Event)
		response := slack.Response{
			Ok: true,
		}
		c.JSON(200, response)
	})

	r.POST("/gpt-cancel", func(c *gin.Context) {
		var event slack.Event
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		go func() {
			if eventManager.HasChannel(event.Channel) {
				eventChannel := eventManager.GetChannel(event.Channel)
				eventChannel <- system.EventMessage{
					ChannelID: event.Channel,
					Text:      "cancel",
				}
				eventManager.DeleteChannel(event.Channel)
			}
		}()

		c.JSON(200, gin.H{"success": eventManager.HasChannel(event.Channel)})
	})

	r.POST("/gpt", func(c *gin.Context) {
		chat := gpt.NewChat(eventManager)
		var event slack.Event
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		chat.SendAsync(event)
		response := slack.Response{
			Ok: true,
		}
		c.JSON(200, response)
	})

	// mockingService
	shopify := mock.NewShopify()
	shopify.Routes(r)

	r.NoRoute(func(c *gin.Context) {
		c.String(404, "Nicht gefunden")
	})

	// Lade die SSL-Zertifikate
	var err error
	if os.Getenv("GIN_SSL") == "true" {
		err = r.RunTLS(":8080", "server.pem", "server.key")
	} else {
		err = r.Run(":8080")
	}

	if err != nil {
		return
	}
}

func Start() {
	app := App{}
	app.ServeHTTP()
}
