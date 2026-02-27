package endpoint_app

import "time"

// EndpointAppData holds converted endpoint app data ready for Ent insertion.
type EndpointAppData struct {
	ResourceID    string
	AgentID       string
	Name          string
	Version       string
	Publisher     string
	Size          int
	InstalledDate *time.Time
	CollectedAt   time.Time
}

// ConvertEndpointApp converts an API endpoint app to EndpointAppData.
func ConvertEndpointApp(agentID string, app APIEndpointApp, collectedAt time.Time) *EndpointAppData {
	data := &EndpointAppData{
		ResourceID:  agentID + "||" + app.Name + "||" + app.Version,
		AgentID:     agentID,
		Name:        app.Name,
		Version:     app.Version,
		Publisher:   app.Publisher,
		Size:        app.Size,
		CollectedAt: collectedAt,
	}

	if app.InstalledDate != nil {
		if t, err := time.Parse("2006-01-02", *app.InstalledDate); err == nil {
			data.InstalledDate = &t
		}
	}

	return data
}
