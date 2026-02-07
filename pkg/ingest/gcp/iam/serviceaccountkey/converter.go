package serviceaccountkey

import (
	"path"
	"time"

	"hotpot/pkg/base/models/bronze"
)

func ConvertServiceAccountKey(kwa KeyWithAccount, projectID string, collectedAt time.Time) bronze.GCPIAMServiceAccountKey {
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

	return bronze.GCPIAMServiceAccountKey{
		ResourceID:          resourceID,
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
