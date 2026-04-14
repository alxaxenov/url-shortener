package audit

type ActionType string

const (
	Shorten ActionType = "shorten"
	Follow  ActionType = "follow"
)

type Message struct {
	Ts     int64      `json:"ts"`
	Action ActionType `json:"action"`
	UserId int        `json:"user_id,omitempty"`
	URL    string     `json:"url"`
}
