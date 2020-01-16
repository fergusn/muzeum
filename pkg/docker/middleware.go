package docker

import (
	"context"
	"fmt"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/auth"
	repomiddleware "github.com/docker/distribution/registry/middleware/repository"
	"github.com/fergusn/muzeum/pkg/events"
	"github.com/fergusn/muzeum/pkg/model"
)

func init() {
	repomiddleware.Register("muzeum", func(ctx context.Context, repository distribution.Repository, options map[string]interface{}) (distribution.Repository, error) {
		return repositoryDecorator{repository, options["name"].(string)}, nil
	})
}

type repositoryDecorator struct {
	distribution.Repository
	name string
}
type manifestDecorator struct {
	distribution.ManifestService
	repository distribution.Repository
}
type tagsDecorator struct {
	distribution.TagService
	repository distribution.Repository
	name       string
}

func (d repositoryDecorator) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	inner, err := d.Repository.Manifests(ctx, options...)
	return manifestDecorator{inner, d.Repository}, err
}
func (d repositoryDecorator) Tags(ctx context.Context) distribution.TagService {
	inner := d.Repository.Tags(ctx)
	return tagsDecorator{inner, d.Repository, d.name}
}

func (d tagsDecorator) Tag(ctx context.Context, tag string, desc distribution.Descriptor) error {
	if resources := auth.AuthorizedResources(ctx); resources != nil {
		fmt.Println(resources)
	}
	return d.TagService.Tag(ctx, tag, desc)
}

func (d tagsDecorator) Get(ctx context.Context, tag string) (distribution.Descriptor, error) {
	events.Package.Pulled.Emit(&events.Pulled{
		Registry: d.name,
		Package: &model.Package{
			Type:      "docker",
			Namespace: d.repository.Named().Name(),
			Version:   tag,
		},
		Location: "", // middleware.ClientIP(ctx),
	})

	return d.TagService.Get(ctx, tag)
}
