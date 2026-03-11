package packetmirroring

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// PacketMirroringData holds converted packet mirroring data ready for Ent insertion.
type PacketMirroringData struct {
	ID                    string
	Name                  string
	Description           string
	SelfLink              string
	Region                string
	Network               string
	Priority              int
	Enable                string
	CollectorIlbJSON      json.RawMessage
	MirroredResourcesJSON json.RawMessage
	FilterJSON            json.RawMessage
	CreationTimestamp      string
	ProjectID             string
	CollectedAt           time.Time
}

// ConvertPacketMirroring converts a GCP API PacketMirroring to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertPacketMirroring(pm *computepb.PacketMirroring, projectID string, collectedAt time.Time) (*PacketMirroringData, error) {
	data := &PacketMirroringData{
		ID:                fmt.Sprintf("%d", pm.GetId()),
		Name:              pm.GetName(),
		Description:       pm.GetDescription(),
		SelfLink:          pm.GetSelfLink(),
		Region:            pm.GetRegion(),
		Network:           pm.GetNetwork().GetUrl(),
		Priority:          int(pm.GetPriority()),
		Enable:            pm.GetEnable(),
		CreationTimestamp: pm.GetCreationTimestamp(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	// Marshal collector ILB
	if pm.GetCollectorIlb() != nil {
		collectorIlbBytes, err := json.Marshal(pm.GetCollectorIlb())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal collector ILB for %s: %w", pm.GetName(), err)
		}
		data.CollectorIlbJSON = collectorIlbBytes
	}

	// Marshal mirrored resources
	if pm.GetMirroredResources() != nil {
		mirroredResourcesBytes, err := json.Marshal(pm.GetMirroredResources())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal mirrored resources for %s: %w", pm.GetName(), err)
		}
		data.MirroredResourcesJSON = mirroredResourcesBytes
	}

	// Marshal filter
	if pm.GetFilter() != nil {
		filterBytes, err := json.Marshal(pm.GetFilter())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filter for %s: %w", pm.GetName(), err)
		}
		data.FilterJSON = filterBytes
	}

	return data, nil
}
