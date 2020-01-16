package plugins

import (
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/gorilla/mux"
)

var (
	// Plugins are repositories
	Plugins = map[string]func(router *mux.Route, name string, config map[string]interface{}, bucket driver.StorageDriver) error{}
)
