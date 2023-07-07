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

type Command struct {
	Token               string `url:"token"`
	TeamID              string `url:"team_id"`
	TeamDomain          string `url:"team_domain"`
	ChannelID           string `url:"channel_id"`
	ChannelName         string `url:"channel_name"`
	UserID              string `url:"user_id"`
	UserName            string `url:"user_name"`
	Command             string `url:"command"`
	Text                string `url:"text"`
	APIAppID            string `url:"api_app_id"`
	IsEnterpriseInstall bool   `url:"is_enterprise_install"`
	ResponseURL         string `url:"response_url"`
	TriggerID           string `url:"trigger_id"`
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
