package audit

import "time"

type observer interface {
	notify(message Message)
}

type Publisher struct {
	observers []observer
}

func NewPublisher(filePath, URLPath string) *Publisher {
	p := &Publisher{}
	if filePath != "" {
		p.registerObserver(newFileObserver(filePath))
	}
	if URLPath != "" {
		p.registerObserver(newRemoteObserver(URLPath))
	}
	return p
}

func (p *Publisher) registerObserver(obs observer) {
	p.observers = append(p.observers, obs)
}

func (p *Publisher) Publish(action ActionType, userID int, URL string) {
	msg := Message{
		Ts:     time.Now().Unix(),
		Action: action,
		UserId: userID,
		URL:    URL,
	}
	for _, obs := range p.observers {
		obs.notify(msg)
	}
}
