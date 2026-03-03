package gcp

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	"github.com/dannyota/hotpot/pkg/ingest"
)

// GCPInventoryWorkflowParams contains parameters for the GCP inventory workflow.
type GCPInventoryWorkflowParams struct{}

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
	TotalSinks         int
	TotalLogBuckets    int
	TotalLogMetrics    int
	TotalLogExclusions int

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

	// Resource Manager (org-scoped)
	TotalProjects           int
	TotalOrganizations      int
	TotalFolders            int
	TotalOrgIamPolicies     int
	TotalFolderIamPolicies  int
	TotalProjectIamPolicies int
}

// ProjectResult contains the ingestion result for a single project.
type ProjectResult struct {
	ProjectID       string
	Error           string
	SkippedServices int

	// Compute
	InstanceCount        int
	InterconnectCount    int
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
	ApplicationCount int
	AppServiceCount  int

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

// aggregateFunc is the function signature for merging a service result into the provider-level results.
type aggregateFunc = func(*GCPInventoryWorkflowResult, *ProjectResult, any)

// GCPInventoryWorkflow ingests all GCP resources across multiple projects.
// It orchestrates compute, GKE, IAM, and other GCP resource ingestion.
func GCPInventoryWorkflow(ctx workflow.Context, _ GCPInventoryWorkflowParams) (*GCPInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPInventoryWorkflow")

	// Discover projects
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var discoverResult DiscoverProjectsResult
	err := workflow.ExecuteActivity(activityCtx, DiscoverProjectsActivity, DiscoverProjectsParams{}).
		Get(ctx, &discoverResult)
	if err != nil {
		logger.Error("Failed to discover projects", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Discovered projects", "count", len(discoverResult.ProjectIDs))

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
		ProjectResults: make([]ProjectResult, 0, len(discoverResult.ProjectIDs)),
	}

	services := ingest.Services("gcp")

	// Per-project services (ScopeRegional)
	for _, projectID := range discoverResult.ProjectIDs {
		projectResult := ProjectResult{ProjectID: projectID}

		// Discover which APIs are enabled for this project.
		var enabledAPIs map[string]bool
		var discoverAPIsResult DiscoverEnabledAPIsResult
		err := workflow.ExecuteActivity(activityCtx, DiscoverEnabledAPIsActivity,
			DiscoverEnabledAPIsParams{ProjectID: projectID}).Get(ctx, &discoverAPIsResult)
		if err != nil {
			logger.Error("Failed to discover enabled APIs; running all services",
				"projectID", projectID, "error", err)
		} else {
			enabledAPIs = make(map[string]bool, len(discoverAPIsResult.EnabledAPIs))
			for _, api := range discoverAPIsResult.EnabledAPIs {
				enabledAPIs[api] = true
			}
		}

		// Run regional services, skipping those whose API isn't enabled.
		for _, svc := range services {
			if svc.Scope != ingest.ScopeRegional {
				continue
			}
			if enabledAPIs != nil && svc.APIName != "" && !enabledAPIs[svc.APIName] {
				logger.Info("Skipping service: API not enabled",
					"service", svc.Name, "api", svc.APIName, "projectID", projectID)
				projectResult.SkippedServices++
				continue
			}
			res := svc.NewResult()
			err := workflow.ExecuteChildWorkflow(ctx, svc.Workflow,
				svc.NewParams(projectID, "")).Get(ctx, res)
			if err != nil {
				logger.Error("Failed ingestion", "service", svc.Name, "projectID", projectID, "error", err)
				appendError(&projectResult, err)
			} else {
				svc.Aggregate.(aggregateFunc)(result, &projectResult, res)
			}
		}

		result.ProjectResults = append(result.ProjectResults, projectResult)
	}

	// Org-scoped services (ScopeGlobal) — run once, not per-project
	for _, svc := range services {
		if svc.Scope != ingest.ScopeGlobal {
			continue
		}
		res := svc.NewResult()
		err := workflow.ExecuteChildWorkflow(ctx, svc.Workflow,
			svc.NewParams("", "")).Get(ctx, res)
		if err != nil {
			logger.Error("Failed ingestion", "service", svc.Name, "error", err)
		} else {
			svc.Aggregate.(aggregateFunc)(result, nil, res)
		}
	}

	logger.Info("Completed GCPInventoryWorkflow",
		"projectCount", len(discoverResult.ProjectIDs),
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
		"totalProjects", result.TotalProjects,
		"totalOrganizations", result.TotalOrganizations,
		"totalFolders", result.TotalFolders,
		"totalOrgIamPolicies", result.TotalOrgIamPolicies,
		"totalFolderIamPolicies", result.TotalFolderIamPolicies,
		"totalProjectIamPolicies", result.TotalProjectIamPolicies,
	)

	return result, nil
}

func appendError(pr *ProjectResult, err error) {
	if pr.Error == "" {
		pr.Error = err.Error()
	} else {
		pr.Error += "; " + err.Error()
	}
}
