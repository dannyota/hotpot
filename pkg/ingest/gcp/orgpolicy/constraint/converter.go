package constraint

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/orgpolicy/apiv2/orgpolicypb"
	"google.golang.org/protobuf/encoding/protojson"
)

type ConstraintData struct {
	ID                 string
	DisplayName        string
	Description        string
	ConstraintDefault  int
	SupportsDryRun     bool
	SupportsSimulation bool
	ListConstraint     map[string]any
	BooleanConstraint  map[string]any
	OrganizationID     string
	CollectedAt        time.Time
}

func ConvertConstraint(orgName string, c *orgpolicypb.Constraint, collectedAt time.Time) *ConstraintData {
	data := &ConstraintData{
		ID:                 c.GetName(),
		DisplayName:        c.GetDisplayName(),
		Description:        c.GetDescription(),
		ConstraintDefault:  int(c.GetConstraintDefault()),
		SupportsDryRun:     c.GetSupportsDryRun(),
		SupportsSimulation: c.GetSupportsSimulation(),
		OrganizationID:     orgName,
		CollectedAt:        collectedAt,
	}

	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	if lc := c.GetListConstraint(); lc != nil {
		if b, err := marshaler.Marshal(lc); err == nil {
			var m map[string]any
			if json.Unmarshal(b, &m) == nil {
				data.ListConstraint = m
			}
		}
	}

	if bc := c.GetBooleanConstraint(); bc != nil {
		if b, err := marshaler.Marshal(bc); err == nil {
			var m map[string]any
			if json.Unmarshal(b, &m) == nil {
				data.BooleanConstraint = m
			}
		}
	}

	return data
}
