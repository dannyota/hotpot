package certificate

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// CertificateDiff represents changes between old and new certificate states.
type CertificateDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffCertificateData compares old Ent entity and new CertificateData.
func DiffCertificateData(old *ent.BronzeGreenNodeLoadBalancerCertificate, new *CertificateData) *CertificateDiff {
	if old == nil {
		return &CertificateDiff{IsNew: true}
	}

	return &CertificateDiff{
		IsChanged: old.Name != new.Name ||
			old.CertificateType != new.CertificateType ||
			old.ExpiredAt != new.ExpiredAt ||
			old.ImportedAt != new.ImportedAt ||
			old.NotAfter != new.NotAfter ||
			old.KeyAlgorithm != new.KeyAlgorithm ||
			old.Serial != new.Serial ||
			old.Subject != new.Subject ||
			old.DomainName != new.DomainName ||
			old.InUse != new.InUse ||
			old.Issuer != new.Issuer ||
			old.SignatureAlgorithm != new.SignatureAlgorithm ||
			old.NotBefore != new.NotBefore,
	}
}

// HasAnyChange returns true if the certificate changed.
func (d *CertificateDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
