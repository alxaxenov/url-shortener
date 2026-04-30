package utils

// Semaphore структура локальной реализации семафора.
type Semaphore struct {
	semaCh chan any
}

// NewSemaphore конструктор Semaphore.
func NewSemaphore(maxReq int) *Semaphore {
	return &Semaphore{
		semaCh: make(chan any, maxReq),
	}
}

// Acquire занимает ресурс.
func (s *Semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

// Release освобождает ресурс.
func (s *Semaphore) Release() {
	<-s.semaCh
}
