package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/service/shortener_service"
)

func AddValue(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != "text/plain" {
		http.Error(w, "Unexpected content-type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := url.ParseRequestURI(string(b)); err != nil {
		http.Error(w, "Invalid body URL", http.StatusBadRequest)
		return
	}
	short, err := shortener_service.ShortenerService.AddURL(string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, short)
}

func GetValue(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != "text/plain" {
		http.Error(w, "Unexpected content-type", http.StatusBadRequest)
		return
	}
	u, err := shortener_service.ShortenerService.GetURL(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
	io.WriteString(w, u)
}
