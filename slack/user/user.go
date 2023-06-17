package user

import "os"

type User struct {
	Slack string
}

var Users map[string]string

func InitUsers() {
	Users = make(map[string]string)
	Users["Frank"] = os.Getenv("SLACK_FRANK")
	Users["Wolf"] = os.Getenv("SLACK_WOLF")
}
