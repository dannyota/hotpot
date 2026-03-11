package secgroup

import (
	"time"

	networkv2 "danny.vn/gnode/services/network/v2"
)

// SecgroupData represents a converted security group ready for Ent insertion.
type SecgroupData struct {
	ID          string
	Name        string
	Description string
	Status      string
	CreatedAt   string
	IsSystem    bool
	Region      string
	ProjectID   string
	CollectedAt time.Time

	Rules []SecgroupRuleData
}

// SecgroupRuleData represents a security group rule ready for Ent insertion.
type SecgroupRuleData struct {
	RuleID         string
	Direction      string
	EtherType      string
	Protocol       string
	Description    string
	RemoteIPPrefix string
	PortRangeMax   int
	PortRangeMin   int
}

// ConvertSecgroup converts a GreenNode SDK Secgroup and its rules to SecgroupData.
func ConvertSecgroup(sg *networkv2.Secgroup, rules []*networkv2.SecgroupRule, projectID, region string, collectedAt time.Time) *SecgroupData {
	data := &SecgroupData{
		ID:          sg.ID,
		Name:        sg.Name,
		Description: sg.Description,
		Status:      sg.Status,
		CreatedAt:   sg.CreatedAt,
		IsSystem:    sg.IsSystem,
		Region:      region,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	data.Rules = ConvertSecgroupRules(rules)

	return data
}

// ConvertSecgroupRules converts SDK security group rules to SecgroupRuleData.
func ConvertSecgroupRules(rules []*networkv2.SecgroupRule) []SecgroupRuleData {
	if len(rules) == 0 {
		return nil
	}
	result := make([]SecgroupRuleData, 0, len(rules))
	for _, r := range rules {
		result = append(result, SecgroupRuleData{
			RuleID:         r.ID,
			Direction:      r.Direction,
			EtherType:      r.EtherType,
			Protocol:       r.Protocol,
			Description:    r.Description,
			RemoteIPPrefix: r.RemoteIPPrefix,
			PortRangeMax:   r.PortRangeMax,
			PortRangeMin:   r.PortRangeMin,
		})
	}
	return result
}
