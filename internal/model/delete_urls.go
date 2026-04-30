package model

type (
	DeleteURLs []string

	DeleteRequest struct {
		UserID int
		URLs   DeleteURLs
	}
)
