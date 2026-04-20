package model

type (
	UserURLs struct {
		Short  string `json:"short_url"`
		Origin string `json:"original_url"`
	}

	UserURLsResponse []UserURLs
)
