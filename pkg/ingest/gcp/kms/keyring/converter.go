package keyring

import (
	"strings"
	"time"
)

// KeyRingData holds converted key ring data ready for Ent insertion.
type KeyRingData struct {
	ID         string
	Name       string
	CreateTime string
	ProjectID  string
	Location   string
	CollectedAt time.Time
}

// ConvertKeyRing converts a KMS API KeyRing to Ent-compatible data.
func ConvertKeyRing(name string, createTime string, projectID string, collectedAt time.Time) *KeyRingData {
	// Extract location from resource name: projects/{project}/locations/{location}/keyRings/{keyring}
	location := extractLocation(name)

	return &KeyRingData{
		ID:          name,
		Name:        name,
		CreateTime:  createTime,
		ProjectID:   projectID,
		Location:    location,
		CollectedAt: collectedAt,
	}
}

// extractLocation extracts the location from a KMS resource name.
func extractLocation(name string) string {
	parts := strings.Split(name, "/")
	for i, p := range parts {
		if p == "locations" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
