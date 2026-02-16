package gcp

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/alloydb"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/appengine"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigquery"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigtable"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/binaryauthorization"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudfunctions"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/container"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/containeranalysis"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/dataproc"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/dns"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/filestore"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iap"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/monitoring"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/pubsub"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/redis"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/run"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/secretmanager"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/securitycenter"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/serviceusage"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/spanner"
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
	ProjectResults []ProjectResult

	// Compute
	TotalInstances        int
	TotalInterconnects    int
	TotalPacketMirrorings int
	TotalProjectMetadata  int

	// Container (GKE)
	TotalClusters int

	// IAM
	TotalServiceAccounts int

	// VPC Access
	TotalConnectors int

	// Storage
	TotalBuckets           int
	TotalBucketIamPolicies int

	// KMS
	TotalKeyRings   int
	TotalCryptoKeys int

	// Logging
	TotalSinks          int
	TotalLogBuckets     int
	TotalLogMetrics     int
	TotalLogExclusions  int

	// DNS
	TotalManagedZones int
	TotalDNSPolicies  int

	// Secret Manager
	TotalSecrets int

	// Cloud SQL
	TotalSQLInstances int

	// Security Command Center (org-scoped)
	TotalSources  int
	TotalFindings int

	// Organization Policy (org-scoped)
	TotalConstraints int
	TotalOrgPolicies int

	// Service Usage
	TotalEnabledServices int

	// Cloud Functions
	TotalFunctions int

	// Memorystore Redis
	TotalRedisInstances int

	// Dataproc
	TotalDataprocClusters int

	// IAP
	TotalIAPSettings int
	TotalIAPPolicies int

	// AlloyDB
	TotalAlloyDBClusters int

	// Filestore
	TotalFilestoreInstances int

	// Pub/Sub
	TotalTopics        int
	TotalSubscriptions int

	// App Engine
	TotalApplications int
	TotalAppServices  int

	// Cloud Asset (org-scoped)
	TotalAssets         int
	TotalAssetPolicies  int
	TotalAssetResources int

	// Binary Authorization
	TotalBinAuthPolicies int
	TotalAttestors       int

	// Monitoring
	TotalAlertPolicies int
	TotalUptimeChecks  int

	// Cloud Run
	TotalRunServices  int
	TotalRunRevisions int

	// Access Context Manager (org-scoped)
	TotalAccessPolicies    int
	TotalAccessLevels      int
	TotalServicePerimeters int

	// Container Analysis
	TotalNotes       int
	TotalOccurrences int

	// Spanner
	TotalSpannerInstances int
	TotalSpannerDatabases int

	// BigQuery
	TotalDatasets int
	TotalTables   int

	// Bigtable
	TotalBigtableInstances int
	TotalBigtableClusters  int
}

// ProjectResult contains the ingestion result for a single project.
type ProjectResult struct {
	ProjectID string
	Error     string

	// Compute
	InstanceCount       int
	InterconnectCount   int
	PacketMirroringCount int
	ProjectMetadataCount int

	// Container (GKE)
	ClusterCount int

	// IAM
	ServiceAccountCount int

	// VPC Access
	ConnectorCount int

	// Storage
	BucketCount          int
	BucketIamPolicyCount int

	// KMS
	KeyRingCount   int
	CryptoKeyCount int

	// Logging
	SinkCount         int
	LogBucketCount    int
	LogMetricCount    int
	LogExclusionCount int

	// DNS
	ManagedZoneCount int
	DNSPolicyCount   int

	// Secret Manager
	SecretCount int

	// Cloud SQL
	SQLInstanceCount int

	// Service Usage
	EnabledServiceCount int

	// Cloud Functions
	FunctionCount int

	// Memorystore Redis
	RedisInstanceCount int

	// Dataproc
	DataprocClusterCount int

	// IAP
	IAPSettingsCount int
	IAPPolicyCount   int

	// AlloyDB
	AlloyDBClusterCount int

	// Filestore
	FilestoreInstanceCount int

	// Pub/Sub
	TopicCount        int
	SubscriptionCount int

	// App Engine
	ApplicationCount  int
	AppServiceCount   int

	// Binary Authorization
	BinAuthPolicyCount int
	AttestorCount      int

	// Monitoring
	AlertPolicyCount int
	UptimeCheckCount int

	// Cloud Run
	RunServiceCount  int
	RunRevisionCount int

	// Container Analysis
	NoteCount       int
	OccurrenceCount int

	// Spanner
	SpannerInstanceCount int
	SpannerDatabaseCount int

	// BigQuery
	DatasetCount int
	TableCount   int

	// Bigtable
	BigtableInstanceCount int
	BigtableClusterCount  int
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

		// Execute GCPServiceUsageWorkflow for this project
		var serviceUsageResult serviceusage.GCPServiceUsageWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, serviceusage.GCPServiceUsageWorkflow, serviceusage.GCPServiceUsageWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &serviceUsageResult)

		if err != nil {
			logger.Error("Failed to execute GCPServiceUsageWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.EnabledServiceCount = serviceUsageResult.ServiceCount
			result.TotalEnabledServices += serviceUsageResult.ServiceCount
		}

		// Execute GCPCloudFunctionsWorkflow for this project
		var cloudFunctionsResult cloudfunctions.GCPCloudFunctionsWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, cloudfunctions.GCPCloudFunctionsWorkflow, cloudfunctions.GCPCloudFunctionsWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &cloudFunctionsResult)

		if err != nil {
			logger.Error("Failed to execute GCPCloudFunctionsWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.FunctionCount = cloudFunctionsResult.FunctionCount
			result.TotalFunctions += cloudFunctionsResult.FunctionCount
		}

		// Execute GCPRedisWorkflow for this project
		var redisResult redis.GCPRedisWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, redis.GCPRedisWorkflow, redis.GCPRedisWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &redisResult)

		if err != nil {
			logger.Error("Failed to execute GCPRedisWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.RedisInstanceCount = redisResult.InstanceCount
			result.TotalRedisInstances += redisResult.InstanceCount
		}

		// Execute GCPDataprocWorkflow for this project
		var dataprocResult dataproc.GCPDataprocWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, dataproc.GCPDataprocWorkflow, dataproc.GCPDataprocWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &dataprocResult)

		if err != nil {
			logger.Error("Failed to execute GCPDataprocWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.DataprocClusterCount = dataprocResult.ClusterCount
			result.TotalDataprocClusters += dataprocResult.ClusterCount
		}

		// Execute GCPIAPWorkflow for this project
		var iapResult iap.GCPIAPWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, iap.GCPIAPWorkflow, iap.GCPIAPWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &iapResult)

		if err != nil {
			logger.Error("Failed to execute GCPIAPWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.IAPSettingsCount = iapResult.SettingsCount
			projectResult.IAPPolicyCount = iapResult.PolicyCount
			result.TotalIAPSettings += iapResult.SettingsCount
			result.TotalIAPPolicies += iapResult.PolicyCount
		}

		// Execute GCPAlloyDBWorkflow for this project
		var alloyDBResult alloydb.GCPAlloyDBWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, alloydb.GCPAlloyDBWorkflow, alloydb.GCPAlloyDBWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &alloyDBResult)

		if err != nil {
			logger.Error("Failed to execute GCPAlloyDBWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.AlloyDBClusterCount = alloyDBResult.ClusterCount
			result.TotalAlloyDBClusters += alloyDBResult.ClusterCount
		}

		// Execute GCPFilestoreWorkflow for this project
		var filestoreResult filestore.GCPFilestoreWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, filestore.GCPFilestoreWorkflow, filestore.GCPFilestoreWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &filestoreResult)

		if err != nil {
			logger.Error("Failed to execute GCPFilestoreWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.FilestoreInstanceCount = filestoreResult.InstanceCount
			result.TotalFilestoreInstances += filestoreResult.InstanceCount
		}

		// Execute GCPPubSubWorkflow for this project
		var pubsubResult pubsub.GCPPubSubWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, pubsub.GCPPubSubWorkflow, pubsub.GCPPubSubWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &pubsubResult)

		if err != nil {
			logger.Error("Failed to execute GCPPubSubWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.TopicCount = pubsubResult.TopicCount
			projectResult.SubscriptionCount = pubsubResult.SubscriptionCount
			result.TotalTopics += pubsubResult.TopicCount
			result.TotalSubscriptions += pubsubResult.SubscriptionCount
		}

		// Execute GCPAppEngineWorkflow for this project
		var appEngineResult appengine.GCPAppEngineWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, appengine.GCPAppEngineWorkflow, appengine.GCPAppEngineWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &appEngineResult)

		if err != nil {
			logger.Error("Failed to execute GCPAppEngineWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.ApplicationCount = appEngineResult.ApplicationCount
			projectResult.AppServiceCount = appEngineResult.ServiceCount
			result.TotalApplications += appEngineResult.ApplicationCount
			result.TotalAppServices += appEngineResult.ServiceCount
		}

		// Execute GCPBinaryAuthorizationWorkflow for this project
		var binAuthResult binaryauthorization.GCPBinaryAuthorizationWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, binaryauthorization.GCPBinaryAuthorizationWorkflow, binaryauthorization.GCPBinaryAuthorizationWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &binAuthResult)

		if err != nil {
			logger.Error("Failed to execute GCPBinaryAuthorizationWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.BinAuthPolicyCount = binAuthResult.PolicyCount
			projectResult.AttestorCount = binAuthResult.AttestorCount
			result.TotalBinAuthPolicies += binAuthResult.PolicyCount
			result.TotalAttestors += binAuthResult.AttestorCount
		}

		// Execute GCPMonitoringWorkflow for this project
		var monitoringResult monitoring.GCPMonitoringWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, monitoring.GCPMonitoringWorkflow, monitoring.GCPMonitoringWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &monitoringResult)

		if err != nil {
			logger.Error("Failed to execute GCPMonitoringWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.AlertPolicyCount = monitoringResult.AlertPolicyCount
			projectResult.UptimeCheckCount = monitoringResult.UptimeCheckCount
			result.TotalAlertPolicies += monitoringResult.AlertPolicyCount
			result.TotalUptimeChecks += monitoringResult.UptimeCheckCount
		}

		// Execute GCPRunWorkflow for this project
		var runResult run.GCPRunWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, run.GCPRunWorkflow, run.GCPRunWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &runResult)

		if err != nil {
			logger.Error("Failed to execute GCPRunWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.RunServiceCount = runResult.ServiceCount
			projectResult.RunRevisionCount = runResult.RevisionCount
			result.TotalRunServices += runResult.ServiceCount
			result.TotalRunRevisions += runResult.RevisionCount
		}

		// Execute GCPContainerAnalysisWorkflow for this project
		var containerAnalysisResult containeranalysis.GCPContainerAnalysisWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, containeranalysis.GCPContainerAnalysisWorkflow, containeranalysis.GCPContainerAnalysisWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &containerAnalysisResult)

		if err != nil {
			logger.Error("Failed to execute GCPContainerAnalysisWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.NoteCount = containerAnalysisResult.NoteCount
			projectResult.OccurrenceCount = containerAnalysisResult.OccurrenceCount
			result.TotalNotes += containerAnalysisResult.NoteCount
			result.TotalOccurrences += containerAnalysisResult.OccurrenceCount
		}

		// Execute GCPSpannerWorkflow for this project
		var spannerResult spanner.GCPSpannerWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, spanner.GCPSpannerWorkflow, spanner.GCPSpannerWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &spannerResult)

		if err != nil {
			logger.Error("Failed to execute GCPSpannerWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.SpannerInstanceCount = spannerResult.InstanceCount
			projectResult.SpannerDatabaseCount = spannerResult.DatabaseCount
			result.TotalSpannerInstances += spannerResult.InstanceCount
			result.TotalSpannerDatabases += spannerResult.DatabaseCount
		}

		// Execute GCPBigQueryWorkflow for this project
		var bigqueryResult bigquery.GCPBigQueryWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, bigquery.GCPBigQueryWorkflow, bigquery.GCPBigQueryWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &bigqueryResult)

		if err != nil {
			logger.Error("Failed to execute GCPBigQueryWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.DatasetCount = bigqueryResult.DatasetCount
			projectResult.TableCount = bigqueryResult.TableCount
			result.TotalDatasets += bigqueryResult.DatasetCount
			result.TotalTables += bigqueryResult.TableCount
		}

		// Execute GCPBigtableWorkflow for this project
		var bigtableResult bigtable.GCPBigtableWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, bigtable.GCPBigtableWorkflow, bigtable.GCPBigtableWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &bigtableResult)

		if err != nil {
			logger.Error("Failed to execute GCPBigtableWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.BigtableInstanceCount = bigtableResult.InstanceCount
			projectResult.BigtableClusterCount = bigtableResult.ClusterCount
			result.TotalBigtableInstances += bigtableResult.InstanceCount
			result.TotalBigtableClusters += bigtableResult.ClusterCount
		}

		result.ProjectResults = append(result.ProjectResults, projectResult)
	}

	// Org-level workflows (run once, not per-project)

	// Execute GCPSecurityCenterWorkflow (org-scoped, queries orgs from DB)
	var sccResult securitycenter.GCPSecurityCenterWorkflowResult
	err := workflow.ExecuteChildWorkflow(ctx, securitycenter.GCPSecurityCenterWorkflow,
		securitycenter.GCPSecurityCenterWorkflowParams{}).Get(ctx, &sccResult)
	if err != nil {
		logger.Error("Failed to execute GCPSecurityCenterWorkflow", "error", err)
	} else {
		result.TotalSources = sccResult.SourceCount
		result.TotalFindings = sccResult.FindingCount
	}

	// Execute GCPOrgPolicyWorkflow (org-scoped, queries orgs from DB)
	var orgPolicyResult orgpolicy.GCPOrgPolicyWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, orgpolicy.GCPOrgPolicyWorkflow,
		orgpolicy.GCPOrgPolicyWorkflowParams{}).Get(ctx, &orgPolicyResult)
	if err != nil {
		logger.Error("Failed to execute GCPOrgPolicyWorkflow", "error", err)
	} else {
		result.TotalConstraints = orgPolicyResult.ConstraintCount
		result.TotalOrgPolicies = orgPolicyResult.PolicyCount
	}

	// Execute GCPCloudAssetWorkflow (org-scoped, queries orgs from DB)
	var cloudAssetResult cloudasset.GCPCloudAssetWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, cloudasset.GCPCloudAssetWorkflow,
		cloudasset.GCPCloudAssetWorkflowParams{}).Get(ctx, &cloudAssetResult)
	if err != nil {
		logger.Error("Failed to execute GCPCloudAssetWorkflow", "error", err)
	} else {
		result.TotalAssets = cloudAssetResult.AssetCount
		result.TotalAssetPolicies = cloudAssetResult.PolicyCount
		result.TotalAssetResources = cloudAssetResult.ResourceCount
	}

	// Execute GCPAccessContextManagerWorkflow (org-scoped, queries orgs from DB)
	var acmResult accesscontextmanager.GCPAccessContextManagerWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, accesscontextmanager.GCPAccessContextManagerWorkflow,
		accesscontextmanager.GCPAccessContextManagerWorkflowParams{}).Get(ctx, &acmResult)
	if err != nil {
		logger.Error("Failed to execute GCPAccessContextManagerWorkflow", "error", err)
	} else {
		result.TotalAccessPolicies = acmResult.PolicyCount
		result.TotalAccessLevels = acmResult.LevelCount
		result.TotalServicePerimeters = acmResult.PerimeterCount
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
		"totalSources", result.TotalSources,
		"totalFindings", result.TotalFindings,
		"totalConstraints", result.TotalConstraints,
		"totalOrgPolicies", result.TotalOrgPolicies,
		"totalEnabledServices", result.TotalEnabledServices,
		"totalFunctions", result.TotalFunctions,
		"totalRedisInstances", result.TotalRedisInstances,
		"totalDataprocClusters", result.TotalDataprocClusters,
		"totalIAPSettings", result.TotalIAPSettings,
		"totalIAPPolicies", result.TotalIAPPolicies,
		"totalAlloyDBClusters", result.TotalAlloyDBClusters,
		"totalFilestoreInstances", result.TotalFilestoreInstances,
		"totalTopics", result.TotalTopics,
		"totalSubscriptions", result.TotalSubscriptions,
		"totalApplications", result.TotalApplications,
		"totalAppServices", result.TotalAppServices,
		"totalAssets", result.TotalAssets,
		"totalAssetPolicies", result.TotalAssetPolicies,
		"totalAssetResources", result.TotalAssetResources,
		"totalBinAuthPolicies", result.TotalBinAuthPolicies,
		"totalAttestors", result.TotalAttestors,
		"totalAlertPolicies", result.TotalAlertPolicies,
		"totalUptimeChecks", result.TotalUptimeChecks,
		"totalRunServices", result.TotalRunServices,
		"totalRunRevisions", result.TotalRunRevisions,
		"totalAccessPolicies", result.TotalAccessPolicies,
		"totalAccessLevels", result.TotalAccessLevels,
		"totalServicePerimeters", result.TotalServicePerimeters,
		"totalNotes", result.TotalNotes,
		"totalOccurrences", result.TotalOccurrences,
		"totalSpannerInstances", result.TotalSpannerInstances,
		"totalSpannerDatabases", result.TotalSpannerDatabases,
		"totalDatasets", result.TotalDatasets,
		"totalTables", result.TotalTables,
		"totalBigtableInstances", result.TotalBigtableInstances,
		"totalBigtableClusters", result.TotalBigtableClusters,
	)

	return result, nil
}
