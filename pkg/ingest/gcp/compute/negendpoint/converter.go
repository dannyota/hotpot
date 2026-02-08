package negendpoint

import (
	"encoding/json"
	"fmt"
	"time"
)

type NegEndpointData struct {
	ID              string
	Instance        string
	IpAddress       string
	Ipv6Address     string
	Port            string
	Fqdn            string
	AnnotationsJSON map[string]interface{}
	NegName         string
	Zone            string
	ProjectID       string
	CollectedAt     time.Time
}

func ConvertNegEndpoint(ewn EndpointWithNeg, projectID string, collectedAt time.Time) (*NegEndpointData, error) {
	ep := ewn.Endpoint.GetNetworkEndpoint()
	if ep == nil {
		return nil, fmt.Errorf("nil network endpoint in NEG %s", ewn.NegName)
	}

	// Synthetic resource ID: {neg_resource_id}/{ip_address}:{port}
	id := fmt.Sprintf("%s/%s:%d", ewn.NegResourceID, ep.GetIpAddress(), ep.GetPort())

	data := &NegEndpointData{
		ID:          id,
		Instance:    ep.GetInstance(),
		IpAddress:   ep.GetIpAddress(),
		Ipv6Address: ep.GetIpv6Address(),
		Port:        fmt.Sprintf("%d", ep.GetPort()),
		Fqdn:        ep.GetFqdn(),
		NegName:     ewn.NegName,
		Zone:        ewn.Zone,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if ep.GetAnnotations() != nil {
		b, err := json.Marshal(ep.GetAnnotations())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal annotations for endpoint in NEG %s: %w", ewn.NegName, err)
		}
		if err := json.Unmarshal(b, &data.AnnotationsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}

	return data, nil
}
