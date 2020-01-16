package pki

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"sync"
	"time"
)

var (
	now = time.Now
)

// CertificateAuthority to issue self signed certificates
type CertificateAuthority struct {
	Certificate *x509.Certificate
	key         *rsa.PrivateKey
	cache       map[string]*tls.Certificate
	lock        sync.RWMutex
}

// NewCertificateAuthority creates a CA that can be used to sign certificates
func NewCertificateAuthority(data []byte) (*CertificateAuthority, error) {
	var blk *pem.Block
	var crt *x509.Certificate
	var key *rsa.PrivateKey
	var err error

	rest := data
	for len(rest) > 0 {
		blk, rest = pem.Decode(rest)
		if blk.Type == "CERTIFICATE" {
			crt, err = x509.ParseCertificate(blk.Bytes)
		}
		if blk.Type == "RSA PRIVATE KEY" {
			key, err = x509.ParsePKCS1PrivateKey(blk.Bytes)
		}
		if err != nil {
			return nil, err
		}
	}

	if crt == nil || key == nil {
		return nil, errors.New("data must include certificate and key")
	}

	ca := &CertificateAuthority{
		Certificate: crt,
		key:         key,
		cache:       make(map[string]*tls.Certificate),
		lock:        sync.RWMutex{},
	}

	return ca, nil
}

func (ca *CertificateAuthority) sign(host string, addr net.Addr) (*tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	names := []string{}
	ips := []net.IP{}

	if len(host) > 0 {
		names = append(names, host)
	} else if addr, ok := addr.(*net.IPAddr); ok {
		ips = append(ips, addr.IP)
	}

	skid, err := skid(key.PublicKey)
	if err != nil {
		return nil, err
	}
	sn, err := sn()
	if err != nil {
		return nil, err
	}

	crt := &x509.Certificate{
		SerialNumber: sn,
		DNSNames:     names,
		IPAddresses:  ips,
		Subject:      pkix.Name{CommonName: host},
		SubjectKeyId: skid,
		NotBefore:    now().Add(time.Minute * -10),
		NotAfter:     now().Add(time.Hour * 24),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	bts, err := x509.CreateCertificate(rand.Reader, crt, ca.Certificate, &key.PublicKey, ca.key)
	if err != nil {
		return nil, err
	}

	pair, err := tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: bts}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
	)
	return &pair, err
}

// Sign a TLS certificate for the domain name
func (ca *CertificateAuthority) Sign(host string, addr net.Addr) (*tls.Certificate, error) {
	ca.lock.RLock()

	if c, exist := ca.cache[host]; exist {
		crt, err := x509.ParseCertificate(c.Certificate[0])
		if err == nil && now().Add(time.Minute*10).Before(crt.NotAfter) {
			ca.lock.RUnlock()
			return c, nil
		}
	}
	ca.lock.RUnlock()
	ca.lock.Lock()
	defer ca.lock.Unlock()

	crt, err := ca.sign(host, addr)
	if err != nil {
		return nil, err
	}

	ca.cache[host] = crt
	return crt, nil
}

// Generate a CA certificate with common name cn and write the PEM to w
func Generate(w io.Writer, cn string) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	skid, err := skid(key.PublicKey)
	if err != nil {
		panic(err)
	}
	sn, err := sn()
	if err != nil {
		panic(err)
	}

	ca := &x509.Certificate{
		SerialNumber:          sn,
		Subject:               pkix.Name{CommonName: cn},
		SubjectKeyId:          skid,
		NotBefore:             now(),
		NotAfter:              now().AddDate(2, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}

	crt, err := x509.CreateCertificate(rand.Reader, ca, ca, key.Public(), key)
	if err != nil {
		panic(err)
	}

	pem.Encode(w, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: crt,
	})

	pem.Encode(w, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func sn() (*big.Int, error) {
	max := big.Int{}
	return rand.Int(rand.Reader, max.Exp(big.NewInt(2), big.NewInt(20*8), nil)) // sn is 20 bytes
}

func skid(pub rsa.PublicKey) (bytes []byte, err error) {
	if bytes, err := asn1.Marshal(pub); err == nil {
		hash := sha1.Sum(bytes)
		return hash[:], nil
	}
	return
}
