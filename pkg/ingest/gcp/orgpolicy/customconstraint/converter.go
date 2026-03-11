package customconstraint

import (
	"time"

	"cloud.google.com/go/orgpolicy/apiv2/orgpolicypb"
)

type CustomConstraintData struct {
	ID             string
	ResourceTypes  []string
	MethodTypes    []int
	Condition      string
	ActionType     int
	DisplayName    string
	Description    string
	UpdateTime     *time.Time
	OrganizationID string
	CollectedAt    time.Time
}

func ConvertCustomConstraint(orgName string, cc *orgpolicypb.CustomConstraint, collectedAt time.Time) *CustomConstraintData {
	data := &CustomConstraintData{
		ID:             cc.GetName(),
		Condition:      cc.GetCondition(),
		ActionType:     int(cc.GetActionType()),
		DisplayName:    cc.GetDisplayName(),
		Description:    cc.GetDescription(),
		OrganizationID: orgName,
		CollectedAt:    collectedAt,
	}

	if rt := cc.GetResourceTypes(); len(rt) > 0 {
		data.ResourceTypes = rt
	}

	if mt := cc.GetMethodTypes(); len(mt) > 0 {
		methodTypes := make([]int, len(mt))
		for i, m := range mt {
			methodTypes[i] = int(m)
		}
		data.MethodTypes = methodTypes
	}

	if ut := cc.GetUpdateTime(); ut != nil {
		t := ut.AsTime()
		data.UpdateTime = &t
	}

	return data
}
