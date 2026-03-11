package certificate

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"time"
)

// CertificateData holds the parsed certificate data for storage.
type CertificateData struct {
	ID           string
	VaultName    string
	MountPath    string
	SerialNumber string
	CommonName   string
	IssuerCN     string
	SubjectCN    string
	SANs         string // JSON array
	KeyType      string
	KeyBits      int
	SigningAlgo   string
	NotBefore    time.Time
	NotAfter     time.Time
	IsRevoked    bool
	RevokedAt    *time.Time
	CertPEM      string
	CollectedAt  time.Time
}

// ConvertCert parses a Vault cert response into CertificateData.
func ConvertCert(resp *CertResponse, vaultName, mountPath, serial string, collectedAt time.Time) (*CertificateData, error) {
	data := &CertificateData{
		ID:           fmt.Sprintf("%s/%s/%s", vaultName, mountPath, serial),
		VaultName:    vaultName,
		MountPath:    mountPath,
		SerialNumber: serial,
		CertPEM:      resp.Data.Certificate,
		CollectedAt:  collectedAt,
	}

	// Check revocation
	if resp.Data.RevocationTime > 0 {
		data.IsRevoked = true
		t := time.Unix(resp.Data.RevocationTime, 0)
		data.RevokedAt = &t
	}

	// Parse X.509 certificate
	block, _ := pem.Decode([]byte(resp.Data.Certificate))
	if block == nil {
		slog.Warn("Failed to decode PEM for certificate", "serial", serial)
		return data, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		slog.Warn("Failed to parse X.509 certificate", "serial", serial, "error", err)
		return data, nil
	}

	data.CommonName = cert.Subject.CommonName
	data.IssuerCN = cert.Issuer.CommonName
	data.SubjectCN = cert.Subject.CommonName
	data.NotBefore = cert.NotBefore
	data.NotAfter = cert.NotAfter
	data.SigningAlgo = cert.SignatureAlgorithm.String()

	// Extract key type and bits
	data.KeyType, data.KeyBits = extractKeyInfo(cert)

	// Extract SANs
	data.SANs = extractSANs(cert)

	return data, nil
}

// extractKeyInfo returns the key type and bit size from a certificate.
func extractKeyInfo(cert *x509.Certificate) (string, int) {
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return "RSA", pub.N.BitLen()
	case *ecdsa.PublicKey:
		return "ECDSA", pub.Curve.Params().BitSize
	case ed25519.PublicKey:
		return "ED25519", 256
	default:
		return "", 0
	}
}

// extractSANs collects DNS names, IP addresses, and email addresses into a JSON array.
func extractSANs(cert *x509.Certificate) string {
	var sans []string

	for _, dns := range cert.DNSNames {
		sans = append(sans, dns)
	}
	for _, ip := range cert.IPAddresses {
		sans = append(sans, ip.String())
	}
	for _, email := range cert.EmailAddresses {
		sans = append(sans, email)
	}
	for _, uri := range cert.URIs {
		sans = append(sans, uri.String())
	}

	if len(sans) == 0 {
		return ""
	}

	b, err := json.Marshal(sans)
	if err != nil {
		return ""
	}
	return string(b)
}

