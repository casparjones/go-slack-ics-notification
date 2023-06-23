package slack

type BotIcon struct {
	Image36 string `json:"image_36"`
	Image48 string `json:"image_48"`
	Image72 string `json:"image_72"`
}

type BotProfile struct {
	ID      string  `json:"id"`
	AppID   string  `json:"app_id"`
	Name    string  `json:"name"`
	Icons   BotIcon `json:"icons"`
	Deleted bool    `json:"deleted"`
	Updated int     `json:"updated"`
	TeamID  string  `json:"team_id"`
}

type ResponseMetadata struct {
	Warnings []string `json:"warnings"`
}

type Response struct {
	Ok               bool             `json:"ok"`
	Channel          string           `json:"channel"`
	Ts               string           `json:"ts"`
	Message          Message          `json:"message"`
	Warning          string           `json:"warning"`
	ResponseMetadata ResponseMetadata `json:"response_metadata"`
}
