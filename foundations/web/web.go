package web

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
)

type ctxkey int

const KeyValue ctxkey = 1

type Values struct {
	TractID    string
	Now        time.Time
	StatusCode int
}

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {

	app := App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}

	return &app
}

func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *App) Handle(method string, path string, handler Handler, mw ...Middleware) {

	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {

		v := Values{
			TractID: uuid.New().String(),
			Now:     time.Now(),
		}

		ctx := context.WithValue(r.Context(), KeyValue, &v)

		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}
	}

	a.ContextMux.Handle(method, path, h)

}
