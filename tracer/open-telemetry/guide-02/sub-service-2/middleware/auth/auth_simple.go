package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type SimpleAuthMiddleware struct {
}

func NewSimpleAuthMiddleware() IAuthMiddleware {
	return &SimpleAuthMiddleware{}
}

func (mdw *SimpleAuthMiddleware) AuthMiddleware(ctx context.Context) (string, error) {
	token, _ := ctx.Value("auth-header").(string)
	if len(strings.Split(token, ".")) != 3 {
		err := errors.New("invalid token")
		return "", err
	}

	result := fmt.Sprintf("Token %v is accepted", token)

	return result, nil
}
