package software

import (
	"fmt"
	"time"
)

// SoftwareData holds converted software data ready for Ent insertion.
type SoftwareData struct {
	ResourceID           string
	SoftwareName         string
	SoftwareVersion      string
	DisplayName          string
	ManufacturerID       int
	ManufacturerName     string
	SwCategoryName       string
	SwType               int
	SwFamily             int
	InstalledFormat      string
	IsUsageProhibited    int
	ManagedInstallations int
	NetworkInstallations int
	ManagedSwID          int
	DetectedTime         int64
	CompliantStatus      string
	TotalCopies          string
	RemainingCopies      string
	CollectedAt          time.Time
}

// cleanDash returns empty string if the value is the MEEC null marker "--".
func cleanDash(s string) string {
	if s == "--" {
		return ""
	}
	return s
}

// ConvertSoftware converts an API software entry to SoftwareData.
func ConvertSoftware(s APISoftware, collectedAt time.Time) *SoftwareData {
	return &SoftwareData{
		ResourceID:           fmt.Sprintf("%d", s.SoftwareID),
		SoftwareName:         s.SoftwareName,
		SoftwareVersion:      cleanDash(s.SoftwareVersion),
		DisplayName:          cleanDash(s.DisplayName),
		ManufacturerID:       s.ManufacturerID,
		ManufacturerName:     cleanDash(s.ManufacturerName),
		SwCategoryName:       cleanDash(s.SwCategoryName),
		SwType:               s.SwType,
		SwFamily:             s.SwFamily,
		InstalledFormat:      cleanDash(s.InstalledFormat),
		IsUsageProhibited:    s.IsUsageProhibited,
		ManagedInstallations: s.ManagedInstallations,
		NetworkInstallations: s.NetworkInstallations,
		ManagedSwID:          s.ManagedSwID,
		DetectedTime:         s.DetectedTime,
		CompliantStatus:      cleanDash(s.CompliantStatus),
		TotalCopies:          cleanDash(s.TotalCopies),
		RemainingCopies:      cleanDash(s.RemainingCopies),
		CollectedAt:          collectedAt,
	}
}
