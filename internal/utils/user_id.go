package utils

import (
	"context"
	"fmt"
)

const UserIDKey string = "userID"

func GetUserID(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, fmt.Errorf("user id not found in context")
	}
	return userID, nil
}

func SetUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
