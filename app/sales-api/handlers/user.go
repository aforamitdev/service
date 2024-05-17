package handlers

import (
	"context"
	"net/http"
	"service2/business/auth"
	"service2/business/data/user"
	"service2/foundations/web"
)

type userGroup struct {
	user user.User
	auth *auth.Auth
}

func (ug userGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	v, ok := ctx.Value(web.KeyValue).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	users, err := ug.user.Query(ctx, v.TractID)

	if err != nil {
		return web.RespondError(ctx, w, err)
	}
	return web.Response(ctx, w, users, http.StatusOK)

}
