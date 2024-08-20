package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"regexp"
	"time"
)

type CertInfo struct {
	ExpiryDate           string
	IssuedDate           string
	DaysRemaining        int
	CertificateAuthority string
}

func GetCertificateInfo(ctx context.Context, domain string) (*CertInfo, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", domain+":443", config)
	if err != nil {
		return nil, NewCertError(domain, err)
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]
	now := time.Now()
	daysRemaining := int(cert.NotAfter.Sub(now).Hours() / 24)

	certInfo := &CertInfo{
		ExpiryDate:           cert.NotAfter.Format(time.RFC3339),
		IssuedDate:           cert.NotBefore.Format(time.RFC3339),
		DaysRemaining:        daysRemaining,
		CertificateAuthority: cert.Issuer.CommonName,
	}

	return certInfo, nil
}

func NewCertError(domain string, err error) error {
	return fmt.Errorf("error processing certificate for domain %s: %w", domain, err)
}

func IsValidDomain(domain string) bool {
	re := regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)
	return re.MatchString(domain)
}
