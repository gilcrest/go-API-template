// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"context"
	"database/sql"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/datastore/pingstore"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"
	"github.com/gilcrest/go-api-basic/handler"
	"github.com/google/wire"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opencensus.io/trace"
	"gocloud.dev/server"
	"gocloud.dev/server/driver"
	"gocloud.dev/server/health"
	"gocloud.dev/server/health/sqlhealth"
	"net/http"
)

// Injectors from inject_main.go:

func newServer(ctx context.Context, logger zerolog.Logger, dsn datastore.PGDatasourceName) (*server.Server, func(), error) {
	googleAccessTokenConverter := authgateway.GoogleAccessTokenConverter{}
	defaultAuthorizer := auth.DefaultAuthorizer{}
	defaultStringGenerator := random.DefaultStringGenerator{}
	db, cleanup, err := datastore.NewDB(dsn, logger)
	if err != nil {
		return nil, nil, err
	}
	defaultDatastore := datastore.NewDefaultDatastore(db)
	defaultTransactor := moviestore.NewDefaultTransactor(defaultDatastore)
	defaultSelector := moviestore.NewDefaultSelector(defaultDatastore)
	defaultMovieHandlers := handler.DefaultMovieHandlers{
		AccessTokenConverter:  googleAccessTokenConverter,
		Authorizer:            defaultAuthorizer,
		RandomStringGenerator: defaultStringGenerator,
		Transactor:            defaultTransactor,
		Selector:              defaultSelector,
	}
	createMovieHandler := handler.ProvideCreateMovieHandler(defaultMovieHandlers)
	findMovieByIDHandler := handler.ProvideFindMovieByIDHandler(defaultMovieHandlers)
	findAllMoviesHandler := handler.ProvideFindAllMoviesHandler(defaultMovieHandlers)
	updateMovieHandler := handler.ProvideUpdateMovieHandler(defaultMovieHandlers)
	deleteMovieHandler := handler.ProvideDeleteMovieHandler(defaultMovieHandlers)
	defaultPinger := pingstore.NewDefaultPinger(defaultDatastore)
	defaultPingHandler := handler.DefaultPingHandler{
		Pinger: defaultPinger,
	}
	pingHandler := handler.ProvidePingHandler(defaultPingHandler)
	handlers := handler.Handlers{
		CreateMovieHandler:   createMovieHandler,
		FindMovieByIDHandler: findMovieByIDHandler,
		FindAllMoviesHandler: findAllMoviesHandler,
		UpdateMovieHandler:   updateMovieHandler,
		DeleteMovieHandler:   deleteMovieHandler,
		PingHandler:          pingHandler,
	}
	router := handler.NewMuxRouter(logger, handlers)
	v, cleanup2 := appHealthChecks(db)
	exporter := _wireExporterValue
	sampler := trace.AlwaysSample()
	defaultDriver := server.NewDefaultDriver()
	options := &server.Options{
		HealthChecks:          v,
		TraceExporter:         exporter,
		DefaultSamplingPolicy: sampler,
		Driver:                defaultDriver,
	}
	serverServer := server.New(router, options)
	return serverServer, func() {
		cleanup2()
		cleanup()
	}, nil
}

var (
	_wireExporterValue = trace.Exporter(nil)
)

// inject_main.go:

var pingHandlerSet = wire.NewSet(pingstore.NewDefaultPinger, wire.Bind(new(pingstore.Pinger), new(pingstore.DefaultPinger)), wire.Struct(new(handler.DefaultPingHandler), "*"), handler.ProvidePingHandler)

var movieHandlerSet = wire.NewSet(wire.Struct(new(random.DefaultStringGenerator), "*"), wire.Bind(new(random.StringGenerator), new(random.DefaultStringGenerator)), wire.Struct(new(authgateway.GoogleAccessTokenConverter), "*"), wire.Bind(new(auth.AccessTokenConverter), new(authgateway.GoogleAccessTokenConverter)), wire.Struct(new(auth.DefaultAuthorizer), "*"), wire.Bind(new(auth.Authorizer), new(auth.DefaultAuthorizer)), moviestore.NewDefaultTransactor, wire.Bind(new(moviestore.Transactor), new(moviestore.DefaultTransactor)), moviestore.NewDefaultSelector, wire.Bind(new(moviestore.Selector), new(moviestore.DefaultSelector)), wire.Struct(new(handler.DefaultMovieHandlers), "*"), handler.ProvideCreateMovieHandler, handler.ProvideFindMovieByIDHandler, handler.ProvideFindAllMoviesHandler, handler.ProvideUpdateMovieHandler, handler.ProvideDeleteMovieHandler, wire.Struct(new(handler.Handlers), "*"))

var datastoreSet = wire.NewSet(datastore.NewDB, datastore.NewDefaultDatastore, wire.Bind(new(datastore.Datastorer), new(datastore.DefaultDatastore)))

// goCloudServerSet
var goCloudServerSet = wire.NewSet(trace.AlwaysSample, server.New, server.NewDefaultDriver, wire.Bind(new(driver.Server), new(*server.DefaultDriver)))

var routerSet = wire.NewSet(handler.NewMuxRouter, wire.Bind(new(http.Handler), new(*mux.Router)))

// appHealthChecks returns a health check for the database. This will signal
// to Kubernetes or other orchestrators that the server should not receive
// traffic until the server is able to connect to its database.
func appHealthChecks(db *sql.DB) ([]health.Checker, func()) {
	dbCheck := sqlhealth.New(db)
	list := []health.Checker{dbCheck}
	return list, func() {
		dbCheck.Stop()
	}
}
