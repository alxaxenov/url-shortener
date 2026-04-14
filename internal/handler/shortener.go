package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config/db"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/utils"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/worker/audit"
)

//go:generate mockery --name Reader --srcpkg io --with-expecter=true --output ./mocks --outpkg mocks --filename mock_io_reader.go
//go:generate mockery --name ShortenerService --with-expecter=true --filename mock_shortener_service.go
type ShortenerService interface {
	AddURL(context.Context, string, int) (string, error)
	GetURL(context.Context, string) (string, error)
	SaveBatch(context.Context, model.LoadBatchRequest, int) ([]model.BatchResponse, error)
	UserURLs(context.Context, int) ([]model.UserURLs, error)
	AppendDelete(int, model.DeleteURLs)
}

type SemaphoreInt interface {
	Acquire()
	Release()
}

//go:generate mockery --name AuditPublisher --with-expecter=true --filename mock_audit_publisher.go
type AuditPublisher interface {
	Publish(action audit.ActionType, userID int, URL string)
}

type ShortenerHandler struct {
	Service         ShortenerService
	DB              db.DBTX
	deleteSemaphore SemaphoreInt
	audit           AuditPublisher
}

func NewShortenerHandler(s ShortenerService, d db.DBTX, audit AuditPublisher) Handler {
	semaphore := utils.NewSemaphore(5)
	return &ShortenerHandler{Service: s, DB: d, deleteSemaphore: semaphore, audit: audit}
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
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	responseStatus := http.StatusCreated
	short, err := h.Service.AddURL(r.Context(), string(b), userID)
	if err != nil {
		var badURL *service.BadURL
		var alreadyExists *service.AlreadyExists
		if errors.As(err, &alreadyExists) {
			responseStatus = http.StatusConflict
		} else if errors.As(err, &badURL) {
			logger.Logger.Info("AddValue service.AddURL badURL", "error", err)
			http.Error(w, badURL.Error(), http.StatusBadRequest)
			return
		} else {
			logger.Logger.Error("AddValue service.AddURL", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	if responseStatus == http.StatusCreated && h.audit != nil {
		go h.audit.Publish(audit.Shorten, userID, string(b))
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(responseStatus)
	io.WriteString(w, short)
}

func (h *ShortenerHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	u, err := h.Service.GetURL(r.Context(), r.PathValue("id"))
	if err != nil {
		var status int
		if errors.Is(err, service.ErrURLDeleted) {
			status = http.StatusGone
		} else {
			logger.Logger.Error("GetValue service.GetURL", "error", err)
			status = http.StatusInternalServerError
		}

		http.Error(w, http.StatusText(status), status)
		return
	}
	if h.audit != nil {
		userId, _ := utils.GetUserID(r.Context())
		go h.audit.Publish(audit.Follow, userId, u)
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
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	responseStatus := http.StatusCreated
	short, err := h.Service.AddURL(r.Context(), req.URL, userID)
	if err != nil {
		var badURL *service.BadURL
		var alreadyExists *service.AlreadyExists
		if errors.As(err, &alreadyExists) {
			responseStatus = http.StatusConflict
		} else if errors.As(err, &badURL) {
			logger.Logger.Info("AddValueJSON service.AddValueJSON badURL", "error", err)
			http.Error(w, badURL.Error(), http.StatusBadRequest)
			return
		} else {
			logger.Logger.Error("AddValueJSON service.AddValueJSON", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	respData, err := json.Marshal(model.AddURLResponse{Result: short})
	if err != nil {
		logger.Logger.Error("Failed to marshal AddValueJSON", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if responseStatus == http.StatusCreated && h.audit != nil {
		go h.audit.Publish(audit.Shorten, userID, req.URL)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(responseStatus)
	w.Write(respData)
}

func (h *ShortenerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.DB != nil {
		if err := h.DB.PingContext(r.Context()); err != nil {
			logger.Logger.Error("Failed to ping database", "error", err)
			http.Error(w, "database unavailable", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ShortenerHandler) SaveBatch(w http.ResponseWriter, r *http.Request) {
	req := model.LoadBatchRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	data, err := h.Service.SaveBatch(r.Context(), req, userID)
	if err != nil {
		logger.Logger.Error("SaveBatch service.SaveBatch", "error", err)
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
		logger.Logger.Error("SaveBatch response marshal", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respData)
}

func (h *ShortenerHandler) UserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	data, err := h.Service.UserURLs(r.Context(), userID)
	if err != nil {
		logger.Logger.Error("UserURLs service.UserUrls", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	status := http.StatusOK
	var respData []byte
	if len(data) == 0 {
		status = http.StatusNoContent
		respData = []byte(http.StatusText(http.StatusNoContent))
	} else {
		respData, err = json.Marshal(model.UserURLsResponse(data))
		if err != nil {
			logger.Logger.Error("UserURLs response marshal", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(respData)
}

func (h *ShortenerHandler) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := model.DeleteURLs{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	go func() {
		h.deleteSemaphore.Acquire()
		h.Service.AppendDelete(userID, req)
		defer h.deleteSemaphore.Release()
	}()
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(http.StatusText(http.StatusAccepted)))
}
