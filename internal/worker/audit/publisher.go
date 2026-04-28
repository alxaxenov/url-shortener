package audit

import (
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
)

// observer интерфейс обработчика.
type observer interface {
	notify(message Message)
	close() error
}

// Publisher структура паблишера для реализации паттерна наблюдатель.
type Publisher struct {
	observers []observer
}

// NewPublisher конструктор Publisher.
func NewPublisher(filePath, URLPath string) (*Publisher, error) {
	p := &Publisher{}
	if filePath != "" {
		fileObs, err := newFileObserver(filePath)
		if err != nil {
			return nil, err
		}
		p.registerObserver(fileObs)
	}
	if URLPath != "" {
		p.registerObserver(newRemoteObserver(URLPath))
	}
	return p, nil
}

// registerObserver регистрация нового обработчика.
func (p *Publisher) registerObserver(obs observer) {
	p.observers = append(p.observers, obs)
}

// Publish отправка уведомлений зарегистрированным обработчикам.
func (p *Publisher) Publish(action ActionType, userID int, URL string) {
	msg := Message{
		Ts:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    URL,
	}
	for _, obs := range p.observers {
		go obs.notify(msg)
	}
}

// Close закрытие ресурсов
func (p *Publisher) Close() {
	for i, obs := range p.observers {
		if err := obs.close(); err != nil {
			logger.Logger.Errorf("AuditPublisher.Close observer idx %d close error: %v", i, err)
		}
	}
}
