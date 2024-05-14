package handlers

import (
	"context"
	"log"
	"net/http"
	"service2/foundations/database"
	"service2/foundations/web"

	"github.com/jmoiron/sqlx"
)

type check struct {
	log *log.Logger
	db  *sqlx.DB
}

func (c check) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	status := "OK"
	statusCode := http.StatusOK

	if err := database.StatusCheck(ctx, c.db); err != nil {
		log.Println(err)
		status = "not ready"
		statusCode = http.StatusInternalServerError
	}

	health := struct {
		Status string `json:"string"`
	}{
		Status: status,
	}
	return web.Response(ctx, w, health, statusCode)
}
