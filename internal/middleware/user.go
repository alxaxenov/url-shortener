package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/utils"
	"github.com/golang-jwt/jwt/v4"
)

type (
	Claims struct {
		jwt.RegisteredClaims
		UserID int `json:"user_id"`
	}

	// IUserRepo интерфейс слоя репозитория.
	IUserRepo interface {
		CreateUser(ctx context.Context) (int, error)
	}
)

// UserMiddleware структура middleware для аутентификации пользователя.
type UserMiddleware struct {
	secretKey     string
	CookieAuthKey string
	UserRepo      IUserRepo
}

// NewUserMiddleware конструктор UserMiddleware.
func NewUserMiddleware(secretKey string, userRepo IUserRepo) *UserMiddleware {
	return &UserMiddleware{secretKey: secretKey, CookieAuthKey: config.CookieAuthKey, UserRepo: userRepo}
}

// Use основная логика middleware.
// Если в запросе пользователя не содержится куки или она недействительна, выдаем новую.
// Если кука есть, но не содержит информацию о пользователе - отдаем 401.
func (m *UserMiddleware) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.CookieAuthKey)
		var userID int

		if err == nil {
			userID, err = m.getUserID(cookie.Value)
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
			token, err := m.buildTokenString(newID)
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

// getUserID получение id пользователя из токена.
func (m *UserMiddleware) getUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(m.secretKey), nil
		})
	if err != nil {
		return 0, fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return 0, errors.New("token is not valid")
	}
	return claims.UserID, nil
}

// buildTokenString генерация токена.
func (m *UserMiddleware) buildTokenString(id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: id,
	})
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return tokenString, nil
}
