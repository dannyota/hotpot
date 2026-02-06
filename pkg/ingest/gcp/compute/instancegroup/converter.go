package instancegroup

import (
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertInstanceGroup converts a GCP API InstanceGroup and its members to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertInstanceGroup(g *computepb.InstanceGroup, members []*computepb.InstanceWithNamedPorts, projectID string, collectedAt time.Time) bronze.GCPComputeInstanceGroup {
	group := bronze.GCPComputeInstanceGroup{
		ResourceID:        fmt.Sprintf("%d", g.GetId()),
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

// ConvertNamedPorts converts instance group named ports from GCP API to Bronze models.
func ConvertNamedPorts(ports []*computepb.NamedPort) []bronze.GCPComputeInstanceGroupNamedPort {
	if len(ports) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceGroupNamedPort, 0, len(ports))
	for _, p := range ports {
		result = append(result, bronze.GCPComputeInstanceGroupNamedPort{
			Name: p.GetName(),
			Port: p.GetPort(),
		})
	}

	return result
}

// ConvertMembers converts instance group members from GCP API to Bronze models.
func ConvertMembers(members []*computepb.InstanceWithNamedPorts) []bronze.GCPComputeInstanceGroupMember {
	if len(members) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceGroupMember, 0, len(members))
	for _, m := range members {
		instanceURL := m.GetInstance()
		result = append(result, bronze.GCPComputeInstanceGroupMember{
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
