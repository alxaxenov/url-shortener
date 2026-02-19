package handler

import (
	"net/http"
	"time"
)

func timeoutHandler(handler http.HandlerFunc, timeout time.Duration, msg string) http.Handler {
	if msg == "" {
		msg = "Service unavailable: request timeout"
	}
	return http.TimeoutHandler(handler, timeout, msg)
}
