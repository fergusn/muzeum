package metrics

import (
	"testing"

	"github.com/fergusn/muzeum/pkg/events"
	"github.com/fergusn/muzeum/pkg/model"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPulledEventIncrementPulledCounter(t *testing.T) {
	chx := make(chan *events.Pulled, 2)
	chx <- &events.Pulled{
		Registry: "testregistry",
		Package: &model.Package{
			Type: "testtype",
			Name: "testname",
		},
	}
	close(chx)
	inc(chx)

	ch := make(chan prometheus.Metric, 2)
	pulled.Collect(ch)
	metric := dto.Metric{}
	(<-ch).Write(&metric)

	if *metric.Counter.Value != 1 {
		t.Error("pulled counter should be 1")
	}
	assertLabel(t, metric, "registry", "testregistry")
	assertLabel(t, metric, "type", "testtype")
	assertLabel(t, metric, "name", "testname")
}

func assertLabel(t *testing.T, metric dto.Metric, name, value string) {
	if !contains(metric, name, value) {
		t.Errorf("counter must have label %s=%s", name, value)
	}
}

func contains(metric dto.Metric, name, value string) bool {
	for _, label := range metric.Label {
		if *label.Name == name && *label.Value == value {
			return true
		}
	}
	return false
}
