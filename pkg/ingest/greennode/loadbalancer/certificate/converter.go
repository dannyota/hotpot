package certificate

import (
	"time"

	lbv2 "danny.vn/greennode/services/loadbalancer/v2"
)

// CertificateData represents a converted certificate ready for Ent insertion.
type CertificateData struct {
	ID                 string
	Name               string
	CertificateType    string
	ExpiredAt          string
	ImportedAt         string
	NotAfter           int64
	KeyAlgorithm       string
	Serial             string
	Subject            string
	DomainName         string
	InUse              bool
	Issuer             string
	SignatureAlgorithm string
	NotBefore          int64
	Region             string
	ProjectID          string
	CollectedAt        time.Time
}

// ConvertCertificate converts a GreenNode SDK Certificate to CertificateData.
func ConvertCertificate(c *lbv2.Certificate, projectID, region string, collectedAt time.Time) *CertificateData {
	return &CertificateData{
		ID:                 c.UUID,
		Name:               c.Name,
		CertificateType:    c.CertificateType,
		ExpiredAt:          c.ExpiredAt,
		ImportedAt:         c.ImportedAt,
		NotAfter:           c.NotAfter,
		KeyAlgorithm:       c.KeyAlgorithm,
		Serial:             c.Serial,
		Subject:            c.Subject,
		DomainName:         c.DomainName,
		InUse:              c.InUse,
		Issuer:             c.Issuer,
		SignatureAlgorithm: c.SignatureAlgorithm,
		NotBefore:          c.NotBefore,
		Region:             region,
		ProjectID:          projectID,
		CollectedAt:        collectedAt,
	}
}
