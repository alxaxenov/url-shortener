package model

type (
	AddURLRequest struct {
		URL string `json:"url"`
	}

	AddURLResponse struct {
		Result string `json:"result"`
	}
)
