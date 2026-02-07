package connector

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/vpcaccess/apiv1/vpcaccesspb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertConnector converts a GCP API Connector to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertConnector(c *vpcaccesspb.Connector, projectID string, collectedAt time.Time) (bronze.GCPVpcAccessConnector, error) {
	connector := bronze.GCPVpcAccessConnector{
		ResourceID:    c.GetName(),
		Network:       c.GetNetwork(),
		IpCidrRange:   c.GetIpCidrRange(),
		State:         c.GetState().String(),
		MinThroughput: c.GetMinThroughput(),
		MaxThroughput: c.GetMaxThroughput(),
		MinInstances:  c.GetMinInstances(),
		MaxInstances:  c.GetMaxInstances(),
		MachineType:   c.GetMachineType(),
		Region:        extractRegion(c.GetName()),
		ProjectID:     projectID,
		CollectedAt:   collectedAt,
	}

	// Convert JSONB fields (nil -> SQL NULL, data -> JSON bytes)
	var err error
	if sub := c.GetSubnet(); sub != nil {
		connector.SubnetJSON, err = json.Marshal(sub)
		if err != nil {
			return bronze.GCPVpcAccessConnector{}, fmt.Errorf("failed to marshal subnet JSON for connector %s: %w", c.GetName(), err)
		}
	}
	if projects := c.GetConnectedProjects(); len(projects) > 0 {
		connector.ConnectedProjectsJSON, err = json.Marshal(projects)
		if err != nil {
			return bronze.GCPVpcAccessConnector{}, fmt.Errorf("failed to marshal connected projects JSON for connector %s: %w", c.GetName(), err)
		}
	}

	return connector, nil
}

// extractRegion extracts the region from a connector resource name.
// Format: projects/{project}/locations/{region}/connectors/{name}
func extractRegion(name string) string {
	parts := strings.Split(name, "/")
	for i, part := range parts {
		if part == "locations" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
