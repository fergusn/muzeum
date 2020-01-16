package cache

import (
	"bytes"
	"context"
	"crypto/rand"
	"github.com/fergusn/muzeum/internal/test"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNotModifiedReturnCacheddResource(t *testing.T) {
	etag := "abcdefg"
	status := http.StatusOK
	body := make([]byte, 10)
	rand.Read(body)

	client := test.HTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/ubuntu/dists/xenial/InRelease" {
			return &http.Response{StatusCode: http.StatusNotFound}, nil
		}
		rsp := &http.Response{
			StatusCode: status,
			Header: http.Header{
				httpHeaderETag: []string{etag},
			},
		}
		if status == http.StatusOK {
			rsp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}
		return rsp, nil
	})

	res := NewResourceWithHTTPClient(client, "http://archive.ubuntu.com/ubuntu/dists/xenial/InRelease")

	rsp1, _, err := res.Get(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	defer rsp1.Close()

	expected := body
	body = make([]byte, 15)
	rand.Read(body)
	status = http.StatusNotModified

	rsp2, _, err := res.Get(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	defer rsp2.Close()

	actual, err := ioutil.ReadAll(rsp2)
	if err != nil || bytes.Compare(expected, actual) != 0 {
		t.Errorf("expected cached resource %v, got %v", expected, actual)
	}
}

func TestModifiedUpdatesCache(t *testing.T) {
	etag := "abcdef"
	status := http.StatusOK
	body := make([]byte, 10)
	rand.Read(body)

	client := test.HTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/ubuntu/dists/xenial/InRelease" {
			return &http.Response{StatusCode: http.StatusNotFound}, nil
		}
		rsp := &http.Response{
			StatusCode: status,
			Header: http.Header{
				httpHeaderETag: []string{etag},
			},
		}
		if status == http.StatusOK {
			rsp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}
		return rsp, nil
	})

	res := NewResourceWithHTTPClient(client, "http://archive.ubuntu.com/ubuntu/dists/xenial/InRelease")

	rsp1, _, err := res.Get(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	defer rsp1.Close()

	etag = "ghijkl"
	rand.Read(body)

	rsp2, _, err := res.Get(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	defer rsp2.Close()

	actual, err := ioutil.ReadAll(rsp2)
	if err != nil || bytes.Compare(body, actual) != 0 {
		t.Errorf("expected cached resource %v, got %v", body, actual)
	}
}
