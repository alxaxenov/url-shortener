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
	"github.com/go-chi/chi/v5"
)

// IShortenerService интерфейс слоя сервиса.
//
//go:generate mockery --name Reader --srcpkg io --with-expecter=true --output ./mocks --outpkg mocks --filename mock_io_reader.go
//go:generate mockery --name IShortenerService --with-expecter=true --filename mock_shortener_service.go
type IShortenerService interface {
	AddURL(context.Context, string, int) (string, error)
	GetURL(context.Context, string) (string, error)
	SaveBatch(context.Context, model.LoadBatchRequest, int) ([]model.BatchResponse, error)
	UserURLs(context.Context, int) ([]model.UserURLs, error)
	AppendDelete(int, model.DeleteURLs)
}

// ISemaphore интерфейс реализации семафора.
//
//go:generate mockery --name ISemaphore --with-expecter=true --filename mock_semathor.go
type ISemaphore interface {
	Acquire()
	Release()
}

// AuditPublisher интерфейс интерфейс для взаимодействия с аудитом запросов.
//
//go:generate mockery --name AuditPublisher --with-expecter=true --filename mock_audit_publisher.go
type AuditPublisher interface {
	Publish(action audit.ActionType, userID int, URL string)
}

// ShortenerHandler общая структура хендлера сервиса, методы - отдельные ручки.
type ShortenerHandler struct {
	Service         IShortenerService
	DB              db.DBTX
	deleteSemaphore ISemaphore
	audit           AuditPublisher
}

// NewShortenerHandler конструктор ShortenerHandler.
func NewShortenerHandler(s IShortenerService, d db.DBTX, audit AuditPublisher) *ShortenerHandler {
	semaphore := utils.NewSemaphore(5)
	return &ShortenerHandler{Service: s, DB: d, deleteSemaphore: semaphore, audit: audit}
}

// AddValue ручка генерации короткого URL принимает и отдает text/plain.
//
// @Summary Генерация короткого URL
// @Accept  text/plain
// @Produce text/plain
// @Param original_url body string true "Оригинальный URL"
// @Success 201 {string} string http://localhost:8080/5ufzFM0w
// @Failure 400 {string} string "Внутренняя ошибка"
// @Failure 409 {string} string http://localhost:8080/5ufzFM0w
// @Failure 500 {string} string "Internal Server Error"
// @Router / [post]
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
		return
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

// GetValue ручка получения ранее сохраненного URL.
//
// @Summary Получение оригинального URL
// @Param short_url path string true "короткий URL"
// @Success 307
// @Failure 410 {string} string "Gone"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{short_url} [get]
func (h *ShortenerHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	u, err := h.Service.GetURL(r.Context(), chi.URLParam(r, "id"))
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

// AddValueJSON ручка генерации короткого URL принимает и отдает application/json.
//
// @Summary Генерация короткого URL
// @Accept  json
// @Produce json
// @Param original_url body model.AddURLRequest true "Оригинальный URL"
// @Success 201 {object} model.AddURLResponse
// @Failure 400 {string} string "Внутренняя ошибка"
// @Failure 409 {object} model.AddURLResponse
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/shorten [post]
func (h *ShortenerHandler) AddValueJSON(w http.ResponseWriter, r *http.Request) {
	req := model.AddURLRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

// Ping техническая ручка проверки работоспособности сервиса, проверяет жива ли база данных.
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

// SaveBatch ручка генерации короткого URL батчем.
//
// @Summary Генерация короткого URL батчем
// @Accept  json
// @Produce json
// @Param original_urls body model.LoadBatchRequest true "Оригинальные URL"
// @Success 201 {object} model.LoadBatchResponse
// @Failure 400 {string} string "Внутренняя ошибка"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/shorten/batch [post]
func (h *ShortenerHandler) SaveBatch(w http.ResponseWriter, r *http.Request) {
	req := model.LoadBatchRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

// UserURLs ручка получения активных URL текущего пользователя.
//
// @Summary Получение загруженных пользователем URL
// @Security     CookieAuth
// @Produce json
// @Success 200 {object} model.UserURLsResponse
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Внутренняя ошибка"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/user/urls [get]
func (h *ShortenerHandler) UserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
		respData = []byte(http.StatusText(http.StatusNoContent) + "\n")
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

// DeleteURLs архивация URL текущего пользователя.
// Обработка запроса происходит асинхронно через очередь, пользователю сразу возвращается 202 статус.
//
// @Summary Удаление коротких URL пользователя
// @Security     CookieAuth
// @Accept  json
// @Produce text/plain
// @Param urls_to_delete body model.DeleteURLs true "Короткие URL для удаления"
// @Success 202 {string} model.UserURLsResponse
// @Success 204 {string} string "Accepted"
// @Failure 400 {string} string "Внутренняя ошибка"
// @Router /api/user/urls [delete]
func (h *ShortenerHandler) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
