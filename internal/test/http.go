package test

import (
	"net/http"
)

// HTTPClient created a http.Client with the provided RoundTrip
func HTTPClient(roundTrip func(r *http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{Transport: roundTripFunc(roundTrip)}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

