package audit

type (
	ActionType string

	// Message структура для уведомления обработчиков.
	Message struct {
		Ts     int64      `json:"ts"`
		Action ActionType `json:"action"`
		UserId int        `json:"user_id,omitempty"`
		URL    string     `json:"url"`
	}
)

// Тип произошедшего события.
const (
	// Shorten генерация нового URL.
	Shorten ActionType = "shorten"
	// Follow запрос существующего сохраненного URL.
	Follow ActionType = "follow"
)
