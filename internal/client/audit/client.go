// Package audit содержит клиент к сервису аудита.
package audit

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Client структура клиента сервиса аудита.
type Client struct {
	URL    string
	client *resty.Client
}

// NewRemoteAuditClient конструктор Client.
func NewRemoteAuditClient(URL string) *Client {
	return &Client{
		URL:    URL,
		client: resty.New().SetHeader("Content-Type", "application/json"),
	}
}

// Notify отправка запроса уведомление сервису аудита.
func (rc *Client) Notify(payload []byte) error {
	resp, err := rc.client.R().SetBody(payload).Post(rc.URL)
	if err != nil {
		return fmt.Errorf("RemoteAuditClient Notify client error: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("RemoteAuditClient Notify response status code: %d", resp.StatusCode())
	}
	return nil
}
