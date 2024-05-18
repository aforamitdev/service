package handlers

import (
	"log"
	"net/http"
	"os"
	"service2/business/auth"
	"service2/business/data/user"
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

	ug := userGroup{
		user: user.New(log, db),
		auth: auth,
	}

	app.Handle(http.MethodGet, "/v1/users", ug.query)
	app.Handle(http.MethodGet, "/v1/users/:id", ug.queryById, mid.Authenticate(auth))
	app.Handle(http.MethodPost, "/v1/users/", ug.create)
	return app
}
