package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"thanhldt060802/common/observer"

	"github.com/cardinalby/hureg"
	"github.com/danielgtaylor/huma/v2"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
)

var DefaultAuthSecurity = []map[string][]string{
	{"standard-auth": {""}},
}

type IAuthMiddleware interface {
	AuthMiddleware(ctx context.Context) error
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
	tmpCtx, span := observer.StartSpanInternal(ctx.Context())
	defer span.End()

	authHeaderValue := ctx.Header("Authorization")
	span.SetAttributes(attribute.String("header.authorization", authHeaderValue))

	if len(authHeaderValue) < 1 {
		log.Error("========> invalid credentials")
		span.Err = fmt.Errorf("missing token")
		huma.WriteErr(api.GetHumaAPI(), ctx, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), span.Err)
		return
	}

	ctx = huma.WithContext(ctx, tmpCtx)
	ctx = huma.WithValue(ctx, "auth_header", authHeaderValue)
	ctx = huma.WithValue(ctx, "token", strings.Replace(authHeaderValue, "Bearer ", "", 1))

	if err := AuthMdw.AuthMiddleware(ctx.Context()); err != nil {
		span.Err = err
		huma.WriteErr(api.GetHumaAPI(), ctx, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	log.Infof("========> authorize success")

	next(ctx)
}
