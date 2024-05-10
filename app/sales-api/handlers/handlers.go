package handlers

import (
	"log"
	"net/http"
	"os"
	"service2/foundations/web"
)

func API(build string, shutdown chan os.Signal, log *log.Logger) *web.App {

	app := web.NewApp(shutdown)
	check := check{
		log: log,
	}

	app.Handle(http.MethodGet, "/", check.readiness)

	return app
}
