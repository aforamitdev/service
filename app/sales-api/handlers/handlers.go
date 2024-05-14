package handlers

import (
	"log"
	"net/http"
	"os"
	"service2/business/auth"
	"service2/business/mid"
	"service2/foundations/web"

	"github.com/jmoiron/sqlx"
)

func API(build string, shutdown chan os.Signal, log *log.Logger, auth *auth.Auth, db *sqlx.DB) *web.App {

	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	check := check{
		log: log,
		db:  db,
	}

	app.Handle(http.MethodGet, "/", check.readiness)

	return app
}
