package debian

import (
	"errors"
	"net/url"

	"github.com/docker/distribution/registry/storage/driver"
	muzeum "github.com/fergusn/muzeum/pkg/plugins"
	"github.com/gorilla/mux"
)

var (
	errConfiguration = errors.New("Debian Repository require proxy configuration")
)

func init() {
	muzeum.Plugins["debian"] = register
}

func register(rt *mux.Route, name string, config map[string]interface{}, bucket driver.StorageDriver) error {
	proxy, ok := config["proxy"]
	if !ok {
		return errConfiguration
	}
	raw, ok := proxy.(string)
	if !ok {
		return errConfiguration
	}

	url, err := url.Parse(raw)
	if err != nil {
		return err
	}


	repo := NewRemote(raw, bucket)
	srv := NewServer(name, url, repo)

	srv.Mount(rt)

	return nil
}
