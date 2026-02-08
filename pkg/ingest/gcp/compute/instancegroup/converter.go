package instancegroup

import (
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// InstanceGroupData holds converted instance group data ready for Ent insertion.
type InstanceGroupData struct {
	ID                string
	Name              string
	Description       string
	Zone              string
	Network           string
	Subnetwork        string
	Size              int32
	SelfLink          string
	CreationTimestamp string
	Fingerprint       string
	NamedPorts        []NamedPortData
	Members           []MemberData
	ProjectID         string
	CollectedAt       time.Time
}

// NamedPortData holds converted named port data.
type NamedPortData struct {
	Name string
	Port int32
}

// MemberData holds converted member data.
type MemberData struct {
	InstanceURL  string
	InstanceName string
	Status       string
}

// ConvertInstanceGroup converts a GCP API InstanceGroup and its members to InstanceGroupData.
// Preserves raw API data with minimal transformation.
func ConvertInstanceGroup(g *computepb.InstanceGroup, members []*computepb.InstanceWithNamedPorts, projectID string, collectedAt time.Time) *InstanceGroupData {
	group := &InstanceGroupData{
		ID:                fmt.Sprintf("%d", g.GetId()),
		Name:              g.GetName(),
		Description:       g.GetDescription(),
		Zone:              g.GetZone(),
		Network:           g.GetNetwork(),
		Subnetwork:        g.GetSubnetwork(),
		Size:              g.GetSize(),
		SelfLink:          g.GetSelfLink(),
		CreationTimestamp: g.GetCreationTimestamp(),
		Fingerprint:       g.GetFingerprint(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	// Convert named ports to separate table
	group.NamedPorts = ConvertNamedPorts(g.NamedPorts)

	// Convert members to separate table
	group.Members = ConvertMembers(members)

	return group
}

// ConvertNamedPorts converts instance group named ports from GCP API to named port data.
func ConvertNamedPorts(ports []*computepb.NamedPort) []NamedPortData {
	if len(ports) == 0 {
		return nil
	}

	result := make([]NamedPortData, 0, len(ports))
	for _, p := range ports {
		result = append(result, NamedPortData{
			Name: p.GetName(),
			Port: p.GetPort(),
		})
	}

	return result
}

// ConvertMembers converts instance group members from GCP API to member data.
func ConvertMembers(members []*computepb.InstanceWithNamedPorts) []MemberData {
	if len(members) == 0 {
		return nil
	}

	result := make([]MemberData, 0, len(members))
	for _, m := range members {
		instanceURL := m.GetInstance()
		result = append(result, MemberData{
			InstanceURL:  instanceURL,
			InstanceName: extractInstanceName(instanceURL),
			Status:       m.GetStatus(),
		})
	}

	return result
}

// extractInstanceName extracts the instance name from a full instance URL.
// Example: ".../instances/my-vm" -> "my-vm"
func extractInstanceName(instanceURL string) string {
	parts := strings.Split(instanceURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
