package mid

import (
	"context"
	"net/http"
	"service2/business/auth"
	"service2/foundations/web"
	"strings"

	"github.com/pkg/errors"
)

func Authenticate(a *auth.Auth) web.Middleware {
	m := func(after web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			parts := strings.Split(r.Header.Get("authorization"), " ")

			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header formate: bearer <token>")
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return web.NewRequestError(err, http.StatusUnauthorized)
			}
			ctx = context.WithValue(ctx, auth.Key, claims)

			return after(ctx, w, r)

		}
		return h
	}
	return m
}

func Authorize(roles ...string) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			claims, ok := ctx.Value(auth.Key).(auth.Claims)

			if !ok {
				return errors.New("claims missing form context: HasRole called without/before")
			}

			if !claims.HasRoles(roles...) {
				return auth.ErrForbidden
			}
			return handler(ctx, w, r)

		}
		return h
	}
	return m
}
