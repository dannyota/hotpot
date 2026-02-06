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
func ConvertInstance(inst *computepb.Instance, projectID string, collectedAt time.Time) (bronze.GCPComputeInstance, error) {
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

	// Convert scheduling to JSONB (nil → SQL NULL, data → JSON bytes)
	if inst.Scheduling != nil {
		var err error
		instance.SchedulingJSON, err = json.Marshal(inst.Scheduling)
		if err != nil {
			return bronze.GCPComputeInstance{}, fmt.Errorf("failed to marshal scheduling for instance %s: %w", inst.GetName(), err)
		}
	}

	// Convert related entities
	var convertErr error
	instance.Disks, convertErr = ConvertDisks(inst.Disks)
	if convertErr != nil {
		return bronze.GCPComputeInstance{}, fmt.Errorf("failed to convert disks for instance %s: %w", inst.GetName(), convertErr)
	}
	instance.NICs = ConvertNICs(inst.NetworkInterfaces)
	instance.Labels = ConvertLabels(inst.Labels)
	instance.Tags = ConvertTags(inst.Tags)
	instance.Metadata = ConvertMetadata(inst.Metadata)
	instance.ServiceAccounts, convertErr = ConvertServiceAccounts(inst.ServiceAccounts)
	if convertErr != nil {
		return bronze.GCPComputeInstance{}, fmt.Errorf("failed to convert service accounts for instance %s: %w", inst.GetName(), convertErr)
	}

	return instance, nil
}

// ConvertDisks converts attached disk info from GCP API to Bronze models.
func ConvertDisks(disks []*computepb.AttachedDisk) ([]bronze.GCPComputeInstanceDisk, error) {
	if len(disks) == 0 {
		return nil, nil
	}

	var err error
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
func ConvertServiceAccounts(serviceAccounts []*computepb.ServiceAccount) ([]bronze.GCPComputeInstanceServiceAccount, error) {
	if len(serviceAccounts) == 0 {
		return nil, nil
	}

	var err error
	result := make([]bronze.GCPComputeInstanceServiceAccount, 0, len(serviceAccounts))
	for _, sa := range serviceAccounts {
		s := bronze.GCPComputeInstanceServiceAccount{
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
