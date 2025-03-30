package domain

type MattermostRequest struct {
	Command   string `json:"command"`
	Text      string `json:"text"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
}

type MattermostResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

type Result struct {
	Question  string `json:"question"`
	Option    string `json:"option"`
	Count     int    `json:"count"`
	ExpiresAt string `json:"expires_at"`
}

type Results []Result
