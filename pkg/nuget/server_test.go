package nuget

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fergusn/muzeum/pkg/events"
	"github.com/gorilla/mux"
)

func TestRouteToVersion(t *testing.T) {
	name := "abcdef"
	hit := false

	repo := &mockRepository{
		versions: func(ctx context.Context, id string) Versions {
			hit = true
			if id != name {
				t.Error()
			}
			return NewVersions([]string{"1.2.3"})
		},
	}

	repo.request(fmt.Sprintf("/content/%s/index.json", name))

	if !hit {
		t.Error("versions request not send to repository")
	}
}

func TestDownloadEmitPulledEvent(t *testing.T) {
	repo := &mockRepository{
		download: func(ctx context.Context, d, version string) (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewBuffer([]byte{1, 2, 3})), nil
		},
	}
	pulled := make(chan *events.Pulled)

	go func() {
		pulled <- <-events.Package.Pulled.Receive()
	}()

	repo.request("/content/abcd/1.1/abcd.1.1.nupkg")

	ev := <-pulled
	if ev.Package.Name != "abcd" || ev.Package.Version != "1.1" {
		t.Error("incorrect name and/or version in pulled event")
	}
}

type mockRepository struct {
	versions func(ctx context.Context, id string) Versions
	download func(ctx context.Context, id, version string) (io.ReadCloser, error)
	upload   func(ctx context.Context, nupkg io.Reader) error
	delete   func(ctx context.Context, id, version string) error
	search   func(ctx context.Context, text string) (io.ReadCloser, error)
}

func (repo *mockRepository) Versions(ctx context.Context, id string) Versions {
	return repo.versions(ctx, id)
}
func (repo *mockRepository) Download(ctx context.Context, id, version string) (io.ReadCloser, error) {
	return repo.download(ctx, id, version)
}
func (repo *mockRepository) Upload(ctx context.Context, nupkg io.Reader) error {
	return repo.upload(ctx, nupkg)
}
func (repo *mockRepository) Delete(ctx context.Context, id, version string) error {
	return repo.delete(ctx, id, version)
}
func (repo *mockRepository) Search(ctx context.Context, text string) (io.ReadCloser, error) {
	return repo.search(ctx, text)
}

func (repo *mockRepository) request(url string) *httptest.ResponseRecorder {
	router := &mux.Router{}
	route := router.NewRoute()
	srv := Server{"test", repo}
	srv.Mount(route)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rsp := httptest.NewRecorder()
	router.ServeHTTP(rsp, req)

	return rsp
}
