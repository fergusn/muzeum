package nuget

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
)

type Versions struct {
	io.ReadCloser
	Status int
}

func (v *Versions) Unmarshal() ([]string, error) {
	versions := struct {
		Versions []string `json:"versions"`
	}{}

	if err := json.NewDecoder(v).Decode(&versions); err != nil {
		return nil, err
	}
	return versions.Versions, nil
}
func NewVersions(versions []string) Versions {
	vx := struct {
		Versions []string `json:"versions"`
	}{versions}
	b := &bytes.Buffer{}

	json.NewEncoder(b).Encode(vx)

	return Versions{ioutil.NopCloser(b), 0}
}

// Repository is an interface for a NuGet repository
type Repository interface {
	Versions(ctx context.Context, id string) Versions
	Download(ctx context.Context, id, version string) (io.ReadCloser, error)
	Upload(ctx context.Context, nupkg io.Reader) error
	Delete(ctx context.Context, id, version string) error
	Search(ctx context.Context, text string) (io.ReadCloser, error)
}
