package storage

import (
	"context"
	"io"
	"strings"

	"github.com/docker/distribution/registry/storage/driver"
)

const name = "directory"

// NewDirectoryDriver wrap inner with a decorator that prepend the path to all operation, i.e. effectively storing items in a sub-directory
func NewDirectoryDriver(path string, inner driver.StorageDriver) driver.StorageDriver {
	return &directoryDriver{"/" + path, inner}
}

type directoryDriver struct {
	path  string
	inner driver.StorageDriver
}

func (d directoryDriver) Name() string {
	return name
}

func (d directoryDriver) GetContent(ctx context.Context, path string) ([]byte, error) {
	return d.inner.GetContent(ctx, d.subpath(path))
}

func (d directoryDriver) PutContent(ctx context.Context, path string, content []byte) error {
	return d.inner.PutContent(ctx, d.subpath(path), content)
}

func (d directoryDriver) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	return d.inner.Reader(ctx, d.subpath(path), offset)
}

func (d directoryDriver) Writer(ctx context.Context, path string, append bool) (driver.FileWriter, error) {
	return d.inner.Writer(ctx, d.subpath(path), append)
}

func (d directoryDriver) Stat(ctx context.Context, path string) (driver.FileInfo, error) {
	fi, err := d.inner.Stat(ctx, d.subpath(path))
	return fileInfoDecorator{fi, d.path}, err
}

func (d directoryDriver) List(ctx context.Context, path string) ([]string, error) {
	xs, err := d.inner.List(ctx, strings.TrimRight(d.subpath(path), "/"))
	for i, x := range xs {
		xs[i] = strings.TrimPrefix(x, d.path)
	}
	return xs, err
}

func (d directoryDriver) Move(ctx context.Context, sourcePath string, destPath string) error {
	return d.inner.Move(ctx, d.subpath(sourcePath), d.subpath(destPath))
}

func (d directoryDriver) Delete(ctx context.Context, path string) error {
	return d.inner.Delete(ctx, d.subpath(path))
}

func (d directoryDriver) URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error) {
	return d.inner.URLFor(ctx, d.subpath(path), options)
}

func (d directoryDriver) Walk(ctx context.Context, path string, f driver.WalkFn) error {
	return d.inner.Walk(ctx, d.subpath(path), f)
}

func (d directoryDriver) subpath(path string) string {
	return d.path + path
}

type fileInfoDecorator struct {
	driver.FileInfo
	path string
}

func (d fileInfoDecorator) Path() string {
	return strings.TrimPrefix(d.FileInfo.Path(), d.path)
}
