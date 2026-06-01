package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type UserIDType string

// UserIDKey ключ хранения id пользователя в контексте.
const UserIDKey UserIDType = "userID"

// GetUserID достает id пользователя из контекста.
func GetUserID(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, fmt.Errorf("user id not found in context")
	}
	return userID, nil
}

// SetUserID устанавливает id пользователя в контекст.
func SetUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

type JWTManager struct {
	SecretKey string
}

// GetUserID получение id пользователя из токена.
func (td *JWTManager) GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(td.SecretKey), nil
		})
	if err != nil {
		return 0, fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return 0, errors.New("token is not valid")
	}
	return claims.UserID, nil
}

// BuildTokenString генерация токена.
func (td *JWTManager) BuildTokenString(id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: id,
	})
	tokenString, err := token.SignedString([]byte(td.SecretKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return tokenString, nil
}
