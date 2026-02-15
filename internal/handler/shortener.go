package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"encoding/json"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config/db"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
)

//go:generate mockery --name ShortenerService --with-expecter=true
type ShortenerService interface {
	AddURL(context.Context, string) (string, error)
	GetURL(context.Context, string) (string, error)
	LoadBatch(context.Context, model.LoadBatchRequest) ([]model.BatchResponse, error)
}

type ShortenerHandler struct {
	Service ShortenerService
	DB      db.DBTX
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
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	responseStatus := http.StatusCreated
	short, err := h.Service.AddURL(ctx, string(b))
	if err != nil {
		logger.Logger.Info("AddValue service.AddURL error:", err)
		var badURL *service.BadURL
		var alreadyExists *service.AlreadyExists
		if errors.As(err, &alreadyExists) {
			responseStatus = http.StatusConflict
		} else if errors.As(err, &badURL) {
			http.Error(w, badURL.Error(), http.StatusBadRequest)
			return
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(responseStatus)
	io.WriteString(w, short)
}

func (h *ShortenerHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	u, err := h.Service.GetURL(ctx, r.PathValue("id"))
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
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	responseStatus := http.StatusCreated
	short, err := h.Service.AddURL(ctx, req.URL)
	if err != nil {
		logger.Logger.Info("AddValueJSON service.AddURL error:", err)
		var badURL *service.BadURL
		var alreadyExists *service.AlreadyExists
		if errors.As(err, &alreadyExists) {
			responseStatus = http.StatusConflict
		} else if errors.As(err, &badURL) {
			http.Error(w, badURL.Error(), http.StatusBadRequest)
			return
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	respData, err := json.Marshal(model.AddURLResponse{Result: short})
	if err != nil {
		logger.Logger.Info("AddValueJSON response marshal error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(responseStatus)
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

func (h *ShortenerHandler) LoadBatch(w http.ResponseWriter, r *http.Request) {
	req := model.LoadBatchRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	data, err := h.Service.LoadBatch(newCtx, req)
	if err != nil {
		logger.Logger.Info("LoadBatch service.LoadBatch error:", err)
		var badURL *service.BadURL
		if errors.As(err, &badURL) {
			http.Error(w, badURL.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	respData, err := json.Marshal(model.LoadBatchResponse(data))
	if err != nil {
		logger.Logger.Info("LoadBatch response marshal error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respData)
}
