package audit

import (
	"encoding/json"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/client/audit"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
)

// IremoteClient интерфейс клиента для взаимодействия с сервисом аудита.
type IremoteClient interface {
	Notify([]byte) error
}

// remoteObserver структура обработчика, отправляет запросы айдита в удаленный сервис.
type remoteObserver struct {
	client IremoteClient
}

// newRemoteObserver конструктор remoteObserver.
func newRemoteObserver(URL string) *remoteObserver {
	return &remoteObserver{audit.NewRemoteAuditClient(URL)}
}

// notify логика отправки запроса удаленному сервису.
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
