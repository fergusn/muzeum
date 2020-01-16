package metrics

import (
	"github.com/fergusn/muzeum/pkg/events"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	pulled = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "package_pulled",
		Help: "The number of times a package was pulled",
	}, []string{"type", "registry", "name", "version", "location"})
)

func init() {
	go inc(events.Package.Pulled.Receive())
}

func inc(events <-chan *events.Pulled) {
	for pull := range events {
		pulled.With(prometheus.Labels{
			"type":     pull.Package.Type,
			"registry": pull.Registry,
			"name":     pull.Package.Name,
			"version":  pull.Package.Version,
			"location": pull.Location,
		}).Inc()
	}
}
