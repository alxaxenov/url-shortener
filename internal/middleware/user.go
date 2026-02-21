package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

type UserRepoInt interface {
	CreateUser(ctx context.Context) (int, error)
}

type UserMiddleware struct {
	secretKey     string
	CookieAuthKey string
	UserRepo      UserRepoInt
}

func (m *UserMiddleware) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.CookieAuthKey)
		var userID int

		if err == nil {
			userID, _ = m.getUserID(cookie.Value)
			if userID == 0 {
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
				Name:  m.CookieAuthKey,
				Value: token,
				//Path:     "/",
				//HttpOnly: true,
				//Secure:   true,
				//SameSite: http.SameSiteStrictMode,
			})
			userID = newID
		}

		ctx := context.WithValue(r.Context(), config.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *UserMiddleware) getUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(m.secretKey), nil
		})
	if err != nil {
		return -1, fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return -1, errors.New("token is not valid")
	}
	return claims.UserID, nil
}

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

func NewUserMiddleware(secretKey string, userRepo UserRepoInt) *UserMiddleware {
	return &UserMiddleware{secretKey: secretKey, CookieAuthKey: config.CookieAuthKey, UserRepo: userRepo}
}
