package cache

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

// ErrHTTP is typically a 4xx and 5xx HTTP response.
type ErrHTTP struct {
	StatusCode int
	Status     string
}

func (e ErrHTTP) Error() string {
	return e.Status
}

// Resource is a HTTP resource that cache responses. It currently only support ETag and does not support Cache-Control yet.
type Resource interface {
	Get(ctx context.Context) (io.ReadCloser, bool, error)
}

// NewResource created a Resource for url that will use HTTP caching policy
func NewResource(url string) Resource {
	return &resource{
		client: http.DefaultClient,
		url:    url,
	}
}

// NewResourceWithHTTPClient created a Resource for url that will use HTTP caching policy. It uses the provided http client.
func NewResourceWithHTTPClient(client *http.Client, url string) Resource {
	return &resource{
		client: client,
		url:    url,
	}
}

var (
	httpHeaderETag        = "Etag"
	httpHeaderIfNoneMatch = "If-None-Match"
)

type resource struct {
	client *http.Client
	url    string
	etag   string
	cache  []byte
	mu     sync.RWMutex
}

func (r *resource) Get(ctx context.Context) (io.ReadCloser, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return nil, false, err
	}

	r.mu.RLock()
	etag := r.etag
	cache := r.cache
	r.mu.RUnlock()

	if len(r.etag) > 0 {
		req.Header.Add(httpHeaderIfNoneMatch, etag)
	}

	rsp, err := r.client.Do(req)

	if err != nil {
		return nil, false, err
	}

	if rsp.StatusCode == http.StatusNotModified {
		return ioutil.NopCloser(bytes.NewReader(cache)), false, nil
	}

	if rsp.StatusCode != http.StatusOK {
		return nil, false, ErrHTTP{rsp.StatusCode, rsp.Status}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.etag = rsp.Header.Get(httpHeaderETag)

	if len(r.etag) == 0 {
		return rsp.Body, true, nil
	}

	r.cache, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, false, err
	}

	return ioutil.NopCloser(bytes.NewReader(r.cache)), true, nil
}
