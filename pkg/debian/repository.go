package debian

import (
	"context"
	"io"

	"github.com/fergusn/muzeum/pkg/model"
)

// Repository is a Debian repositry
type Repository interface {
	// Release reads the InRelease file from the repository
	Release(ctx context.Context, dist string) (io.ReadCloser, error)

	// Index reads the Index file for a disttribution/component/architecture
	Index(ctx context.Context, dist, comp, arch, compression string) (io.ReadCloser, error)

	// File reads the deb package
	File(ctx context.Context, path string) (io.ReadCloser, *model.Package, error)
}
