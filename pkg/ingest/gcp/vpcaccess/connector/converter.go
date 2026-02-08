package connector

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/vpcaccess/apiv1/vpcaccesspb"
)

// ConnectorData holds converted VPC Access connector data ready for Ent insertion.
type ConnectorData struct {
	ID                     string
	Network                string
	IpCidrRange            string
	State                  string
	MinThroughput          int32
	MaxThroughput          int32
	MinInstances           int32
	MaxInstances           int32
	MachineType            string
	Region                 string
	SubnetJSON             json.RawMessage
	ConnectedProjectsJSON  json.RawMessage
	ProjectID              string
	CollectedAt            time.Time
}

// ConvertConnector converts a GCP API Connector to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertConnector(c *vpcaccesspb.Connector, projectID string, collectedAt time.Time) (*ConnectorData, error) {
	data := &ConnectorData{
		ID:            c.GetName(),
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

	// Convert JSONB fields
	var err error
	if sub := c.GetSubnet(); sub != nil {
		data.SubnetJSON, err = json.Marshal(sub)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal subnet JSON for connector %s: %w", c.GetName(), err)
		}
	}
	if projects := c.GetConnectedProjects(); len(projects) > 0 {
		data.ConnectedProjectsJSON, err = json.Marshal(projects)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal connected projects JSON for connector %s: %w", c.GetName(), err)
		}
	}

	return data, nil
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
