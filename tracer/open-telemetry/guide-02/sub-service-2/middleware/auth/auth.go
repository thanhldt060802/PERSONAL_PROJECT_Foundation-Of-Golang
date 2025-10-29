package auth

import (
	"context"
	"errors"
	"net/http"
	"thanhldt060802/common/tracer"

	"github.com/cardinalby/hureg"
	"github.com/danielgtaylor/huma/v2"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
)

var DefaultAuthSecurity = []map[string][]string{
	{"standard-auth": {""}},
}

type IAuthMiddleware interface {
	AuthMiddleware(ctx context.Context) (string, error)
}

var AuthMdw IAuthMiddleware

func NewAuthMiddleware(api hureg.APIGen) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		log.Info("========> standard-auth middelware request")
		isAuthorizationRequired := false
		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if _, ok = opScheme["standard-auth"]; ok {
				log.Info("========> standard-auth middelware validate")
				isAuthorizationRequired = true
				break
			}
		}
		log.Infof("========> require authorization: %v", isAuthorizationRequired)
		if isAuthorizationRequired {
			HumaAuthMiddleware(api, ctx, next)
		} else {
			next(ctx)
		}
	}
}

func HumaAuthMiddleware(api hureg.APIGen, ctx huma.Context, next func(huma.Context)) {
	tmpCtx, span := tracer.StartSpanInternal(ctx.Context())
	defer span.End()

	authHeaderValue := ctx.Header("Authorization")
	if len(authHeaderValue) < 1 {
		log.Error("========> invalid credentials")
		err := errors.New("missing token")
		huma.WriteErr(api.GetHumaAPI(), ctx, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		span.Err = err
		return
	}

	ctx = huma.WithContext(ctx, tmpCtx)
	ctx = huma.WithValue(ctx, "auth-header", authHeaderValue)
	span.SetAttributes(attribute.String("auth.header", authHeaderValue))

	authResult, err := AuthMdw.AuthMiddleware(ctx.Context())
	if err != nil {
		huma.WriteErr(api.GetHumaAPI(), ctx, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		span.Err = err
		return
	}
	log.Infof("========> auth result: %v", authResult)
	span.SetAttributes(attribute.String("auth.result", authResult))

	next(ctx)
}
