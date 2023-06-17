package calendar

import (
	"fmt"
	"github.com/apognu/gocal"
	"go-slack-ics/slack"
	"log"
	"os"
	"time"
)

type Calendar struct {
	events []gocal.Event
	start  time.Time
	end    time.Time
}

func (c *Calendar) GetStartDateForYear(year int) (time.Time, time.Time) {
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Fatal("Fehler beim Laden der Zeitzone:", err)
	}
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, location)
	end := time.Date(year, time.December, 31, 23, 59, 59, 59, location)

	return start, end
}

func (c *Calendar) GetStartDateForDate(datetime time.Time) (time.Time, time.Time) {
	midnight := time.Date(datetime.Year(), datetime.Month(), datetime.Day(), 0, 0, 0, 0, datetime.Location())
	return midnight, midnight.Add(2 * 24 * time.Hour)
}

func (c *Calendar) GetStartDateForToday() (time.Time, time.Time) {
	datetime := time.Now()
	midnight := time.Date(datetime.Year(), datetime.Month(), datetime.Day(), 0, 0, 0, 0, datetime.Location())
	return midnight, midnight.Add(24 * time.Hour)
}

func (c *Calendar) Init() {
	f, err := os.Open("./calendar/awb-abfuhrtermine.ics")
	if err != nil {
		fmt.Println("Fehler beim Öffnen der Datei:", err)
		return
	}
	defer f.Close()

	cal := gocal.NewParser(f)
	cal.Start, cal.End = &c.start, &c.end
	cal.Parse()
	c.events = cal.Events
}

func (c *Calendar) Notify() {
	for _, e := range c.events {
		// fmt.Printf("%s on %s\r\n", e.Summary, e.Start)
		slack.Instance.SendCalenderEvent(e)
	}
}

func Run() {
	c := Calendar{}

	c.start, c.end = c.GetStartDateForDate(time.Now())
	c.Init()
	c.Notify()
}
