package handlers

import (
	"log"
	"net/http"
	"os"
	"service2/business/mid"
	"service2/foundations/web"
)

func API(build string, shutdown chan os.Signal, log *log.Logger) *web.App {

	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	check := check{
		log: log,
	}

	app.Handle(http.MethodGet, "/", check.readiness)

	return app
}
