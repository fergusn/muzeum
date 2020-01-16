package cache

import (
	"context"
	"io"
	"sync"

	"github.com/docker/distribution/registry/storage/driver"
)

// Cache is a read-through cache that cache packages locally
type Cache interface {
	Read(ctx context.Context, path string, loader func() (io.ReadCloser, error)) (io.ReadCloser, error)
}

// NewCache return and instance of Cache
func NewCache(storage driver.StorageDriver) Cache {
	return &cache{
		storage:  storage,
		inflight: make(map[string]struct{}),
	}
}

type cache struct {
	storage  driver.StorageDriver
	inflight map[string]struct{}
	mu       sync.Mutex
}


// Read an file from the cache. If it does not exists, read it from loader and prime the cache
func (c *cache) Read(ctx context.Context, path string, loader func() (io.ReadCloser, error)) (io.ReadCloser, error) {
	if rd, err := c.storage.Reader(ctx, path, 0); err == nil {
		return rd, nil
	}

	// if we have an in-flight request for package, send another request and don't handle cache for this one
	c.mu.Lock()
	if _, ok := c.inflight[path]; ok {	
		c.mu.Unlock()
		return loader()
	}

	c.inflight[path] = struct{}{}
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.inflight, path)
		c.mu.Unlock()
	}()

	rd, err := loader()
	if err != nil {
		return nil, err
	}

	wr, err := c.storage.Writer(ctx, path, false)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(wr, rd)
	if err != nil {
		return nil, err
	}

	err = wr.Commit()
	if err != nil {
		return nil, err
	}

	return c.storage.Reader(ctx, path, 0)
}
