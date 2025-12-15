package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"thanhldt060802/internal/lib/otel"

	"github.com/cardinalby/hureg"
	"github.com/danielgtaylor/huma/v2"
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
		otel.InfoLog(ctx.Context(), "========> standard-auth middelware request")
		isAuthorizationRequired := false
		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if _, ok = opScheme["standard-auth"]; ok {
				otel.InfoLog(ctx.Context(), "========> standard-auth middelware validate")
				isAuthorizationRequired = true
				break
			}
		}
		otel.InfoLog(ctx.Context(), "========> require authorization: %v", isAuthorizationRequired)
		if isAuthorizationRequired {
			HumaAuthMiddleware(api, ctx, next)
		} else {
			next(ctx)
		}
	}
}

func HumaAuthMiddleware(api hureg.APIGen, ctx huma.Context, next func(huma.Context)) {
	tmpCtx, span := otel.NewHybridSpan(ctx.Context(), "HumaAuthMiddleware")
	defer span.Done()

	authHeaderValue := ctx.Header("Authorization")
	span.SetAttribute("header.authorization", authHeaderValue)

	if len(authHeaderValue) < 1 {
		otel.ErrorLog(ctx.Context(), "========> invalid credentials")
		err := errors.New("missing token")
		span.SetError(err)
		huma.WriteErr(api.GetHumaAPI(), ctx, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}

	ctx = huma.WithContext(ctx, tmpCtx)
	ctx = huma.WithValue(ctx, "auth_header", authHeaderValue)
	ctx = huma.WithValue(ctx, "token", strings.Replace(authHeaderValue, "Bearer ", "", 1))

	if err := AuthMdw.AuthMiddleware(ctx.Context()); err != nil {
		span.SetError(err)
		huma.WriteErr(api.GetHumaAPI(), ctx, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	otel.InfoLog(ctx.Context(), "========> authorize success")

	next(ctx)
}
