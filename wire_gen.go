// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gorilla/mux"
)

// Injectors from wireInject.go:

func setupRouter(flags *cliFlags) (*mux.Router, error) {
	envName := provideName(flags)
	dbName := provideDBName(flags)
	db, err := datastore.ProvideDB(dbName)
	if err != nil {
		return nil, err
	}
	datastoreDatastore := datastore.ProvideDS(db)
	level := provideLogLevel(flags)
	logger := provideLogger(level)
	application := app.ProvideApplication(envName, datastoreDatastore, logger)
	router := provideRouter(application)
	return router, nil
}