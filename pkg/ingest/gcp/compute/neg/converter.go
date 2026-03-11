package neg

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

type NegData struct {
	ID                  string
	Name                string
	Description         string
	CreationTimestamp    string
	SelfLink            string
	Network             string
	Subnetwork          string
	Zone                string
	NetworkEndpointType string
	DefaultPort         string
	Size                string
	Region              string
	AnnotationsJSON     map[string]interface{}
	AppEngineJSON       map[string]interface{}
	CloudFunctionJSON   map[string]interface{}
	CloudRunJSON        map[string]interface{}
	PscDataJSON         map[string]interface{}
	ProjectID           string
	CollectedAt         time.Time
}

func ConvertNeg(neg *computepb.NetworkEndpointGroup, projectID string, collectedAt time.Time) (*NegData, error) {
	data := &NegData{
		ID:                  fmt.Sprintf("%d", neg.GetId()),
		Name:                neg.GetName(),
		Description:         neg.GetDescription(),
		CreationTimestamp:    neg.GetCreationTimestamp(),
		SelfLink:            neg.GetSelfLink(),
		Network:             neg.GetNetwork(),
		Subnetwork:          neg.GetSubnetwork(),
		Zone:                neg.GetZone(),
		NetworkEndpointType: neg.GetNetworkEndpointType(),
		DefaultPort:         fmt.Sprintf("%d", neg.GetDefaultPort()),
		Size:                fmt.Sprintf("%d", neg.GetSize()),
		Region:              neg.GetRegion(),
		ProjectID:           projectID,
		CollectedAt:         collectedAt,
	}

	// Convert JSON fields
	if neg.GetAnnotations() != nil {
		b, err := json.Marshal(neg.GetAnnotations())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal annotations for NEG %s: %w", neg.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.AnnotationsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}

	if neg.GetAppEngine() != nil {
		b, err := json.Marshal(neg.GetAppEngine())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal app engine for NEG %s: %w", neg.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.AppEngineJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal app engine: %w", err)
		}
	}

	if neg.GetCloudFunction() != nil {
		b, err := json.Marshal(neg.GetCloudFunction())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cloud function for NEG %s: %w", neg.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.CloudFunctionJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cloud function: %w", err)
		}
	}

	if neg.GetCloudRun() != nil {
		b, err := json.Marshal(neg.GetCloudRun())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cloud run for NEG %s: %w", neg.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.CloudRunJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cloud run: %w", err)
		}
	}

	if neg.GetPscData() != nil {
		b, err := json.Marshal(neg.GetPscData())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal psc data for NEG %s: %w", neg.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.PscDataJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal psc data: %w", err)
		}
	}

	return data, nil
}
