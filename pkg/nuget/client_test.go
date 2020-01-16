package nuget

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func mock(roundTrip func(r *http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{Transport: roundTripFunc(roundTrip)}
}

func TestVersions(t *testing.T) {
	httpClient = mock(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v3/index.json" {
			return &http.Response{
				StatusCode: 200,
				Body:       read(t, "index.json"),
			}, nil
		} else if r.URL.Path == "/v3-flatcontainer/xunit/index.json" {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{ "versions": [ "1.9.1", "2.4.1" ] }`)),
			}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound}, nil
	})

	client, err := NewClient("https://api.nuget.org/v3/index.json")

	if err != nil {
		t.Fatal(err)
	}

	rsp := client.Versions(context.TODO(), "xunit")
	vs, err := rsp.Unmarshal()
	if err != nil {
		t.Fatal(err)
	}

	if len(vs) != 2 || vs[0] != "1.9.1" || vs[1] != "2.4.1" {
		t.Errorf("expected versions [1.9.1 2.4.1] got %v", vs)
	}
}

func TestDownLoad(t *testing.T) {
	httpClient = mock(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v3/index.json" {
			return &http.Response{
				StatusCode: 200,
				Body:       read(t, "index.json"),
			}, nil
		} else if r.URL.Path == "/v3-flatcontainer/xunit/2.4.1/xunit.2.4.1.nupkg" {
			return &http.Response{
				StatusCode: 200,
				Body:       read(t, "xunit.2.4.1.nupkg"),
			}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound}, nil
	})

	client, err := NewClient("https://api.nuget.org/v3/index.json")

	if err != nil {
		t.Fatal(err)
	}

	r, err := client.Download(context.TODO(), "xunit", "2.4.1")

	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	buf, err := ioutil.ReadAll(r)

	if len(buf) < 1 {
		t.Error("Package not downloaded")
	}
}

func TestSearch(t *testing.T) {
	client, err := NewClient("https://api.nuget.org/v3/index.json")
	if err != nil {
		t.Fatal(err)
	}

	client.Search(context.TODO(), "xunit")

}
