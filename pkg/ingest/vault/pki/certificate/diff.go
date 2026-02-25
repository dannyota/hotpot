package certificate

import (
	"time"

	entpki "github.com/dannyota/hotpot/pkg/storage/ent/vault/pki"
)

// CertDiff represents changes between old and new certificate states.
type CertDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffCertData compares old Ent entity and new data.
func DiffCertData(old *entpki.BronzeVaultPKICertificate, new *CertificateData) *CertDiff {
	if old == nil {
		return &CertDiff{IsNew: true}
	}

	diff := &CertDiff{}

	diff.IsChanged = old.CommonName != new.CommonName ||
		old.IssuerCn != new.IssuerCN ||
		old.SubjectCn != new.SubjectCN ||
		old.Sans != new.SANs ||
		old.KeyType != new.KeyType ||
		old.KeyBits != new.KeyBits ||
		old.SigningAlgo != new.SigningAlgo ||
		!old.NotBefore.Equal(new.NotBefore) ||
		!old.NotAfter.Equal(new.NotAfter) ||
		old.IsRevoked != new.IsRevoked ||
		!revokedAtEqual(old.RevokedAt, new.RevokedAt) ||
		old.CertificatePem != new.CertPEM

	return diff
}

func revokedAtEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
