package debian

import (
	"context"
	"io"

	"github.com/docker/distribution/registry/storage/driver"
	"github.com/fergusn/muzeum/pkg/cache"
	"github.com/fergusn/muzeum/pkg/model"
)

type remote struct {
	Repository
	cache cache.Cache
}

// NewRemote initialize a remote repository
func NewRemote(url string, storage driver.StorageDriver) Repository {
	return &remote{
		Repository: NewClient(url),
		cache:      cache.NewCache(storage),
	}
}

// File read the package from the upstream repository and cache it locally.
func (r *remote) File(ctx context.Context, path string) (rd io.ReadCloser, pkg *model.Package, err error) {
	// TODO: Return package metadata
	rd, err = r.cache.Read(ctx, path, func() (rd io.ReadCloser, err error) {
		rd, _, err = r.Repository.File(ctx, path)
		return
	})
	return
}
