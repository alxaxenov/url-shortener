package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

type ShortenerService interface {
	AddURL(string) (string, error)
	GetURL(string) (string, error)
}

type ShortenerHandler struct {
	Service ShortenerService
}

func (h *ShortenerHandler) AddValue(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("content-type") != "text/plain" {
		http.Error(w, "unexpected content-type", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := url.ParseRequestURI(string(b)); err != nil {
		http.Error(w, "invalid body URL", http.StatusBadRequest)
		return
	}
	short, err := h.Service.AddURL(string(b))
	if err != nil {
		log.Println("AddValue service.AddURL error:", err)
		text := http.StatusText(http.StatusInternalServerError)
		http.Error(w, text, http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, short)
}

func (h *ShortenerHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	u, err := h.Service.GetURL(r.PathValue("id"))
	if err != nil {
		log.Println("GetValue service.GetURL error:", err)
		text := http.StatusText(http.StatusInternalServerError)
		http.Error(w, text, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
