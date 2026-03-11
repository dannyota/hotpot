package blockvolume

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	volumev2 "danny.vn/gnode/services/volume/v2"
)

// BlockVolumeData represents a converted block volume ready for Ent insertion.
type BlockVolumeData struct {
	ID                  string
	Name                string
	VolumeTypeID        string
	ClusterID           string
	VMID                string
	Size                string
	IopsID              string
	Status              string
	CreatedAtAPI        string
	UpdatedAtAPI        string
	PersistentVolume    bool
	AttachedMachineJSON json.RawMessage
	UnderID             string
	MigrateState        string
	MultiAttach         bool
	ZoneID              string
	Region              string
	ProjectID           string
	CollectedAt         time.Time

	Snapshots []SnapshotData
}

// SnapshotData represents a converted snapshot ready for Ent insertion.
type SnapshotData struct {
	SnapshotID   string
	Name         string
	Size         int64
	VolumeSize   int64
	Status       string
	CreatedAtAPI string
}

// ConvertBlockVolume converts a GreenNode SDK Volume to BlockVolumeData.
func ConvertBlockVolume(v *volumev2.Volume, projectID, region string, collectedAt time.Time) (*BlockVolumeData, error) {
	data := &BlockVolumeData{
		ID:               v.ID,
		Name:             v.Name,
		VolumeTypeID:     v.VolumeTypeID,
		VMID:             v.VmID,
		Size:             strconv.FormatUint(v.Size, 10),
		IopsID:           strconv.FormatUint(v.IopsID, 10),
		Status:           v.Status,
		CreatedAtAPI:     v.CreatedAt,
		PersistentVolume: v.PersistentVolume,
		UnderID:          v.UnderID,
		MigrateState:     v.MigrateState,
		MultiAttach:      v.MultiAttach,
		ZoneID:           v.ZoneID,
		Region:           region,
		ProjectID:        projectID,
		CollectedAt:      collectedAt,
	}

	if v.ClusterID != nil {
		data.ClusterID = *v.ClusterID
	}

	if v.UpdatedAt != nil {
		data.UpdatedAtAPI = *v.UpdatedAt
	}

	if len(v.AttachedMachine) > 0 {
		machineJSON, err := json.Marshal(v.AttachedMachine)
		if err != nil {
			return nil, fmt.Errorf("marshal attached machines for volume %s: %w", v.ID, err)
		}
		data.AttachedMachineJSON = machineJSON
	}

	return data, nil
}

// ConvertSnapshots converts SDK snapshots to SnapshotData.
func ConvertSnapshots(snapshots []*volumev2.Snapshot) []SnapshotData {
	if len(snapshots) == 0 {
		return nil
	}
	result := make([]SnapshotData, 0, len(snapshots))
	for _, s := range snapshots {
		result = append(result, SnapshotData{
			SnapshotID:   s.ID,
			Name:         s.Name,
			Size:         s.Size,
			VolumeSize:   s.VolumeSize,
			Status:       s.Status,
			CreatedAtAPI: s.CreatedAt,
		})
	}
	return result
}
