package app

import "time"

// AppData holds converted app data ready for Ent insertion.
type AppData struct {
	ResourceID            string
	Name                  string
	Publisher             string
	Version               string
	Size                  int64
	AppType               string
	OsType                string
	InstalledDate         *time.Time
	AgentID               string
	AgentComputerName     string
	AgentMachineType      string
	AgentIsActive         bool
	AgentIsDecommissioned bool
	RiskLevel             string
	Signed                bool
	APICreatedAt          *time.Time
	APIUpdatedAt          *time.Time
	AgentUUID             string
	AgentDomain           string
	AgentVersion          string
	AgentOsType           string
	AgentNetworkStatus    string
	AgentInfected         bool
	AgentOperationalState string
	CollectedAt           time.Time
}

// ConvertApp converts an API app to AppData.
func ConvertApp(a APIApp, collectedAt time.Time) *AppData {
	data := &AppData{
		ResourceID:            a.ID,
		Name:                  a.Name,
		Publisher:             a.Publisher,
		Version:               a.Version,
		Size:                  a.Size,
		AppType:               a.Type,
		OsType:                a.OsType,
		AgentID:               a.AgentID,
		AgentComputerName:     a.AgentComputerName,
		AgentMachineType:      a.AgentMachineType,
		AgentIsActive:         a.AgentIsActive,
		AgentIsDecommissioned: a.AgentIsDecommissioned,
		RiskLevel:             a.RiskLevel,
		Signed:                a.Signed,
		AgentUUID:             a.AgentUUID,
		AgentDomain:           a.AgentDomain,
		AgentVersion:          a.AgentVersion,
		AgentOsType:           a.AgentOsType,
		AgentNetworkStatus:    a.AgentNetworkStatus,
		AgentInfected:         a.AgentInfected,
		AgentOperationalState: a.AgentOperationalState,
		CollectedAt:           collectedAt,
	}

	if a.InstalledDate != nil {
		if t, err := time.Parse(time.RFC3339, *a.InstalledDate); err == nil {
			data.InstalledDate = &t
		}
	}
	if a.CreatedAt != nil {
		if t, err := time.Parse(time.RFC3339, *a.CreatedAt); err == nil {
			data.APICreatedAt = &t
		}
	}
	if a.UpdatedAt != nil {
		if t, err := time.Parse(time.RFC3339, *a.UpdatedAt); err == nil {
			data.APIUpdatedAt = &t
		}
	}

	return data
}
