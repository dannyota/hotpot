package settings

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/iap/apiv1/iappb"
)

// SettingsData holds converted IAP settings data ready for Ent insertion.
type SettingsData struct {
	ID                      string
	Name                    string
	AccessSettingsJSON      json.RawMessage
	ApplicationSettingsJSON json.RawMessage
	ProjectID               string
	CollectedAt             time.Time
}

// ConvertSettings converts a raw GCP API IAP settings to Ent-compatible data.
func ConvertSettings(settings *iappb.IapSettings, projectID string, collectedAt time.Time) (*SettingsData, error) {
	data := &SettingsData{
		ID:          settings.GetName(),
		Name:        settings.GetName(),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if settings.GetAccessSettings() != nil {
		accessJSON, err := json.Marshal(settings.GetAccessSettings())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal access_settings for %s: %w", settings.GetName(), err)
		}
		data.AccessSettingsJSON = accessJSON
	}

	if settings.GetApplicationSettings() != nil {
		appJSON, err := json.Marshal(settings.GetApplicationSettings())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal application_settings for %s: %w", settings.GetName(), err)
		}
		data.ApplicationSettingsJSON = appJSON
	}

	return data, nil
}
