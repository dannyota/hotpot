package instance

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertInstance converts a GCP API Instance to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertInstance(inst *computepb.Instance, projectID string, collectedAt time.Time) bronze.GCPComputeInstance {
	instance := bronze.GCPComputeInstance{
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

	// Convert scheduling to JSON (kept as JSONB)
	if inst.Scheduling != nil {
		if data, err := json.Marshal(inst.Scheduling); err == nil {
			instance.SchedulingJSON = string(data)
		}
	}

	// Convert related entities
	instance.Disks = ConvertDisks(inst.Disks)
	instance.NICs = ConvertNICs(inst.NetworkInterfaces)
	instance.Labels = ConvertLabels(inst.Labels)
	instance.Tags = ConvertTags(inst.Tags)
	instance.Metadata = ConvertMetadata(inst.Metadata)
	instance.ServiceAccounts = ConvertServiceAccounts(inst.ServiceAccounts)

	return instance
}

// ConvertDisks converts attached disk info from GCP API to Bronze models.
func ConvertDisks(disks []*computepb.AttachedDisk) []bronze.GCPComputeInstanceDisk {
	if len(disks) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceDisk, 0, len(disks))
	for _, disk := range disks {
		d := bronze.GCPComputeInstanceDisk{
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

		// Convert encryption key to JSON (kept as JSONB)
		if disk.DiskEncryptionKey != nil {
			if data, err := json.Marshal(disk.DiskEncryptionKey); err == nil {
				d.DiskEncryptionKeyJSON = string(data)
			}
		}

		// Convert initialize params to JSON (kept as JSONB)
		if disk.InitializeParams != nil {
			if data, err := json.Marshal(disk.InitializeParams); err == nil {
				d.InitializeParamsJSON = string(data)
			}
		}

		// Convert licenses to separate table
		d.Licenses = ConvertDiskLicenses(disk.Licenses)

		result = append(result, d)
	}

	return result
}

// ConvertNICs converts network interfaces from GCP API to Bronze models.
func ConvertNICs(nics []*computepb.NetworkInterface) []bronze.GCPComputeInstanceNIC {
	if len(nics) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceNIC, 0, len(nics))
	for _, nic := range nics {
		n := bronze.GCPComputeInstanceNIC{
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
		n.AliasIpRanges = ConvertNICAliasRanges(nic.AliasIpRanges)

		result = append(result, n)
	}

	return result
}

// ConvertLabels converts instance labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeInstanceLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeInstanceLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// ConvertTags converts network tags from GCP API to Bronze models.
func ConvertTags(tags *computepb.Tags) []bronze.GCPComputeInstanceTag {
	if tags == nil || len(tags.Items) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceTag, 0, len(tags.Items))
	for _, tag := range tags.Items {
		result = append(result, bronze.GCPComputeInstanceTag{
			Tag: tag,
		})
	}

	return result
}

// ConvertMetadata converts instance metadata from GCP API to Bronze models.
func ConvertMetadata(metadata *computepb.Metadata) []bronze.GCPComputeInstanceMetadata {
	if metadata == nil || len(metadata.Items) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceMetadata, 0, len(metadata.Items))
	for _, item := range metadata.Items {
		m := bronze.GCPComputeInstanceMetadata{
			Key: item.GetKey(),
		}
		if item.Value != nil {
			m.Value = *item.Value
		}
		result = append(result, m)
	}

	return result
}

// ConvertServiceAccounts converts service accounts from GCP API to Bronze models.
func ConvertServiceAccounts(serviceAccounts []*computepb.ServiceAccount) []bronze.GCPComputeInstanceServiceAccount {
	if len(serviceAccounts) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceServiceAccount, 0, len(serviceAccounts))
	for _, sa := range serviceAccounts {
		s := bronze.GCPComputeInstanceServiceAccount{
			Email: sa.GetEmail(),
		}

		// Convert scopes to JSON
		if len(sa.Scopes) > 0 {
			if data, err := json.Marshal(sa.Scopes); err == nil {
				s.ScopesJSON = string(data)
			}
		}

		result = append(result, s)
	}

	return result
}

// ConvertDiskLicenses converts disk licenses from GCP API to Bronze models.
func ConvertDiskLicenses(licenses []string) []bronze.GCPComputeInstanceDiskLicense {
	if len(licenses) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceDiskLicense, 0, len(licenses))
	for _, license := range licenses {
		result = append(result, bronze.GCPComputeInstanceDiskLicense{
			License: license,
		})
	}

	return result
}

// ConvertNICAccessConfigs converts NIC access configs from GCP API to Bronze models.
func ConvertNICAccessConfigs(accessConfigs []*computepb.AccessConfig) []bronze.GCPComputeInstanceNICAccessConfig {
	if len(accessConfigs) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceNICAccessConfig, 0, len(accessConfigs))
	for _, ac := range accessConfigs {
		result = append(result, bronze.GCPComputeInstanceNICAccessConfig{
			Type:        ac.GetType(),
			Name:        ac.GetName(),
			NatIP:       ac.GetNatIP(),
			NetworkTier: ac.GetNetworkTier(),
		})
	}

	return result
}

// ConvertNICAliasRanges converts NIC alias IP ranges from GCP API to Bronze models.
func ConvertNICAliasRanges(aliasRanges []*computepb.AliasIpRange) []bronze.GCPComputeInstanceNICAliasRange {
	if len(aliasRanges) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeInstanceNICAliasRange, 0, len(aliasRanges))
	for _, ar := range aliasRanges {
		result = append(result, bronze.GCPComputeInstanceNICAliasRange{
			IpCidrRange:         ar.GetIpCidrRange(),
			SubnetworkRangeName: ar.GetSubnetworkRangeName(),
		})
	}

	return result
}
