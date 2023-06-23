package web

import (
	"github.com/gin-gonic/gin"
	"go-slack-ics/gpt"
)

type App struct{}

func (App) ServeHTTP() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hallo Welt!")
	})

	r.GET("/go", func(c *gin.Context) {
		c.String(200, "Du bist im Go-Pfad!")
	})

	r.POST("/gpt", func(c *gin.Context) {
		chat := gpt.NewChat()
		var event gpt.Event
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		rChan := chat.SendAsync(event)
		response := <-rChan
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
