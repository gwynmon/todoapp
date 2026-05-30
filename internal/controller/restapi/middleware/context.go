package middleware

import (
	"context"
	"errors"
)

func GetUserID(ctx context.Context) (int, error) {
	id, ok := ctx.Value(userIDKey).(int)
	if !ok {
		return 0, errors.New("user id not found in context")
	}
	return id, nil
}
