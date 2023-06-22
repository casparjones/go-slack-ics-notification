package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go-slack-ics/calendar"
	"go-slack-ics/slack"
	slackUser "go-slack-ics/slack/user"
	"go-slack-ics/web"
	"log"
	"time"
)

func startTwelveHourlyTicker() {
	// Aktuelle Uhrzeit
	now := time.Now()

	// Nächste 12-Stunden-Marke berechnen
	nextTwelve := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	hour := now.Hour()
	if hour >= 12 {
		nextTwelve = nextTwelve.Add(24 * time.Hour)
	} else if hour < 12 {
		nextTwelve = nextTwelve.Add(12 * time.Hour)
	}

	dur := nextTwelve.Sub(time.Now())

	// Erstelle einen Ticker, der nach der berechneten Dauer und dann alle 12 Stunden tickt
	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	// Nach der ersten Ausführung, setze den Ticker auf ein 12-stündiges Intervall zurück
	go func() {
		time.Sleep(dur)
		ticker.Reset(12 * time.Hour)
	}()

	fmt.Println("Warte jetzt bis: ", nextTwelve.Format("02.01.2006 15:04"))
	for {
		<-ticker.C

		if now.Hour() >= 12 {
			slack.Instance = slack.Slack{}
			slack.Instance.User = slackUser.Users["Frank"]
			calendar.Run()

			nextTwelve = nextTwelve.Add(24 * time.Hour)
		} else if now.Hour() < 12 {
			slack.Instance = slack.Slack{}
			slack.Instance.User = slackUser.Users["Wolf"]
			calendar.Run()

			nextTwelve = nextTwelve.Add(12 * time.Hour)
		}

		fmt.Println("Es sind 12 Stunden vergangen: ", time.Now().Format("15:04"))
	}
}

func main() {
	err := godotenv.Load(".env")
	slackUser.InitUsers()

	if err != nil {
		log.Printf("Error loading .env file")
	}

	go func() {
		fmt.Printf("Start Slack Notification for Users: %s \n", slackUser.Users)
		startTwelveHourlyTicker()
	}()

	web.Start()
}
