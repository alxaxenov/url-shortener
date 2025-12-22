package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
)

type ShortenerHandler struct {
	Service *service.ShortenerServiceSt
}

func (h *ShortenerHandler) AddValue(w http.ResponseWriter, r *http.Request) {
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
	short, err := h.Service.AddURL(string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, short)
}

func (h *ShortenerHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	u, err := h.Service.GetURL(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("content-type", "text/plain")
	w.Header().Set("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
