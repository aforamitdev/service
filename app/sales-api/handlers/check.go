package handlers

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"service2/foundations/web"
)

type check struct {
	log *log.Logger
}

func (c check) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	if n := rand.Intn(100); n%2 == 0 {
		return web.RespondError(ctx, w, errors.New("errors"))
		// return errors.New("untrusted error")
	}

	status := struct {
		Status string `json:"status"`
	}{Status: "OK"}

	return web.Response(ctx, w, status, http.StatusOK)
}
