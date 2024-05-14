package main

import (
	"context"
	"crypto/rsa"
	"expvar"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"service2/app/sales-api/handlers"
	"service2/business/auth"
	"service2/foundations/database"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt"
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
		log.Printf("error running app ")
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
		Auth struct {
			KeyID          string `conf:"default:3f433e9a-1bbc-4925-98f8-f4e119cd6bce"`
			PrivateKeyFile string `conf:"default:/home/amit/go/src/github.com/service2/private.pem"`
			Algorithm      string `conf:"default:RS256"`
		}
		DB struct {
			User       string `conf:"default:admin"`
			Password   string `conf:"default:admin,noprint"`
			Host       string `conf:"default:127.0.0.0:5432"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:true"`
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

	// auth initializations

	privatePEM, err := ioutil.ReadFile(cfg.Auth.PrivateKeyFile)

	if err != nil {
		return errors.Wrap(err, "reading auth private key")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return errors.Wrap(err, "parsing auth private key")
	}

	lookup := func(kid string) (*rsa.PublicKey, error) {
		switch kid {
		case cfg.Auth.KeyID:
			return &privateKey.PublicKey, nil
		}
		return nil, fmt.Errorf("no public key found for the specific kid: %s", kid)
	}

	auth, err := auth.New(cfg.Auth.Algorithm, lookup, auth.Keys{cfg.Auth.KeyID: privateKey})

	if err != nil {
		return errors.Wrap(err, "constructing auth")
	}

	log.Println("main: Initializing debugging support ")

	go func() {
		log.Printf("main: debug listening %s", cfg.Web.DebugHost)
		if err := http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux); err != nil {
			log.Printf("main: debug listener closed :%v", err)
		}
	}()

	db, err := database.Open(database.Config{User: cfg.DB.User, Password: cfg.DB.Password, Host: cfg.DB.Host, Name: cfg.DB.Name, DisableTLS: cfg.DB.DisableTLS})

	if err != nil {
		return errors.Wrap(err, "connect to db ")
	}

	defer func() {
		log.Printf("main:database stopped : %s", cfg.DB.Host)
		db.Close()
	}()

	// select {}

	log.Println("main: initializing API Support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handlers.API(build, shutdown, log, auth, db),
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
