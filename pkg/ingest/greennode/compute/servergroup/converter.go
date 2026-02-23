package servergroup

import (
	"time"

	computev2 "danny.vn/greennode/services/compute/v2"
)

// ServerGroupData represents a converted server group ready for Ent insertion.
type ServerGroupData struct {
	ID          string
	Name        string
	Description string
	PolicyID    string
	PolicyName  string
	Region      string
	ProjectID   string
	CollectedAt time.Time

	Members []MemberData
}

// MemberData represents a server in a server group.
type MemberData struct {
	UUID string
	Name string
}

// ConvertServerGroup converts a GreenNode SDK ServerGroup to ServerGroupData.
func ConvertServerGroup(sg *computev2.ServerGroup, projectID, region string, collectedAt time.Time) *ServerGroupData {
	data := &ServerGroupData{
		ID:          sg.UUID,
		Name:        sg.Name,
		Description: sg.Description,
		PolicyID:    sg.PolicyID,
		PolicyName:  sg.PolicyName,
		Region:      region,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	data.Members = ConvertMembers(sg.Servers)

	return data
}

// ConvertMembers converts SDK server group members to MemberData.
func ConvertMembers(servers []computev2.ServerGroupMember) []MemberData {
	if len(servers) == 0 {
		return nil
	}
	result := make([]MemberData, 0, len(servers))
	for _, s := range servers {
		result = append(result, MemberData{
			UUID: s.UUID,
			Name: s.Name,
		})
	}
	return result
}
