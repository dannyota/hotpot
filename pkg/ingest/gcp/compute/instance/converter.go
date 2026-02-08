package instance

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// InstanceData holds converted instance data ready for Ent insertion.
type InstanceData struct {
	ResourceID             string
	Name                   string
	Zone                   string
	MachineType            string
	Status                 string
	StatusMessage          string
	CpuPlatform            string
	Hostname               string
	Description            string
	CreationTimestamp      string
	LastStartTimestamp     string
	LastStopTimestamp      string
	LastSuspendedTimestamp string
	DeletionProtection     bool
	CanIpForward           bool
	SelfLink               string
	SchedulingJSON         json.RawMessage
	ProjectID              string
	CollectedAt            time.Time

	// Child data
	Disks           []DiskData
	NICs            []NICData
	Labels          []LabelData
	Tags            []TagData
	Metadata        []MetadataData
	ServiceAccounts []ServiceAccountData
}

type DiskData struct {
	Source                string
	DeviceName            string
	Index                 int
	Boot                  bool
	AutoDelete            bool
	Mode                  string
	Interface             string
	Type                  string
	DiskSizeGb            int64
	DiskEncryptionKeyJSON json.RawMessage
	InitializeParamsJSON  json.RawMessage
	Licenses              []DiskLicenseData
}

type DiskLicenseData struct {
	License string
}

type NICData struct {
	Name            string
	Network         string
	Subnetwork      string
	NetworkIP       string
	StackType       string
	NicType         string
	AccessConfigs   []AccessConfigData
	AliasIPRanges   []AliasRangeData
}

type AccessConfigData struct {
	Type        string
	Name        string
	NatIP       string
	NetworkTier string
}

type AliasRangeData struct {
	IPCidrRange         string
	SubnetworkRangeName string
}

type LabelData struct {
	Key   string
	Value string
}

type TagData struct {
	Tag string
}

type MetadataData struct {
	Key   string
	Value string
}

type ServiceAccountData struct {
	Email      string
	ScopesJSON json.RawMessage
}

// ConvertInstance converts a GCP API Instance to InstanceData.
// Preserves raw API data with minimal transformation.
func ConvertInstance(inst *computepb.Instance, projectID string, collectedAt time.Time) (*InstanceData, error) {
	instance := &InstanceData{
		ResourceID:             fmt.Sprintf("%d", inst.GetId()),
		Name:                   inst.GetName(),
		Zone:                   inst.GetZone(),
		MachineType:            inst.GetMachineType(),
		Status:                 inst.GetStatus(),
		StatusMessage:          inst.GetStatusMessage(),
		CpuPlatform:            inst.GetCpuPlatform(),
		Hostname:               inst.GetHostname(),
		Description:            inst.GetDescription(),
		CreationTimestamp:      inst.GetCreationTimestamp(),
		LastStartTimestamp:     inst.GetLastStartTimestamp(),
		LastStopTimestamp:      inst.GetLastStopTimestamp(),
		LastSuspendedTimestamp: inst.GetLastSuspendedTimestamp(),
		DeletionProtection:     inst.GetDeletionProtection(),
		CanIpForward:           inst.GetCanIpForward(),
		SelfLink:               inst.GetSelfLink(),
		ProjectID:              projectID,
		CollectedAt:            collectedAt,
	}

	// Convert scheduling to JSONB (nil → SQL NULL, data → JSON bytes)
	if inst.Scheduling != nil {
		var err error
		instance.SchedulingJSON, err = json.Marshal(inst.Scheduling)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal scheduling for instance %s: %w", inst.GetName(), err)
		}
	}

	// Convert related entities
	var convertErr error
	instance.Disks, convertErr = ConvertDisks(inst.Disks)
	if convertErr != nil {
		return nil, fmt.Errorf("failed to convert disks for instance %s: %w", inst.GetName(), convertErr)
	}
	instance.NICs = ConvertNICs(inst.NetworkInterfaces)
	instance.Labels = ConvertLabels(inst.Labels)
	instance.Tags = ConvertTags(inst.Tags)
	instance.Metadata = ConvertMetadata(inst.Metadata)
	instance.ServiceAccounts, convertErr = ConvertServiceAccounts(inst.ServiceAccounts)
	if convertErr != nil {
		return nil, fmt.Errorf("failed to convert service accounts for instance %s: %w", inst.GetName(), convertErr)
	}

	return instance, nil
}

// ConvertDisks converts attached disk info from GCP API to DiskData.
func ConvertDisks(disks []*computepb.AttachedDisk) ([]DiskData, error) {
	if len(disks) == 0 {
		return nil, nil
	}

	var err error
	result := make([]DiskData, 0, len(disks))
	for _, disk := range disks {
		d := DiskData{
			Source:     disk.GetSource(),
			DeviceName: disk.GetDeviceName(),
			Index:      int(disk.GetIndex()),
			Boot:       disk.GetBoot(),
			AutoDelete: disk.GetAutoDelete(),
			Mode:       disk.GetMode(),
			Interface:  disk.GetInterface(),
			Type:       disk.GetType(),
			DiskSizeGb: disk.GetDiskSizeGb(),
		}

		// Convert encryption key to JSONB (nil → SQL NULL, data → JSON bytes)
		if disk.DiskEncryptionKey != nil {
			d.DiskEncryptionKeyJSON, err = json.Marshal(disk.DiskEncryptionKey)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal disk encryption key for %s: %w", disk.GetDeviceName(), err)
			}
		}

		// Convert initialize params to JSONB (nil → SQL NULL, data → JSON bytes)
		if disk.InitializeParams != nil {
			d.InitializeParamsJSON, err = json.Marshal(disk.InitializeParams)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal initialize params for %s: %w", disk.GetDeviceName(), err)
			}
		}

		// Convert licenses to separate table
		d.Licenses = ConvertDiskLicenses(disk.Licenses)

		result = append(result, d)
	}

	return result, nil
}

// ConvertNICs converts network interfaces from GCP API to NICData.
func ConvertNICs(nics []*computepb.NetworkInterface) []NICData {
	if len(nics) == 0 {
		return nil
	}

	result := make([]NICData, 0, len(nics))
	for _, nic := range nics {
		n := NICData{
			Name:       nic.GetName(),
			Network:    nic.GetNetwork(),
			Subnetwork: nic.GetSubnetwork(),
			NetworkIP:  nic.GetNetworkIP(),
			StackType:  nic.GetStackType(),
			NicType:    nic.GetNicType(),
		}

		// Convert access configs to separate table
		n.AccessConfigs = ConvertNICAccessConfigs(nic.AccessConfigs)

		// Convert alias IP ranges to separate table
		n.AliasIPRanges = ConvertNICAliasRanges(nic.AliasIpRanges)

		result = append(result, n)
	}

	return result
}

// ConvertLabels converts instance labels from GCP API to LabelData.
func ConvertLabels(labels map[string]string) []LabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]LabelData, 0, len(labels))
	for key, value := range labels {
		result = append(result, LabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// ConvertTags converts network tags from GCP API to TagData.
func ConvertTags(tags *computepb.Tags) []TagData {
	if tags == nil || len(tags.Items) == 0 {
		return nil
	}

	result := make([]TagData, 0, len(tags.Items))
	for _, tag := range tags.Items {
		result = append(result, TagData{
			Tag: tag,
		})
	}

	return result
}

// ConvertMetadata converts instance metadata from GCP API to MetadataData.
func ConvertMetadata(metadata *computepb.Metadata) []MetadataData {
	if metadata == nil || len(metadata.Items) == 0 {
		return nil
	}

	result := make([]MetadataData, 0, len(metadata.Items))
	for _, item := range metadata.Items {
		m := MetadataData{
			Key: item.GetKey(),
		}
		if item.Value != nil {
			m.Value = *item.Value
		}
		result = append(result, m)
	}

	return result
}

// ConvertServiceAccounts converts service accounts from GCP API to ServiceAccountData.
func ConvertServiceAccounts(serviceAccounts []*computepb.ServiceAccount) ([]ServiceAccountData, error) {
	if len(serviceAccounts) == 0 {
		return nil, nil
	}

	var err error
	result := make([]ServiceAccountData, 0, len(serviceAccounts))
	for _, sa := range serviceAccounts {
		s := ServiceAccountData{
			Email: sa.GetEmail(),
		}

		// Convert scopes to JSONB (nil → SQL NULL, data → JSON bytes)
		if sa.Scopes != nil {
			s.ScopesJSON, err = json.Marshal(sa.Scopes)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal scopes for %s: %w", sa.GetEmail(), err)
			}
		}

		result = append(result, s)
	}

	return result, nil
}

// ConvertDiskLicenses converts disk licenses from GCP API to DiskLicenseData.
func ConvertDiskLicenses(licenses []string) []DiskLicenseData {
	if len(licenses) == 0 {
		return nil
	}

	result := make([]DiskLicenseData, 0, len(licenses))
	for _, license := range licenses {
		result = append(result, DiskLicenseData{
			License: license,
		})
	}

	return result
}

// ConvertNICAccessConfigs converts NIC access configs from GCP API to AccessConfigData.
func ConvertNICAccessConfigs(accessConfigs []*computepb.AccessConfig) []AccessConfigData {
	if len(accessConfigs) == 0 {
		return nil
	}

	result := make([]AccessConfigData, 0, len(accessConfigs))
	for _, ac := range accessConfigs {
		result = append(result, AccessConfigData{
			Type:        ac.GetType(),
			Name:        ac.GetName(),
			NatIP:       ac.GetNatIP(),
			NetworkTier: ac.GetNetworkTier(),
		})
	}

	return result
}

// ConvertNICAliasRanges converts NIC alias IP ranges from GCP API to AliasRangeData.
func ConvertNICAliasRanges(aliasRanges []*computepb.AliasIpRange) []AliasRangeData {
	if len(aliasRanges) == 0 {
		return nil
	}

	result := make([]AliasRangeData, 0, len(aliasRanges))
	for _, ar := range aliasRanges {
		result = append(result, AliasRangeData{
			IPCidrRange:         ar.GetIpCidrRange(),
			SubnetworkRangeName: ar.GetSubnetworkRangeName(),
		})
	}

	return result
}
