package key

import (
	"strconv"
	"time"

	"github.com/digitalocean/godo"
)

// KeyData holds converted SSH key data ready for Ent insertion.
type KeyData struct {
	ResourceID  string
	Name        string
	Fingerprint string
	PublicKey   string
	CollectedAt time.Time
}

// ConvertKey converts a godo Key to KeyData.
func ConvertKey(v godo.Key, collectedAt time.Time) *KeyData {
	return &KeyData{
		ResourceID:  strconv.Itoa(v.ID),
		Name:        v.Name,
		Fingerprint: v.Fingerprint,
		PublicKey:   v.PublicKey,
		CollectedAt: collectedAt,
	}
}
