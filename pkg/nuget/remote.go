package nuget

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/distribution/registry/storage/driver"
	"github.com/fergusn/muzeum/pkg/cache"
)

// NewRemote initialize a repository that fetch and cache packages from upstream
func NewRemote(remoteURL string, storage driver.StorageDriver) (Repository, error) {
	client, err := NewClient(remoteURL)

	if err != nil {
		return nil, err
	}

	return &remote{client, cache.NewCache(storage)}, nil
}

type remote struct {
	Repository
	cache cache.Cache
}

func (r *remote) Download(ctx context.Context, id, version string) (io.ReadCloser, error) {
	path := fmt.Sprintf("/%s/%s/%s.%s.nupkg", id, version, id, version)

	return r.cache.Read(ctx, path, func() (io.ReadCloser, error) {
		return r.Repository.Download(ctx, id, version)
	})
}
