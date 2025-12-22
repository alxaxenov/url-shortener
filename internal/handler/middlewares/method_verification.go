package middlewares

import (
	"net/http"
	"slices"
)

func MethodVerification(methods ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !slices.Contains(methods, r.Method) {
				http.Error(w, "Method not allowed", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
