package auth

import (
	"context"
	"fmt"
)

type SimpleAuthMiddleware struct {
}

func NewSimpleAuthMiddleware() IAuthMiddleware {
	return &SimpleAuthMiddleware{}
}

func (mdw *SimpleAuthMiddleware) AuthMiddleware(ctx context.Context) error {
	token, _ := ctx.Value("token").(string)
	if token != "XXX" {
		return fmt.Errorf("invalid token")
	}

	return nil
}
