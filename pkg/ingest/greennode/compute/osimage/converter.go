package osimage

import (
	"time"

	computev1 "danny.vn/greennode/services/compute/v1"
)

// OSImageData represents a converted OS image ready for Ent insertion.
type OSImageData struct {
	ID                   string
	ImageType            string
	ImageVersion         string
	Licence              *bool
	LicenseKey           *string
	Description          string
	ZoneID               string
	FlavorZoneIDs        []string
	DefaultTagIDs        []string
	PackageLimitCpu      int64
	PackageLimitMemory   int64
	PackageLimitDiskSize int64
	Region               string
	ProjectID            string
	CollectedAt          time.Time
}

// ConvertOSImage converts a GreenNode SDK OSImage to OSImageData.
func ConvertOSImage(img *computev1.OSImage, projectID, region string, collectedAt time.Time) *OSImageData {
	return &OSImageData{
		ID:                   img.ID,
		ImageType:            img.ImageType,
		ImageVersion:         img.ImageVersion,
		Licence:              img.Licence,
		LicenseKey:           img.LicenseKey,
		Description:          img.Description,
		ZoneID:               img.ZoneID,
		FlavorZoneIDs:        img.FlavorZoneIDs,
		DefaultTagIDs:        img.DefaultTagIDs,
		PackageLimitCpu:      img.PackageLimit.Cpu,
		PackageLimitMemory:   img.PackageLimit.Memory,
		PackageLimitDiskSize: img.PackageLimit.DiskSize,
		Region:               region,
		ProjectID:            projectID,
		CollectedAt:          collectedAt,
	}
}
