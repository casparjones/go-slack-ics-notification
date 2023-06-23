package slack

type Message struct {
	Channel   string  `json:"channel,omitempty"`
	TimeStamp string  `json:"ts,omitempty"`
	User      string  `json:"user,omitempty"`
	Blocks    []Block `json:"blocks"`
	Color     string  `json:"color"`
}

type Block struct {
	Type   string `json:"type"`
	Text   *Text  `json:"text,omitempty"`
	Fields []Text `json:"fields,omitempty"`
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func GetSimpleMessage(user string, channel string, message string) Message {
	return Message{
		User:    user,
		Channel: channel,
		Blocks: []Block{
			{
				Type: "section",
				Text: &Text{
					Type: "mrkdwn",
					Text: message,
				},
			},
		},
	}
}
