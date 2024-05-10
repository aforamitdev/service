package main

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"os"
	"os/signal"
	"service2/app/sales-api/handlers"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"net/http"
	_ "net/http/pprof"

	"github.com/ardanlabs/conf/v3"
)

var build = "develop"

func main() {
	log := log.New(os.Stdout, "SALES : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	err := run(log)
	if err != nil {
		fmt.Printf("error running app ")
	}

}

func run(log *log.Logger) error {

	var cfg struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeOut    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
	}
	cfg.Version.Build = build
	// cfg.Version.SVN = "12"
	cfg.Version.Desc = "copyright information here"
	parse, err := conf.Parse("APP", &cfg)
	if err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usege, err := conf.UsageInfo("SALES", &cfg)
			if err != nil {
				return errors.Wrap(err, "generation config uses")
			}
			fmt.Println(usege)
			return nil

		}
		return errors.Wrap(err, "parsing config")
	}
	fmt.Print(parse)

	expvar.NewString("build").Set(build)
	log.Printf("main: started : Application initializing : version %q", build)
	defer log.Println("main: Completed")

	out, err := conf.String(&cfg)

	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: config:\n %v\n", out)

	log.Println("main: Initializing debugging support ")

	go func() {
		log.Printf("main:: debug listening %s", cfg.Web.DebugHost)
		if err := http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux); err != nil {
			log.Printf("main: debug listener closed :%v", err)
		}
	}()

	// select {}

	log.Println("main: initializing API Support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handlers.API(build, shutdown, log),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeOut,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("main: API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		log.Printf("main: %v : start shutdown", sig)
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}

	}
	return nil
}
