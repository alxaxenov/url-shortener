package utils

import (
	"context"
	"fmt"
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
