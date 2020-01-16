package main

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"

	_ "github.com/docker/distribution/registry/auth/token"
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/factory"
	_ "github.com/docker/distribution/registry/storage/driver/filesystem"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/fergusn/muzeum/internal/config"
	"github.com/fergusn/muzeum/internal/pki"
	_ "github.com/fergusn/muzeum/pkg/debian"
	_ "github.com/fergusn/muzeum/pkg/docker"
	_ "github.com/fergusn/muzeum/pkg/nuget"
	"github.com/fergusn/muzeum/pkg/plugins"
	"github.com/fergusn/muzeum/pkg/proxy"
	"github.com/fergusn/muzeum/pkg/storage"
)

func init() {
	httpsAddr := ":8443"
	httpAddr := ":8080"
	configFile := "config.yaml"

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run the muzeum server",
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("Listening on %s ...\n (HTTPS)", httpsAddr)
			log.Printf("Listening on %s ...\n (HTTP)", httpAddr)

			cfg := config.Parse(configFile)

			driver.PathRegexp = regexp.MustCompile(`^(/[\+\:A-Za-z0-9~._-]+)+$`)
			s, err := factory.Create(cfg.Storage.Type(), cfg.Storage.Parameters())
			if err != nil {
				log.Fatal(err)
			}

			router := mux.NewRouter()
			router.Use(handlers.ProxyHeaders)

			for _, repo := range cfg.Repositories {
				if len(repo.Plugin) == 1 {
					for name, cfg := range repo.Plugin {
						if register, ok := plugins.Plugins[name]; ok {
							route := router.NewRoute()
							if len(repo.Path) > 0 {
								route = route.PathPrefix(repo.Path)
							}
							if len(repo.Host) > 0 {
								route = route.Host(repo.Host)
							}

							register(route, repo.Name, cfg, storage.NewDirectoryDriver(repo.Name, s))
						}
					}
				}
			}

			router.Handle("/metrics", promhttp.Handler())

			ca, err := pki.NewCertificateAuthority(cat(cfg.Certificate.Crt, cfg.Certificate.Key))
			if err != nil {
				log.Fatal(err)
			}

			srv, err := proxy.NewProxy(ca, httpAddr, httpsAddr)
			if err != nil {
				log.Fatal(err)
			}

			err = <-srv.Serve(router)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "--config config.yaml")
	cmd.PersistentFlags().StringVar(&httpsAddr, "https", ":8443", "--https :8443")
	cmd.PersistentFlags().StringVar(&httpAddr, "http", ":8080", "--http :8080")

	cli.AddCommand(cmd)
}

func cat(files ...string) (buf []byte) {
	for _, f := range files {
		if c, err := ioutil.ReadFile(os.ExpandEnv(f)); err == nil {
			buf = append(buf, c...)
		}
	}
	return
}
