package installed_software

import (
	"fmt"
	"time"
)

// InstalledSoftwareData holds converted installed software data ready for Ent insertion.
type InstalledSoftwareData struct {
	ResourceID         string
	ComputerResourceID string
	SoftwareID         int
	SoftwareName       string
	SoftwareVersion    string
	DisplayName        string
	ManufacturerName   string
	InstalledDate      int64
	Architecture       string
	Location           string
	SwType             int
	SwCategoryName     string
	DetectedTime       int64
	CollectedAt        time.Time
}

// cleanString extracts a string from an any value, treating "--" as empty.
func cleanString(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	if s == "--" {
		return ""
	}
	return s
}

// ConvertInstalledSoftware converts an API installed software to InstalledSoftwareData.
func ConvertInstalledSoftware(computerResourceID string, s APIInstalledSoftware, collectedAt time.Time) *InstalledSoftwareData {
	return &InstalledSoftwareData{
		ResourceID:         fmt.Sprintf("%s||%d", computerResourceID, s.SoftwareID),
		ComputerResourceID: computerResourceID,
		SoftwareID:         s.SoftwareID,
		SoftwareName:       s.SoftwareName,
		SoftwareVersion:    s.SoftwareVersion,
		DisplayName:        cleanString(s.DisplayName),
		ManufacturerName:   cleanString(s.ManufacturerName),
		InstalledDate:      s.InstalledDate,
		Architecture:       cleanString(s.Architecture),
		Location:           cleanString(s.Location),
		SwType:             s.SwType,
		SwCategoryName:     cleanString(s.SwCategoryName),
		DetectedTime:       s.DetectedTime,
		CollectedAt:        collectedAt,
	}
}
