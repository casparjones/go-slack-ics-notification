package slack

type Message struct {
	Channel     string       `json:"channel,omitempty"`
	TimeStamp   string       `json:"ts,omitempty"`
	User        string       `json:"user,omitempty"`
	Text        string       `json:"text,omitempty"`
	Blocks      []Block      `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Color       string       `json:"color"`
}

type Attachment struct {
	Fallback string `json:"fallback"`
	ImageURL string `json:"image_url"`
}

type Input struct {
	Content  string
	ImageUrl string
}

type Event struct {
	Token       string `json:"token"`
	TeamID      string `json:"team_id"`
	TeamDomain  string `json:"team_domain"`
	Channel     string `json:"channel"`
	ChannelName string `json:"channel_name"`
	User        string `json:"user"`
	UserName    string `json:"user_name"`
	Commands    string `json:"commands"`
	Text        string `json:"text"`
	Timestamp   string `json:"ts,omitempty"`
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
