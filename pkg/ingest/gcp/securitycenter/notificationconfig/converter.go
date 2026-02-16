package notificationconfig

import (
	"time"

	"cloud.google.com/go/securitycenter/apiv1/securitycenterpb"
)

// NotificationConfigData holds converted SCC notification config data ready for Ent insertion.
type NotificationConfigData struct {
	ID                  string
	Name                string
	Description         string
	PubsubTopic         string
	StreamingConfigJSON string
	ServiceAccount      string
	OrganizationID      string
	CollectedAt         time.Time
}

// ConvertNotificationConfig converts a raw GCP API SCC notification config to Ent-compatible data.
func ConvertNotificationConfig(orgName string, nc *securitycenterpb.NotificationConfig, collectedAt time.Time) *NotificationConfigData {
	return &NotificationConfigData{
		ID:                  nc.GetName(),
		Name:                nc.GetName(),
		Description:         nc.GetDescription(),
		PubsubTopic:         nc.GetPubsubTopic(),
		StreamingConfigJSON: streamingConfigToJSON(nc),
		ServiceAccount:      nc.GetServiceAccount(),
		OrganizationID:      orgName,
		CollectedAt:         collectedAt,
	}
}
