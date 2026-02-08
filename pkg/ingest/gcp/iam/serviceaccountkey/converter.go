package serviceaccountkey

import (
	"path"
	"time"
)

// ServiceAccountKeyData holds converted service account key data ready for Ent insertion.
type ServiceAccountKeyData struct {
	ID                  string
	Name                string
	ServiceAccountEmail string
	KeyOrigin           string
	KeyType             string
	KeyAlgorithm        string
	ValidAfterTime      time.Time
	ValidBeforeTime     time.Time
	Disabled            bool
	ProjectID           string
	CollectedAt         time.Time
}

// ConvertServiceAccountKey converts a GCP API ServiceAccountKey to Ent-compatible data.
func ConvertServiceAccountKey(kwa KeyWithAccount, projectID string, collectedAt time.Time) *ServiceAccountKeyData {
	key := kwa.Key

	// Extract key ID from Name path (e.g., "projects/.../serviceAccounts/.../keys/abc123" -> "abc123")
	resourceID := path.Base(key.GetName())

	var validAfter, validBefore time.Time
	if key.GetValidAfterTime() != nil {
		validAfter = key.GetValidAfterTime().AsTime()
	}
	if key.GetValidBeforeTime() != nil {
		validBefore = key.GetValidBeforeTime().AsTime()
	}

	return &ServiceAccountKeyData{
		ID:                  resourceID,
		Name:                key.GetName(),
		ServiceAccountEmail: kwa.ServiceAccountEmail,
		KeyOrigin:           key.GetKeyOrigin().String(),
		KeyType:             key.GetKeyType().String(),
		KeyAlgorithm:        key.GetKeyAlgorithm().String(),
		ValidAfterTime:      validAfter,
		ValidBeforeTime:     validBefore,
		Disabled:            key.GetDisabled(),
		ProjectID:           projectID,
		CollectedAt:         collectedAt,
	}
}
