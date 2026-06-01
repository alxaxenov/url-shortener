package middleware

import (
	"context"
	"net/http"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/utils"
)

type (
	// IUserRepo интерфейс слоя репозитория.
	IUserRepo interface {
		CreateUser(ctx context.Context) (int, error)
	}

	ITokenManager interface {
		BuildTokenString(id int) (string, error)
		GetUserID(tokenString string) (int, error)
	}
)

// UserMiddleware структура middleware для аутентификации пользователя.
type UserMiddleware struct {
	CookieAuthKey string
	UserRepo      IUserRepo
	TokenManager  ITokenManager
}

// NewUserMiddleware конструктор UserMiddleware.
func NewUserMiddleware(userRepo IUserRepo, tokenManager ITokenManager) *UserMiddleware {
	return &UserMiddleware{CookieAuthKey: config.CookieAuthKey, UserRepo: userRepo, TokenManager: tokenManager}
}

// Use основная логика middleware.
// Если в запросе пользователя не содержится куки или она недействительна, выдаем новую.
// Если кука есть, но не содержит информацию о пользователе - отдаем 401.
func (m *UserMiddleware) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.CookieAuthKey)
		var userID int

		if err == nil {
			userID, err = m.TokenManager.GetUserID(cookie.Value)
			if err == nil && userID == 0 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}

		// Если кука недействительна или отсутствует, выдаем новую
		if userID < 1 {
			newID, err := m.UserRepo.CreateUser(r.Context())
			if err != nil {
				logger.Logger.Errorf("Failed to generate new userID: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			token, err := m.TokenManager.BuildTokenString(newID)
			if err != nil {
				logger.Logger.Errorf("Failed to build token string: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     m.CookieAuthKey,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
			userID = newID
		}

		ctx := utils.SetUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
