package middleware

import (
	"context"
	"errors"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
)

func GetUserID(ctx context.Context) (int, error) {
	id, ok := ctx.Value(userIDKey).(int)
	if !ok {
		return 0, errors.New("user id not found in context")
	}
	return id, nil
}
