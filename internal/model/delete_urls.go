package model

type DeleteURLs []string

type DeleteRequest struct {
	UserID int
	URLs   DeleteURLs
}
