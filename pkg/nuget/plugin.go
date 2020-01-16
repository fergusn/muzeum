package nuget

import (
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/fergusn/muzeum/pkg/plugins"
	"github.com/gorilla/mux"
)

func init() {
	plugins.Plugins["nuget"] = register
}

func register(route *mux.Route, name string, config map[string]interface{}, bucket driver.StorageDriver) error {
	var repo Repository
	if proxy, ok := config["proxy"]; ok {
		repo, _ = NewRemote(proxy.(string), bucket)
	} else {
		repo = NewLocal(bucket)
	}

	server := Server{name, repo}
	server.Mount(route)

	return nil
}
