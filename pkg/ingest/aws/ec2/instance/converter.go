package instance

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// InstanceData holds converted instance data ready for Ent insertion.
type InstanceData struct {
	ResourceID        string
	Name              string
	InstanceType      string
	State             string
	VpcID             string
	SubnetID          string
	PrivateIPAddress  string
	PublicIPAddress   string
	AmiID             string
	KeyName           string
	LaunchTime        *time.Time
	Platform          string
	Architecture      string
	SecurityGroupJSON json.RawMessage
	AccountID         string
	Region            string
	CollectedAt       time.Time

	// Child data
	Tags []TagData
}

// TagData holds tag key-value pairs.
type TagData struct {
	Key   string
	Value string
}

// SecurityGroupRef is the JSON structure for security group references.
type SecurityGroupRef struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

// ConvertInstance converts an AWS API Instance to InstanceData.
func ConvertInstance(inst types.Instance, accountID, region string, collectedAt time.Time) (*InstanceData, error) {
	data := &InstanceData{
		ResourceID:   derefStr(inst.InstanceId),
		InstanceType: string(inst.InstanceType),
		VpcID:        derefStr(inst.VpcId),
		SubnetID:     derefStr(inst.SubnetId),
		PrivateIPAddress: derefStr(inst.PrivateIpAddress),
		PublicIPAddress:  derefStr(inst.PublicIpAddress),
		AmiID:        derefStr(inst.ImageId),
		KeyName:      derefStr(inst.KeyName),
		LaunchTime:   inst.LaunchTime,
		Platform:     derefStr(inst.PlatformDetails),
		Architecture: string(inst.Architecture),
		AccountID:    accountID,
		Region:       region,
		CollectedAt:  collectedAt,
	}

	// Instance state
	if inst.State != nil {
		data.State = string(inst.State.Name)
	}

	// Extract Name from tags and collect all tags
	data.Tags = ConvertTags(inst.Tags)
	for _, tag := range data.Tags {
		if tag.Key == "Name" {
			data.Name = tag.Value
			break
		}
	}

	// Convert security groups to JSON
	if len(inst.SecurityGroups) > 0 {
		sgRefs := make([]SecurityGroupRef, 0, len(inst.SecurityGroups))
		for _, sg := range inst.SecurityGroups {
			sgRefs = append(sgRefs, SecurityGroupRef{
				GroupID:   derefStr(sg.GroupId),
				GroupName: derefStr(sg.GroupName),
			})
		}
		sgJSON, err := json.Marshal(sgRefs)
		if err != nil {
			return nil, err
		}
		data.SecurityGroupJSON = sgJSON
	}

	return data, nil
}

// ConvertTags converts AWS tags to TagData.
func ConvertTags(tags []types.Tag) []TagData {
	if len(tags) == 0 {
		return nil
	}

	result := make([]TagData, 0, len(tags))
	for _, tag := range tags {
		result = append(result, TagData{
			Key:   derefStr(tag.Key),
			Value: derefStr(tag.Value),
		})
	}
	return result
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
