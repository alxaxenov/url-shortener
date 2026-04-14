package audit

import (
	"encoding/json"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/client/audit"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
)

type remoteClient interface {
	Notify([]byte) error
}

type remoteObserver struct {
	client remoteClient
}

func newRemoteObserver(URL string) *remoteObserver {
	return &remoteObserver{audit.NewRemoteAuditClient(URL)}
}

func (f *remoteObserver) notify(message Message) {
	bytes, err := json.Marshal(message)
	if err != nil {
		logger.Logger.Errorf("remoteObserver failed to marshal JSON: %s", err)
	}
	err = f.client.Notify(bytes)
	if err != nil {
		logger.Logger.Errorf("remoteObserver failed to notify remote: %s", err)
	}
}
