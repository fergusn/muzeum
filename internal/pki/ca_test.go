package pki

import (
	"bytes"
	"crypto/x509"
	"net"
	"testing"
	"time"
)

func TestSignCertContainsHostInSubjectAlternativeName(t *testing.T) {
	host := "www.example.com"

	ca := ca(t)
	crt := issue(t, ca, host, nil)

	if err := crt.VerifyHostname(host); err != nil {
		t.Error(err)
	}
}

func TestSignCertContainsIPInSubjectAlternativeName(t *testing.T) {
	ca := ca(t)

	crt := issue(t, ca, "", &net.IPAddr{IP: net.IPv4(172, 1, 2, 3)})

	if err := crt.VerifyHostname("172.1.2.3"); err != nil {
		t.Error(err)
	}
}

func TestCertificateCachedFor24Hours(t *testing.T) {
	host := "www.example.com"

	ca := ca(t)

	crt1 := issue(t, ca, host, nil)

	defer undo(timetravel(23 * time.Hour))

	crt2 := issue(t, ca, host, nil)

	if crt1.SerialNumber.Cmp(crt2.SerialNumber) != 0 {
		t.Errorf("certificate not cached. %v   %v", crt1.SerialNumber, crt2.SerialNumber)
	}
}

func TestNewCertificateIssuedAfterCacheExpiry(t *testing.T) {
	host := "www.example.com"

	ca := ca(t)

	crt1 := issue(t, ca, host, nil)

	defer undo(timetravel(25 * time.Hour))

	crt2 := issue(t, ca, host, nil)

	if crt1.SerialNumber.Cmp(crt2.SerialNumber) == 0 {
		t.Errorf("certificate not renewed. %v == %v", crt1.NotAfter, crt2.NotAfter)
	}
}

func ca(t *testing.T) *CertificateAuthority {
	var b bytes.Buffer
	Generate(&b, "muzeum")

	ca, err := NewCertificateAuthority(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return ca
}

func issue(t *testing.T, ca *CertificateAuthority, host string, addr *net.IPAddr) *x509.Certificate {
	pem, err := ca.Sign(host, addr)
	if err != nil {
		t.Fatal(err)
	}

	crt, err := x509.ParseCertificate(pem.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}

	return crt
}

func undo(clock func() time.Time) {
	now = clock
}

func timetravel(d time.Duration) func() time.Time {
	clock := now
	now = func() time.Time {
		return clock().Add(d)
	}
	return clock
}
