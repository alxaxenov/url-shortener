package handler

import (
	"net/http"
	"strings"
)

const host = "http://localhost:8080"

func ExampleShortenerHandler_AddValue() {
	// POST /
	body := "http://long_url.com"
	res, err := http.DefaultClient.Post(host+"/", "text/plain", strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Response:
	// 201 Created
	// Content-Type: text/plain
	// http://localhost:8080/kbMRqAx81
}

func ExampleShortenerHandler_AddValueJSON() {
	// POST /api/shorten
	body := `{"url": "http://another_long_url.com"}`
	res, err := http.DefaultClient.Post(host+"/api/shorten", "application/json", strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Response:
	// 201 Created
	// Content-Type: application/json
	// {"result":"http://localhost:8080/D1bc9C7Q"}
}

func ExampleShortenerHandler_GetValue() {
	// GET /{short}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Get(host + "/D1bc9C7Q")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Response:
	// 307 Temporary Redirect
	// Headers:
	// Location: http://yes.com
}

func ExampleShortenerHandler_SaveBatch() {
	// POST /api/shorten/batch
	body := `[{"correlation_id": "correlation_1", "original_url": "http://original_1.com"}, {"correlation_id": "correlation_2", "original_url": "http://original_2.com"}]`
	res, err := http.DefaultClient.Post(host+"/api/shorten/batch", "application/json", strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Response:
	// 201 Created
	// Content-Type: application/json
	// [{"correlation_id":"correlation_1","short_url":"http://localhost:8080/0lGHpSIQ"},{"correlation_id":"correlation_2","short_url":"http://localhost:8080/kYgA8QE6"}]
}

func ExampleShortenerHandler_UserURLs() {
	// GET /user/urls
	req, err := http.NewRequest("GET", host+"/api/user/urls", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Cookie", "AUTH_SHORTEN_KEY=<token>")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Response:
	// 200 Ok
	// Content-Type: application/json
	// [{"short_url": "http://localhost:8080/0lGHpSIQ", "original_url": "http://original_1.com"}, {"short_url": "http://localhost:8080/kYgA8QE6", "original_url": "http://original_2.com"}]
}

func ExampleShortenerHandler_DeleteURLs() {
	body := `["0lGHpSIQ", "kYgA8QE6"]`
	req, err := http.NewRequest("DELETE", host+"/api/user/urls", strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Cookie", "AUTH_SHORTEN_KEY=<token>")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Response:
	// 202 Accepted
}
