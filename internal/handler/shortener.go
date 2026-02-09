package handler

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/url"
	"time"

	"encoding/json"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
)

type ShortenerService interface {
	AddURL(string) (string, error)
	GetURL(string) (string, error)
}

type ShortenerHandler struct {
	Service ShortenerService
	DB      *sql.DB
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
		logger.Logger.Info("AddValue service.AddURL error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, short)
}

func (h *ShortenerHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	u, err := h.Service.GetURL(r.PathValue("id"))
	if err != nil {
		logger.Logger.Info("GetValue service.GetURL error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *ShortenerHandler) AddValueJSON(w http.ResponseWriter, r *http.Request) {
	req := model.AddURLRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		http.Error(w, "invalid body URL", http.StatusBadRequest)
		return
	}
	short, err := h.Service.AddURL(req.URL)
	if err != nil {
		logger.Logger.Info("AddValueJSON service.AddURL error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	respData, err := json.Marshal(model.AddURLResponse{Result: short})
	if err != nil {
		logger.Logger.Info("AddValueJSON response marshal error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respData)
}

func (h *ShortenerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	if h.DB == nil {
		status = http.StatusInternalServerError
	} else {
		ctx := r.Context()
		newCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		if err := h.DB.PingContext(newCtx); err != nil {
			status = http.StatusInternalServerError
		}
	}
	w.WriteHeader(status)
}
