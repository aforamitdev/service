package handlers

import (
	"context"
	"fmt"
	"net/http"
	"service2/business/auth"
	"service2/business/data/user"
	"service2/foundations/web"

	"github.com/pkg/errors"
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

func (ug userGroup) queryById(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	v, ok := ctx.Value(web.KeyValue).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from  context")
	}
	fmt.Println(v.TractID)
	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	fmt.Println(claims, ok)
	if !ok {
		return errors.New("claims missing from context")
	}
	params := web.Params(r)

	usr, err := ug.user.One(ctx, claims, params["id"])

	if err != nil {
		switch err {
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrap(err, "error ")
		}
	}
	return web.Response(ctx, w, usr, http.StatusOK)

}

func (ug userGroup) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	v, ok := ctx.Value(web.KeyValue).(*web.Values)

	fmt.Println(v)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}
	var nu user.NewUser

	if err := web.Decode(r, &nu); err != nil {
		return errors.Wrapf(err, "unable to decode payload ")
	}

	usr, err := ug.user.Create(ctx, v.TractID, nu, v.Now)

	if err != nil {
		return errors.Wrapf(err, "User: %+v", usr)
	}
	return web.Response(ctx, w, usr, http.StatusCreated)

}

func (ug userGroup) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	v, ok := ctx.Value(web.KeyValue).(*web.Values)

	fmt.Println(v)
	if !ok {
		return web.NewShutdownError("web context missing from context")
	}

	params := web.Params(r)

	err := ug.user.Delete(ctx, params["id"])
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}

	}
	return web.Response(ctx, w, nil, http.StatusNoContent)
}
