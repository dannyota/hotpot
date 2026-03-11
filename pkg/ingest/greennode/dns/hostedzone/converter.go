package hostedzone

import (
	"encoding/json"
	"fmt"
	"time"

	dnsv1 "danny.vn/gnode/services/dns/v1"
)

// HostedZoneData represents a converted hosted zone ready for Ent insertion.
type HostedZoneData struct {
	ID                    string
	DomainName            string
	Status                string
	Description           string
	Type                  string
	CountRecords          int
	AssocVpcIdsJSON       json.RawMessage
	AssocVpcMapRegionJSON json.RawMessage
	PortalUserID          int
	CreatedAtAPI          string
	DeletedAtAPI          *string
	UpdatedAtAPI          string
	ProjectID             string
	CollectedAt           time.Time

	Records []RecordData
}

// RecordData represents a converted DNS record ready for Ent insertion.
type RecordData struct {
	RecordID            string
	SubDomain           string
	Status              string
	Type                string
	RoutingPolicy       string
	ValueJSON           json.RawMessage
	TTL                 int
	EnableStickySession *bool
	CreatedAtAPI        string
	DeletedAtAPI        *string
	UpdatedAtAPI        string
}

// ConvertHostedZone converts a GreenNode SDK HostedZone to HostedZoneData.
func ConvertHostedZone(hz *dnsv1.HostedZone, projectID string, collectedAt time.Time) (*HostedZoneData, error) {
	data := &HostedZoneData{
		ID:           hz.HostedZoneID,
		DomainName:   hz.DomainName,
		Status:       hz.Status,
		Description:  hz.Description,
		Type:         hz.Type,
		CountRecords: hz.CountRecords,
		PortalUserID: hz.PortalUserID,
		CreatedAtAPI: hz.CreatedAt.Format(time.RFC3339),
		UpdatedAtAPI: hz.UpdatedAt.Format(time.RFC3339),
		ProjectID:    projectID,
		CollectedAt:  collectedAt,
	}

	if hz.DeletedAt != nil {
		s := hz.DeletedAt.Format(time.RFC3339)
		data.DeletedAtAPI = &s
	}

	// Marshal AssocVpcIDs
	if len(hz.AssocVpcIDs) > 0 {
		vpcJSON, err := json.Marshal(hz.AssocVpcIDs)
		if err != nil {
			return nil, fmt.Errorf("marshal assoc vpc ids for zone %s: %w", hz.HostedZoneID, err)
		}
		data.AssocVpcIdsJSON = vpcJSON
	}

	// Marshal AssocVpcMapRegion
	if len(hz.AssocVpcMapRegion) > 0 {
		regionJSON, err := json.Marshal(hz.AssocVpcMapRegion)
		if err != nil {
			return nil, fmt.Errorf("marshal assoc vpc map region for zone %s: %w", hz.HostedZoneID, err)
		}
		data.AssocVpcMapRegionJSON = regionJSON
	}

	return data, nil
}

// ConvertRecords converts SDK DNS records to RecordData.
func ConvertRecords(records []*dnsv1.DnsRecord) ([]RecordData, error) {
	if len(records) == 0 {
		return nil, nil
	}
	result := make([]RecordData, 0, len(records))
	for _, r := range records {
		rd := RecordData{
			RecordID:            r.RecordID,
			SubDomain:           r.SubDomain,
			Status:              r.Status,
			Type:                r.Type,
			RoutingPolicy:       r.RoutingPolicy,
			TTL:                 r.TTL,
			EnableStickySession: r.EnableStickySession,
			CreatedAtAPI:        r.CreatedAt.Format(time.RFC3339),
			UpdatedAtAPI:        r.UpdatedAt.Format(time.RFC3339),
		}

		if r.DeletedAt != nil {
			s := r.DeletedAt.Format(time.RFC3339)
			rd.DeletedAtAPI = &s
		}

		if len(r.Value) > 0 {
			valJSON, err := json.Marshal(r.Value)
			if err != nil {
				return nil, fmt.Errorf("marshal values for record %s: %w", r.RecordID, err)
			}
			rd.ValueJSON = valJSON
		}

		result = append(result, rd)
	}
	return result, nil
}
