package mid

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"service2/foundations/web"
)

func Errors(log *log.Logger) web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// trace id

			v, ok := ctx.Value(web.KeyValue).(*web.Values)

			if !ok {
				return web.NewShutdownError("web value missing from context")
			}
			// call the handler

			if err := handler(ctx, w, r); err != nil {

				fmt.Println(err, "Error")
				log.Printf("%s : ERROR : %v", v.TractID, err)

				// handler the error, error should handler the error , if now we have issues
				if err := web.RespondError(ctx, w, err); err != nil {
					return err
				}

				if ok := web.IsShutdown(err); ok {
					return err
				}

			}

			return nil
		}

		return h

	}

	return m
}
