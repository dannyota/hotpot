package userimage

import (
	"time"

	computev2 "danny.vn/gnode/services/compute/v2"
)

// UserImageData represents a converted user image ready for Ent insertion.
type UserImageData struct {
	ID          string
	Name        string
	Status      string
	MinDisk     int
	ImageSize   float64
	MetaData    string
	CreatedAt   string
	Region      string
	ProjectID   string
	CollectedAt time.Time
}

// ConvertUserImage converts a GreenNode SDK UserImage to UserImageData.
func ConvertUserImage(img *computev2.UserImage, projectID, region string, collectedAt time.Time) *UserImageData {
	return &UserImageData{
		ID:          img.Uuid,
		Name:        img.Name,
		Status:      img.Status,
		MinDisk:     img.MinDisk,
		ImageSize:   img.ImageSize,
		MetaData:    img.MetaData,
		CreatedAt:   img.CreatedAt,
		Region:      region,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}
}
