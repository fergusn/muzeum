package debian

import (
	"bytes"
	"strings"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/fergusn/muzeum/pkg/cache"
	"github.com/fergusn/muzeum/pkg/model"
)

var (
	httpClient = http.DefaultClient
)

type client struct {
	url      string
	releases map[string]cache.Resource
	indices  map[id]cache.Resource
	packages map[string]*model.Package

	mu sync.RWMutex
}

type id struct {
	name, version, arch string
}

// NewClient initialize a new Debian client repository
func NewClient(url string) Repository {
	return &client{
		url:      url,
		releases: map[string]cache.Resource{},
		indices:  map[id]cache.Resource{},
		packages: map[string]*model.Package{},
	}
}

// Release get the InRelease file for the distrubution, using etag to optimize
func (c *client) Release(ctx context.Context, dist string) (io.ReadCloser, error) {
	c.mu.RLock()
	r, ok := c.releases[dist]
	c.mu.RUnlock()

	if !ok {
		r = cache.NewResourceWithHTTPClient(httpClient, concat(c.url, "dists", dist, "InRelease"))

		c.mu.Lock()
		c.releases[dist] = r
		c.mu.Unlock()
	}

	rd, _, err := r.Get(context.TODO())
	return rd, err
}

// Index get the package index for the distribution/component/architecture, using etag to optimize
func (c *client) Index(ctx context.Context, dist, comp, arch, compression string) (io.ReadCloser, error) {
	c.mu.RLock()
	rc, exist := c.indices[id{dist, comp, arch}]
	c.mu.RUnlock()

	if !exist {
		c.mu.Lock()
		rc = cache.NewResourceWithHTTPClient(httpClient, concat(c.url, "dists", dist, comp, "binary-"+arch, "Packages."+compression))
		c.indices[id{dist, comp, arch}] = rc
		c.mu.Unlock()
	}

	r, updated, err := rc.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !updated {
		return r, nil
	}

	buf, _ := ioutil.ReadAll(r)
	r.Close()

	gz, err := decompress(bytes.NewReader(buf), compression)

	rd := NewControlFileReader(gz)

	px := map[string]*model.Package{}
	for {
		if pkg, more := rd.Read(); more {
			px[pkg.Filename()] = &model.Package{
				Type:    "debian",
				Name:    pkg.Package(),
				Version: pkg.Version(),
			}
		} else {
			break
		}
	}
	c.mu.Lock()
	c.packages = px
	c.mu.Unlock()

	return ioutil.NopCloser(bytes.NewReader(buf)), nil
}

func (c *client) File(ctx context.Context, path string) (io.ReadCloser, *model.Package, error) {
	rsp, err := httpClient.Get(concat(c.url, path))

	if err != nil {
		return nil, nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if pkg, ok := c.packages[strings.TrimLeft(path, "/")]; ok {
		return rsp.Body, pkg, nil
	}

	log.Printf("no package index for path %s", path)
	return rsp.Body, nil, nil
}
