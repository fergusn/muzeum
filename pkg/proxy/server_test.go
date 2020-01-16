package proxy

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/fergusn/muzeum/internal/pki"
)

func TestHttpRequestViaHttpProxy(t *testing.T) {
	proxy, ca := proxy(t)
	client := client(ca.Certificate, "http", proxy.http.Addr().(*net.TCPAddr).Port)

	rsp := get(t, client, "http://www.google.com/")

	assert(t, "www.google.com", rsp)
}

func TestHttpsRequestViaHttpProxy(t *testing.T) {
	proxy, ca := proxy(t)
	client := client(ca.Certificate, "http", proxy.http.Addr().(*net.TCPAddr).Port)

	rsp := get(t, client, "https://www.google.com/")

	assert(t, "www.google.com", rsp)
}

func TestHttpsRequestViaHttpsProxy(t *testing.T) {
	proxy, ca := proxy(t)
	client := client(ca.Certificate, "https", proxy.https.Addr().(*net.TCPAddr).Port)

	rsp := get(t, client, "https://www.google.com/")

	assert(t, "www.google.com", rsp)
}

func proxy(t *testing.T) (*Proxy, *pki.CertificateAuthority) {
	pem := &bytes.Buffer{}
	pki.Generate(pem, "localhost")

	ca, err := pki.NewCertificateAuthority(pem.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	proxy, err := NewProxy(ca, ":0", ":0")
	if err != nil {
		t.Fatal(err)
	}

	proxy.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Host))
	}))

	return proxy, ca
}

func get(t *testing.T, client *http.Client, url string) string {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	rsp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if rsp.StatusCode != http.StatusOK {
		t.Error("Response should be OK")
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return string(body)
}

func client(ca *x509.Certificate, protocol string, port int) *http.Client {
	purl, _ := url.Parse(fmt.Sprintf("%s://localhost:%d/", protocol, port))
	transport := &http.Transport{
		Proxy: http.ProxyURL(purl),
		TLSClientConfig: &tls.Config{
			RootCAs: x509.NewCertPool(),
		},
	}
	transport.TLSClientConfig.RootCAs.AddCert(ca)

	return &http.Client{Transport: transport}
}

func assert(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("expected: %s, got: %s", expected, actual)
	}
}
