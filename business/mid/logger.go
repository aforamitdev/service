package mid

import (
	"context"
	"log"
	"net/http"
	"service2/foundations/web"
	"time"
)

func Logger(log *log.Logger) web.Middleware {

	m := func(handlers web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			v, ok := ctx.Value(web.KeyValue).(*web.Values)

			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			log.Printf("%s : started : %s %s -> %s", v.TractID, r.Method, r.URL.Path, r.RemoteAddr)

			err := handlers(ctx, w, r)
			// why log is not loging err

			log.Printf("%s : ended   : %s %s -> %s (%d) (%s)", v.TractID, r.Method, r.URL.Path, r.RemoteAddr, v.StatusCode, time.Since(v.Now))

			return err
		}
		return h
	}
	return m
}
