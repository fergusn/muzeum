package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/fergusn/muzeum/internal/pki"
)

// Proxy server serve HTTP CONNECT requests
type Proxy struct {
	ca     *pki.CertificateAuthority
	http   net.Listener
	https  net.Listener
	accept chan net.Conn
}

// NewProxy return a listner that will handle HTTP CONNECT
func NewProxy(ca *pki.CertificateAuthority, http, https string) (p *Proxy, err error) {
	p = &Proxy{
		ca:     ca,
		accept: make(chan net.Conn, 5),
	}

	if p.http, err = net.Listen("tcp", http); err != nil {
		return nil, err
	}
	if p.https, err = net.Listen("tcp", https); err != nil {
		return nil, err
	}

	go func() {
		for {
			con, err := p.https.Accept()
			if err != nil {
				close(p.accept)
				return
			}
			p.accept <- con
		}
	}()

	return p, nil
}

func (p *Proxy) Addr() net.Addr {
	return p.https.Addr()
}

func (p *Proxy) Accept() (net.Conn, error) {
	if conn, more := <-p.accept; more {
		return conn, nil
	}
	return nil, fmt.Errorf("closed")
}

func (p *Proxy) Close() error {
	close(p.accept)
	return p.https.Close()
}

// Serve serve request on http and https
func (p *Proxy) Serve(handler http.Handler) chan error {
	cfg := &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return p.ca.Sign(hello.ServerName, hello.Conn.LocalAddr())
		},
	}

	err := make(chan error)
	srv := http.Server{
		Handler: p.proxy(handler),
	}
	go func() { err <- srv.Serve(p.http) }()
	go func() { err <- srv.Serve(tls.NewListener(p, cfg)) }()

	return err
}

func (p *Proxy) proxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			if h, ok := w.(http.Hijacker); ok {
				if c, rw, err := h.Hijack(); err == nil {
					rw.WriteString("HTTP/1.1 200 OK\n\n")
					rw.Flush()

					p.accept <- c
				}
			}
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
