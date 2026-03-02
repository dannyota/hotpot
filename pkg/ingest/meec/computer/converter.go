package computer

import (
	"fmt"
	"time"
)

// ComputerData holds converted computer data ready for Ent insertion.
type ComputerData struct {
	ResourceID           string
	ResourceName         string
	FQDNName             string
	DomainNetbiosName    string
	IPAddress            string
	MACAddress           string
	OsName               string
	OsPlatform           int
	OsPlatformName       string
	OsVersion            string
	ServicePack          string
	AgentVersion         string
	ComputerLiveStatus   int
	InstallationStatus   int
	ManagedStatus        int
	BranchOfficeName     string
	Owner                string
	OwnerEmailID         string
	Description          string
	Location             string
	LastSyncTime         int64
	AgentLastContactTime int64
	AgentInstalledOn     int64
	CustomerName         string
	CustomerID           int
	CollectedAt          time.Time
}

// ConvertComputer converts an API computer to ComputerData.
func ConvertComputer(c APIComputer, collectedAt time.Time) *ComputerData {
	return &ComputerData{
		ResourceID:           fmt.Sprintf("%d", c.ResourceID),
		ResourceName:         c.ResourceName,
		FQDNName:             cleanString(c.FQDNName),
		DomainNetbiosName:    cleanString(c.DomainNetbiosName),
		IPAddress:            cleanString(c.IPAddress),
		MACAddress:           cleanString(c.MACAddress),
		OsName:               cleanString(c.OsName),
		OsPlatform:           cleanInt(c.OsPlatform),
		OsPlatformName:       cleanString(c.OsPlatformName),
		OsVersion:            cleanString(c.OsVersion),
		ServicePack:          cleanString(c.ServicePack),
		AgentVersion:         cleanString(c.AgentVersion),
		ComputerLiveStatus:   cleanInt(c.ComputerLiveStatus),
		InstallationStatus:   cleanInt(c.InstallationStatus),
		ManagedStatus:        cleanInt(c.ManagedStatus),
		BranchOfficeName:     cleanString(c.BranchOfficeName),
		Owner:                cleanString(c.Owner),
		OwnerEmailID:         cleanString(c.OwnerEmailID),
		Description:          cleanString(c.Description),
		Location:             cleanString(c.Location),
		LastSyncTime:         cleanInt64(c.LastSyncTime),
		AgentLastContactTime: cleanInt64(c.AgentLastContactTime),
		AgentInstalledOn:     cleanInt64(c.AgentInstalledOn),
		CustomerName:         cleanString(c.CustomerName),
		CustomerID:           cleanInt(c.CustomerID),
		CollectedAt:          collectedAt,
	}
}

// cleanString extracts a string from an any value.
// MEEC returns "--" for empty/unknown fields; we treat that as empty string.
func cleanString(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	if s == "--" {
		return ""
	}
	return s
}

// cleanInt extracts an int from an any value.
// MEEC can return "--" (string) or a float64 (from JSON number).
func cleanInt(v any) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		if val == "--" {
			return 0
		}
		return 0
	default:
		return 0
	}
}

// cleanInt64 extracts an int64 from an any value.
// MEEC can return "--" (string) or a float64 (from JSON number).
func cleanInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case string:
		if val == "--" {
			return 0
		}
		return 0
	default:
		return 0
	}
}
