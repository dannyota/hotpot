package gcp

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/container"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/dns"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/secretmanager"
	gcpsql "github.com/dannyota/hotpot/pkg/ingest/gcp/sql"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/storage"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/vpcaccess"
)

// GCPInventoryWorkflowParams contains parameters for the GCP inventory workflow.
type GCPInventoryWorkflowParams struct {
	ProjectIDs []string
}

// GCPInventoryWorkflowResult contains the result of the GCP inventory workflow.
type GCPInventoryWorkflowResult struct {
	ProjectResults       []ProjectResult
	TotalInstances       int
	TotalClusters        int
	TotalServiceAccounts int
	TotalConnectors      int
	TotalBuckets         int
	TotalKeyRings        int
	TotalCryptoKeys      int
	TotalSinks           int
	TotalLogBuckets      int
	TotalManagedZones       int
	TotalDNSPolicies        int
	TotalSecrets            int
	TotalSQLInstances       int
	TotalLogMetrics         int
	TotalLogExclusions      int
	TotalBucketIamPolicies  int
	TotalInterconnects      int
	TotalPacketMirrorings   int
	TotalProjectMetadata    int
}

// ProjectResult contains the ingestion result for a single project.
type ProjectResult struct {
	ProjectID           string
	InstanceCount       int
	ClusterCount        int
	ServiceAccountCount int
	ConnectorCount      int
	BucketCount         int
	KeyRingCount        int
	CryptoKeyCount      int
	SinkCount           int
	LogBucketCount      int
	ManagedZoneCount       int
	DNSPolicyCount         int
	SecretCount            int
	SQLInstanceCount       int
	LogMetricCount         int
	LogExclusionCount      int
	BucketIamPolicyCount   int
	InterconnectCount      int
	PacketMirroringCount   int
	ProjectMetadataCount   int
	Error                  string
}

// GCPInventoryWorkflow ingests all GCP resources across multiple projects.
// It orchestrates compute, GKE, IAM, and other GCP resource ingestion.
func GCPInventoryWorkflow(ctx workflow.Context, params GCPInventoryWorkflowParams) (*GCPInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPInventoryWorkflow", "projectCount", len(params.ProjectIDs))

	// Child workflow options
	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &GCPInventoryWorkflowResult{
		ProjectResults: make([]ProjectResult, 0, len(params.ProjectIDs)),
	}

	// Process each project
	for _, projectID := range params.ProjectIDs {
		projectResult := ProjectResult{ProjectID: projectID}

		// Execute GCPComputeWorkflow for this project
		var computeResult compute.GCPComputeWorkflowResult
		err := workflow.ExecuteChildWorkflow(ctx, compute.GCPComputeWorkflow, compute.GCPComputeWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &computeResult)

		if err != nil {
			logger.Error("Failed to execute GCPComputeWorkflow", "projectID", projectID, "error", err)
			projectResult.Error = err.Error()
		} else {
			projectResult.InstanceCount = computeResult.InstanceCount
			projectResult.InterconnectCount = computeResult.InterconnectCount
			projectResult.PacketMirroringCount = computeResult.PacketMirroringCount
			projectResult.ProjectMetadataCount = computeResult.ProjectMetadataCount
			result.TotalInstances += computeResult.InstanceCount
			result.TotalInterconnects += computeResult.InterconnectCount
			result.TotalPacketMirrorings += computeResult.PacketMirroringCount
			result.TotalProjectMetadata += computeResult.ProjectMetadataCount
		}

		// Execute GCPContainerWorkflow for this project
		var containerResult container.GCPContainerWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, container.GCPContainerWorkflow, container.GCPContainerWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &containerResult)

		if err != nil {
			logger.Error("Failed to execute GCPContainerWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.ClusterCount = containerResult.ClusterCount
			result.TotalClusters += containerResult.ClusterCount
		}

		// Execute GCPIAMWorkflow for this project
		var iamResult iam.GCPIAMWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, iam.GCPIAMWorkflow, iam.GCPIAMWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &iamResult)

		if err != nil {
			logger.Error("Failed to execute GCPIAMWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.ServiceAccountCount = iamResult.ServiceAccountCount
			result.TotalServiceAccounts += iamResult.ServiceAccountCount
		}

		// Execute GCPVpcAccessWorkflow for this project (after Compute, needs subnetwork regions)
		var vpcAccessResult vpcaccess.GCPVpcAccessWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, vpcaccess.GCPVpcAccessWorkflow, vpcaccess.GCPVpcAccessWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &vpcAccessResult)

		if err != nil {
			logger.Error("Failed to execute GCPVpcAccessWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.ConnectorCount = vpcAccessResult.ConnectorCount
			result.TotalConnectors += vpcAccessResult.ConnectorCount
		}

		// Execute GCPStorageWorkflow for this project
		var storageResult storage.GCPStorageWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, storage.GCPStorageWorkflow, storage.GCPStorageWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &storageResult)

		if err != nil {
			logger.Error("Failed to execute GCPStorageWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.BucketCount = storageResult.BucketCount
			projectResult.BucketIamPolicyCount = storageResult.BucketIamPolicyCount
			result.TotalBuckets += storageResult.BucketCount
			result.TotalBucketIamPolicies += storageResult.BucketIamPolicyCount
		}

		// Execute GCPKMSWorkflow for this project
		var kmsResult kms.GCPKMSWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, kms.GCPKMSWorkflow, kms.GCPKMSWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &kmsResult)

		if err != nil {
			logger.Error("Failed to execute GCPKMSWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.KeyRingCount = kmsResult.KeyRingCount
			projectResult.CryptoKeyCount = kmsResult.CryptoKeyCount
			result.TotalKeyRings += kmsResult.KeyRingCount
			result.TotalCryptoKeys += kmsResult.CryptoKeyCount
		}

		// Execute GCPLoggingWorkflow for this project
		var loggingResult logging.GCPLoggingWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, logging.GCPLoggingWorkflow, logging.GCPLoggingWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &loggingResult)

		if err != nil {
			logger.Error("Failed to execute GCPLoggingWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.SinkCount = loggingResult.SinkCount
			projectResult.LogBucketCount = loggingResult.BucketCount
			projectResult.LogMetricCount = loggingResult.LogMetricCount
			projectResult.LogExclusionCount = loggingResult.ExclusionCount
			result.TotalSinks += loggingResult.SinkCount
			result.TotalLogBuckets += loggingResult.BucketCount
			result.TotalLogMetrics += loggingResult.LogMetricCount
			result.TotalLogExclusions += loggingResult.ExclusionCount
		}

		// Execute GCPDNSWorkflow for this project
		var dnsResult dns.GCPDNSWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, dns.GCPDNSWorkflow, dns.GCPDNSWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &dnsResult)

		if err != nil {
			logger.Error("Failed to execute GCPDNSWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.ManagedZoneCount = dnsResult.ManagedZoneCount
			projectResult.DNSPolicyCount = dnsResult.PolicyCount
			result.TotalManagedZones += dnsResult.ManagedZoneCount
			result.TotalDNSPolicies += dnsResult.PolicyCount
		}

		// Execute GCPSecretManagerWorkflow for this project
		var secretManagerResult secretmanager.GCPSecretManagerWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, secretmanager.GCPSecretManagerWorkflow, secretmanager.GCPSecretManagerWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &secretManagerResult)

		if err != nil {
			logger.Error("Failed to execute GCPSecretManagerWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.SecretCount = secretManagerResult.SecretCount
			result.TotalSecrets += secretManagerResult.SecretCount
		}

		// Execute GCPSQLWorkflow for this project
		var sqlResult gcpsql.GCPSQLWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, gcpsql.GCPSQLWorkflow, gcpsql.GCPSQLWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &sqlResult)

		if err != nil {
			logger.Error("Failed to execute GCPSQLWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.SQLInstanceCount = sqlResult.InstanceCount
			result.TotalSQLInstances += sqlResult.InstanceCount
		}

		result.ProjectResults = append(result.ProjectResults, projectResult)
	}

	logger.Info("Completed GCPInventoryWorkflow",
		"projectCount", len(params.ProjectIDs),
		"totalInstances", result.TotalInstances,
		"totalClusters", result.TotalClusters,
		"totalServiceAccounts", result.TotalServiceAccounts,
		"totalConnectors", result.TotalConnectors,
		"totalBuckets", result.TotalBuckets,
		"totalKeyRings", result.TotalKeyRings,
		"totalCryptoKeys", result.TotalCryptoKeys,
		"totalSinks", result.TotalSinks,
		"totalLogBuckets", result.TotalLogBuckets,
		"totalManagedZones", result.TotalManagedZones,
		"totalSecrets", result.TotalSecrets,
		"totalSQLInstances", result.TotalSQLInstances,
		"totalLogMetrics", result.TotalLogMetrics,
		"totalLogExclusions", result.TotalLogExclusions,
		"totalDNSPolicies", result.TotalDNSPolicies,
		"totalBucketIamPolicies", result.TotalBucketIamPolicies,
		"totalInterconnects", result.TotalInterconnects,
		"totalPacketMirrorings", result.TotalPacketMirrorings,
		"totalProjectMetadata", result.TotalProjectMetadata,
	)

	return result, nil
}
