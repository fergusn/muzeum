package nuget

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	httpClient = http.DefaultClient
)

// NewClient creates a client repository
func NewClient(indexURL string) (Repository, error) {
	rsp, err := httpClient.Get(indexURL)

	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	idx := &ServiceIndex{}
	if err = json.NewDecoder(rsp.Body).Decode(idx); err != nil {
		return nil, err
	}

	resources := map[ResourceType]string{}
	for _, x := range idx.Resources {
		resources[x.Type] = x.ID
	}

	return &client{resources}, nil
}

type client struct {
	resources map[ResourceType]string
}

func (client *client) Versions(ctx context.Context, id string) Versions {
	url := fmt.Sprintf("%s%s/index.json", client.resources[PackageBaseAddress], id)

	rsp, err := httpClient.Get(url)
	if err != nil {
		return Versions{Status: http.StatusInternalServerError}
	}
	if rsp.StatusCode != 200 {
		return Versions{Status: rsp.StatusCode}
	}

	return Versions{ReadCloser: rsp.Body}
}

func (client *client) Download(ctx context.Context, id, version string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s%s/%s/%s.%s.nupkg", client.resources[PackageBaseAddress], id, version, id, version)

	rsp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != 200 {
		return nil, httpError{rsp.StatusCode, rsp.Status}
	}

	return rsp.Body, nil
}

func (client *client) Upload(ctx context.Context, nupkg io.Reader) error {
	return errNotImplemented
}
func (client *client) Delete(ctx context.Context, id, version string) error {
	return errNotImplemented
}
func (client *client) Search(ctx context.Context, text string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s?q=%s", client.resources[SearchQueryService], text)

	rsp, err := httpClient.Get(url)

	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != 200 {
		return nil, httpError{rsp.StatusCode, rsp.Status}
	}

	return rsp.Body, nil
}
