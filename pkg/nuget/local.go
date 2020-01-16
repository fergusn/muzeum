package nuget

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/docker/distribution/registry/storage/driver"
)

// NewLocalRepository initialize a new local repository
func NewLocal(storage driver.StorageDriver) Repository {
	return &local{storage}
}

type local struct {
	storage driver.StorageDriver
}

func (repo *local) Versions(ctx context.Context, id string) Versions {
	path := "/" + id
	xs, err := repo.storage.List(context.TODO(), path)
	if err != nil {
		return Versions{nil, http.StatusInternalServerError}
	}

	for i, x := range xs {
		xs[i] = strings.TrimPrefix(x, path+"/")
	}
	return NewVersions(xs)
}

func (repo *local) Download(ctx context.Context, id, version string) (io.ReadCloser, error) {
	return repo.storage.Reader(context.TODO(), path(id, version), 0)
}

func (repo *local) Upload(ctx context.Context, nupkg io.Reader) error {
	buf, err := ioutil.ReadAll(nupkg) // nupkg files are generally smallish, so we sacrafice memory for simplicity

	if err != nil {
		return err
	}

	archive, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))

	if err != nil {
		return err
	}

	pkg, spec, err := nuspec(archive)

	if err != nil {
		return err
	}

	repo.storage.PutContent(ctx, path(pkg.Metadata.ID, pkg.Metadata.Version), spec)
	return repo.storage.PutContent(ctx, path(pkg.Metadata.ID, pkg.Metadata.Version), buf)
}

func (repo *local) Delete(ctx context.Context, id, version string) error {
	return repo.storage.Delete(ctx, path(id, version))
}

func (repo *local) Search(ctx context.Context, text string) (io.ReadCloser, error) {
	results := []string{}
	repo.storage.Walk(ctx, "/", func(fi driver.FileInfo) error {
		if strings.Contains(fi.Path(), text) {
			results = append(results, fi.Path())
			repo.storage.List(ctx, fi.Path())
		}
		return nil
	})

	return nil, nil
}
