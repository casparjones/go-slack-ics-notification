package web

import (
	"go-slack-ics/clipdrop"
	"go-slack-ics/gpt"
	"go-slack-ics/slack"
	"log"

	"github.com/gin-gonic/gin"
)

type App struct{}

func (App) ServeHTTP() {
	r := gin.Default()

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
		tti := clipdrop.NewTextToImage()
		var event slack.Event
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		go func() {
			_, err := tti.Prompt(event)
			if err != nil {
				log.Fatal(err.Error())
			}
		}()

		c.JSON(200, "ok")
	})

	r.POST("/gpt", func(c *gin.Context) {
		chat := gpt.NewChat()
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

	r.NoRoute(func(c *gin.Context) {
		c.String(404, "Nicht gefunden")
	})

	r.Run(":8080")
}

func Start() {
	app := App{}
	app.ServeHTTP()
}
