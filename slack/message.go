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

type Payload struct {
	Token               string          `json:"token"`
	TeamID              string          `json:"team_id"`
	ContextTeamID       string          `json:"context_team_id"`
	ContextEnterpriseID *string         `json:"context_enterprise_id"`
	ApiAppID            string          `json:"api_app_id"`
	Event               Event           `json:"event"`
	Type                string          `json:"type"`
	EventID             string          `json:"event_id"`
	EventTime           int             `json:"event_time"`
	Authorizations      []Authorization `json:"authorizations"`
	IsExtSharedChannel  bool            `json:"is_ext_shared_channel"`
	EventContext        string          `json:"event_context"`
}

type MessageDetail struct {
	BotID      string     `json:"bot_id"`
	Type       string     `json:"type"`
	Text       string     `json:"text"`
	User       string     `json:"user"`
	AppID      string     `json:"app_id"`
	Blocks     []Block    `json:"blocks"`
	Team       string     `json:"team"`
	BotProfile BotProfile `json:"bot_profile"`
	Edited     Edited     `json:"edited"`
	Ts         string     `json:"ts"`
	SourceTeam string     `json:"source_team"`
	UserTeam   string     `json:"user_team"`
}

type Icons struct {
	Image36 string `json:"image_36"`
	Image48 string `json:"image_48"`
	Image72 string `json:"image_72"`
}

type Edited struct {
	User string `json:"user"`
	Ts   string `json:"ts"`
}

type Authorization struct {
	EnterpriseID        *string `json:"enterprise_id"`
	TeamID              string  `json:"team_id"`
	UserID              string  `json:"user_id"`
	IsBot               bool    `json:"is_bot"`
	IsEnterpriseInstall bool    `json:"is_enterprise_install"`
}

type Event struct {
	Type            string        `json:"type"`
	Subtype         string        `json:"subtype"`
	Message         MessageDetail `json:"message"`
	PreviousMessage MessageDetail `json:"previous_message"`
	Channel         string        `json:"channel"`
	Hidden          bool          `json:"hidden"`
	EventTs         string        `json:"event_ts"`
	ChannelType     string        `json:"channel_type"`
	Token           string        `json:"token"`
	TeamID          string        `json:"team_id"`
	TeamDomain      string        `json:"team_domain"`
	ChannelName     string        `json:"channel_name"`
	User            string        `json:"user"`
	UserName        string        `json:"user_name"`
	Commands        string        `json:"commands"`
	Text            string        `json:"text"`
	Timestamp       string        `json:"ts,omitempty"`
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
