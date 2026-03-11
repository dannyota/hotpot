package cryptokey

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/kms/apiv1/kmspb"
)

// CryptoKeyData holds converted crypto key data ready for Ent insertion.
type CryptoKeyData struct {
	ID                       string
	Name                     string
	Purpose                  string
	CreateTime               string
	NextRotationTime         string
	RotationPeriod           string
	DestroyScheduledDuration string
	ImportOnly               bool
	CryptoKeyBackend         string
	VersionTemplateJSON      json.RawMessage
	PrimaryJSON              json.RawMessage
	LabelsJSON               json.RawMessage
	ProjectID                string
	Location                 string
	KeyRingName              string
	CollectedAt              time.Time
}

// ConvertCryptoKey converts a KMS API CryptoKey to Ent-compatible data.
func ConvertCryptoKey(key *kmspb.CryptoKey, projectID string, collectedAt time.Time) (*CryptoKeyData, error) {
	data := &CryptoKeyData{
		ID:               key.GetName(),
		Name:             key.GetName(),
		Purpose:          key.GetPurpose().String(),
		ImportOnly:        key.GetImportOnly(),
		CryptoKeyBackend: key.GetCryptoKeyBackend(),
		ProjectID:        projectID,
		Location:         extractLocation(key.GetName()),
		KeyRingName:      extractKeyRingName(key.GetName()),
		CollectedAt:      collectedAt,
	}

	if key.GetCreateTime() != nil {
		data.CreateTime = key.GetCreateTime().AsTime().Format(time.RFC3339)
	}
	if key.GetNextRotationTime() != nil {
		data.NextRotationTime = key.GetNextRotationTime().AsTime().Format(time.RFC3339)
	}
	if key.GetRotationPeriod() != nil {
		data.RotationPeriod = key.GetRotationPeriod().AsDuration().String()
	}
	if key.GetDestroyScheduledDuration() != nil {
		data.DestroyScheduledDuration = key.GetDestroyScheduledDuration().AsDuration().String()
	}

	if key.GetVersionTemplate() != nil {
		j, err := json.Marshal(map[string]interface{}{
			"algorithm":       key.GetVersionTemplate().GetAlgorithm().String(),
			"protectionLevel": key.GetVersionTemplate().GetProtectionLevel().String(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal version_template for key %s: %w", key.GetName(), err)
		}
		data.VersionTemplateJSON = j
	}

	if key.GetPrimary() != nil {
		j, err := json.Marshal(key.GetPrimary())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal primary for key %s: %w", key.GetName(), err)
		}
		data.PrimaryJSON = j
	}

	if len(key.GetLabels()) > 0 {
		j, err := json.Marshal(key.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for key %s: %w", key.GetName(), err)
		}
		data.LabelsJSON = j
	}

	return data, nil
}

func extractLocation(name string) string {
	parts := strings.Split(name, "/")
	for i, p := range parts {
		if p == "locations" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func extractKeyRingName(name string) string {
	// name format: projects/{project}/locations/{location}/keyRings/{keyring}/cryptoKeys/{key}
	parts := strings.Split(name, "/")
	for i, p := range parts {
		if p == "keyRings" && i+1 < len(parts) {
			return strings.Join(parts[:i+2], "/")
		}
	}
	return ""
}
