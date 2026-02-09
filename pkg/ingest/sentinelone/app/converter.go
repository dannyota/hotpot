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
		CollectedAt:           collectedAt,
	}

	if a.InstalledDate != nil {
		if t, err := time.Parse(time.RFC3339, *a.InstalledDate); err == nil {
			data.InstalledDate = &t
		}
	}

	return data
}
