package debian

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/fergusn/muzeum/internal/test"
)

func TestReleaseGetFromUpstreamHttpRepository(t *testing.T) {
	inrelease := []byte{1, 2, 3, 4, 5}

	httpClient = test.HTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/ubuntu/dists/bionic/InRelease" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader(inrelease)),
			}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound}, nil
	})

	c := NewClient("http://archive.ubuntu.com/ubuntu")

	rd, err := c.Release(context.TODO(), "bionic")

	if err != nil {
		t.Fatal(err)
	}
	defer rd.Close()

	actual, err := ioutil.ReadAll(rd)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(actual, inrelease) != 0 {
		t.Errorf("release should read from upstream. expected %v, got %v", inrelease, actual)
	}
}

func TestIndexReadsPackageMetadata(t *testing.T) {
	httpClient = test.HTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/apt/dists/kubernetes-xenial/main/binary-amd64/Packages.gz" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       read(t, "Packages.gz"),
			}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound}, nil
	})

	c := NewClient("https://packages.cloud.google.com/apt")

	_, err := c.Index(context.TODO(), "kubernetes-xenial", "main", "amd64", "gz")

	if err != nil {
		t.Fatal(err)
	}
	s := c.(*client)

	if s.packages["pool/cri-tools_1.11.0-00_amd64_768e5551f9badfde12b10c42c88afb45c412c1bf307a5985a4b29f4499d341bd.deb"].Version != "1.11.0-00" {
		t.Error("packages not loaded")
	}
}

func read(t *testing.T, name string) io.ReadCloser {
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return f
}
