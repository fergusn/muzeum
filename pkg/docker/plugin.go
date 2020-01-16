package docker

import (
	"context"

	"github.com/docker/distribution/configuration"
	"github.com/docker/distribution/registry/handlers"

	"github.com/docker/distribution/registry/storage/driver"
	"github.com/fergusn/muzeum/pkg/plugins"
	"github.com/gorilla/mux"
)

func init() {
	plugins.Plugins["docker"] = register
}

func register(r *mux.Route, name string, config map[string]interface{}, bucket driver.StorageDriver) error {
	cfg := &configuration.Configuration{
		Compatibility: struct {
			Schema1 struct {
				TrustKey string `yaml:"signingkeyfile,omitempty"`
				Enabled  bool   `yaml:"enabled,omitempty"`
			} `yaml:"schema1,omitempty"`
		}{
			Schema1: struct {
				TrustKey string `yaml:"signingkeyfile,omitempty"`
				Enabled  bool   `yaml:"enabled,omitempty"`
			}{
				Enabled: true,
			},
		},
		Middleware: map[string][]configuration.Middleware{
			"repository": []configuration.Middleware{
				configuration.Middleware{
					Name: "muzeum",
					Options: configuration.Parameters{
						"name": name,
					},
				},
			},
		},
		Storage: configuration.Storage{
			muzeumFactory: map[string]interface{}{
				"driver": bucket,
			},
		},
	}
	if proxy, ok := config["proxy"]; ok {
		cfg.Proxy = configuration.Proxy{
			RemoteURL: proxy.(string),
		}
	}

	app := handlers.NewApp(context.Background(), cfg)
	r.Handler(app)

	return nil
}
