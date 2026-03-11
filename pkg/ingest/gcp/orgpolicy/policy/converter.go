package policy

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/orgpolicy/apiv2/orgpolicypb"
	"google.golang.org/protobuf/encoding/protojson"
)

type PolicyData struct {
	ID             string
	Etag           string
	Spec           map[string]any
	DryRunSpec     map[string]any
	OrganizationID string
	CollectedAt    time.Time
}

func ConvertPolicy(orgName string, p *orgpolicypb.Policy, collectedAt time.Time) *PolicyData {
	data := &PolicyData{
		ID:             p.GetName(),
		Etag:           p.GetEtag(),
		OrganizationID: orgName,
		CollectedAt:    collectedAt,
	}

	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	if spec := p.GetSpec(); spec != nil {
		if b, err := marshaler.Marshal(spec); err == nil {
			var m map[string]any
			if json.Unmarshal(b, &m) == nil {
				data.Spec = m
			}
		}
	}

	if dryRunSpec := p.GetDryRunSpec(); dryRunSpec != nil {
		if b, err := marshaler.Marshal(dryRunSpec); err == nil {
			var m map[string]any
			if json.Unmarshal(b, &m) == nil {
				data.DryRunSpec = m
			}
		}
	}

	return data
}
