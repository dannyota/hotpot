# GCP

GCP resource ingestion coverage in the bronze layer.

## Compute Engine API (`compute.googleapis.com`)

### Compute

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| VM Instances | `InstancesClient` | `AggregatedList()` | ✅ |
| Disks | `DisksClient` | `AggregatedList()` | ✅ |
| Instance Groups | `InstanceGroupsClient` | `AggregatedList()` | ✅ |
| Instance Group Members | `InstanceGroupsClient` | `ListInstances()` | ✅ |
| Target Instances | `TargetInstancesClient` | `AggregatedList()` | ✅ |
| Snapshots | `SnapshotsClient` | `List()` | ✅ |
| Images | `ImagesClient` | `List()` | ✅ |
| Health Checks | `HealthChecksClient` | `AggregatedList()` | ✅ |

### Networking

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Networks | `NetworksClient` | `List()` | ✅ |
| Subnetworks | `SubnetworksClient` | `AggregatedList()` | ✅ |
| Firewall Rules | `FirewallsClient` | `List()` | |
| Routers | `RoutersClient` | `AggregatedList()` | |
| Interconnects | `InterconnectsClient` | `List()` | |
| Packet Mirrorings | `PacketMirroringsClient` | `AggregatedList()` | |
| Addresses (Regional) | `AddressesClient` | `AggregatedList()` | ✅ |
| Addresses (Global) | `GlobalAddressesClient` | `List()` | ✅ |
| SSL Policies | `SslPoliciesClient` | `List()` | |
| Security Policies | `SecurityPoliciesClient` | `List()` | |
| Project Metadata | `ProjectsClient` | `Get()` | |

### Load Balancing

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Forwarding Rules (Regional) | `ForwardingRulesClient` | `AggregatedList()` | ✅ |
| Forwarding Rules (Global) | `GlobalForwardingRulesClient` | `List()` | ✅ |
| Backend Services | `BackendServicesClient` | `AggregatedList()` | |
| URL Maps | `UrlMapsClient` | `AggregatedList()` | |
| Target HTTP Proxies | `TargetHttpProxiesClient` | `AggregatedList()` | |
| Target HTTPS Proxies | `TargetHttpsProxiesClient` | `AggregatedList()` | |
| Target SSL Proxies | `TargetSslProxiesClient` | `List()` | |
| Target TCP Proxies | `TargetTcpProxiesClient` | `List()` | |
| Target Pools | `TargetPoolsClient` | `AggregatedList()` | |
| NEGs | `NetworkEndpointGroupsClient` | `AggregatedList()` | |
| NEG Endpoints | `NetworkEndpointGroupsClient` | `ListNetworkEndpoints()` | |

### VPN

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| VPN Gateways (HA) | `VpnGatewaysClient` | `AggregatedList()` | ✅ |
| VPN Gateways (Classic) | `TargetVpnGatewaysClient` | `AggregatedList()` | ✅ |
| VPN Tunnels | `VpnTunnelsClient` | `AggregatedList()` | ✅ |

## Container API (`container.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| GKE Clusters | `ClusterManagerClient` | `ListClusters()` | ✅ |
| Node Pools | (included in cluster response) | — | ✅ |

## Resource Manager API (`cloudresourcemanager.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Organizations | `OrganizationsClient` | `SearchOrganizations()` | |
| Folders | `FoldersClient` | `ListFolders()` | |
| Projects | `ProjectsClient` | `SearchProjects()` | ✅ |
| Organization IAM Policies | `OrganizationsClient` | `GetIamPolicy()` | |
| Folder IAM Policies | `FoldersClient` | `GetIamPolicy()` | |
| Project IAM Policies | `ProjectsClient` | `GetIamPolicy()` | |

## IAM API (`iam.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Service Accounts | `IAMClient` | `ListServiceAccounts()` | ✅ |
| Service Account Keys | `IAMClient` | `ListServiceAccountKeys()` | ✅ |

## Cloud KMS API (`cloudkms.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Key Rings | `KeyManagementClient` | `ListKeyRings()` | |
| Crypto Keys | `KeyManagementClient` | `ListCryptoKeys()` | |

## API Keys API (`apikeys.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| API Keys | `ApiKeysClient` | `ListKeys()` | |

## Essential Contacts API (`essentialcontacts.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Essential Contacts | `EssentialContactsClient` | `ListContacts()` | |

## Cloud Functions API (`cloudfunctions.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Cloud Functions | `CloudFunctionsClient` | `ListFunctions()` | |

## Logging API (`logging.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Log Sinks | `ConfigClient` | `ListSinks()` | |
| Log Metrics | `MetricsClient` | `ListLogMetrics()` | |
| Log Buckets | `ConfigClient` | `ListBuckets()` | |
| Log Exclusions | `ConfigClient` | `ListExclusions()` | |

## Monitoring API (`monitoring.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Alert Policies | `AlertPolicyClient` | `ListAlertPolicies()` | |
| Uptime Check Configs | `UptimeCheckClient` | `ListUptimeCheckConfigs()` | |

## DNS API (`dns.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Managed Zones | `ManagedZonesClient` | `List()` | |
| DNS Policies | `PoliciesClient` | `List()` | |

## Access Approval API (`accessapproval.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Access Approval Settings | `AccessApprovalClient` | `GetAccessApprovalSettings()` | |
| Approval Requests | `AccessApprovalClient` | `ListApprovalRequests()` | |

## Cloud Storage API (`storage.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Buckets | `StorageClient` | `ListBuckets()` | |
| Bucket IAM Policies | `StorageClient` | `GetIamPolicy()` | |

## Cloud SQL Admin API (`sqladmin.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| SQL Instances | `SqlInstancesClient` | `List()` | |

## BigQuery API (`bigquery.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Datasets | `BigQueryClient` | `ListDatasets()` | |
| Tables | `BigQueryClient` | `ListTables()` | |

## Dataproc API (`dataproc.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Clusters | `ClusterControllerClient` | `ListClusters()` | |

## Service Usage API (`serviceusage.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Enabled Services | `ServiceUsageClient` | `ListServices()` | |

## Secret Manager API (`secretmanager.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Secrets | `SecretManagerClient` | `ListSecrets()` | |

## App Engine Admin API (`appengine.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Applications | `ApplicationsClient` | `GetApplication()` | |
| Services | `ServicesClient` | `ListServices()` | |

## Security Command Center API (`securitycenter.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Sources | `SecurityCenterClient` | `ListSources()` | |
| Findings | `SecurityCenterClient` | `ListFindings()` | |
| Notification Configs | `SecurityCenterClient` | `ListNotificationConfigs()` | |

## Organization Policy API (`orgpolicy.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Constraints | `OrgPolicyClient` | `ListConstraints()` | |
| Org Policies | `OrgPolicyClient` | `ListPolicies()` | |
| Custom Constraints | `OrgPolicyClient` | `ListCustomConstraints()` | |

## Access Context Manager API (`accesscontextmanager.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Access Policies | `AccessContextManagerClient` | `ListAccessPolicies()` | |
| Access Levels | `AccessContextManagerClient` | `ListAccessLevels()` | |
| Service Perimeters | `AccessContextManagerClient` | `ListServicePerimeters()` | |

## Cloud Asset API (`cloudasset.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Assets | `AssetServiceClient` | `ListAssets()` | |
| IAM Policy Search | `AssetServiceClient` | `SearchAllIamPolicies()` | |
| Resource Search | `AssetServiceClient` | `SearchAllResources()` | |

## Sensitive Data Protection API (`dlp.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Inspect Templates | `DlpServiceClient` | `ListInspectTemplates()` | |
| Deidentify Templates | `DlpServiceClient` | `ListDeidentifyTemplates()` | |
| Job Triggers | `DlpServiceClient` | `ListJobTriggers()` | |
| DLP Jobs | `DlpServiceClient` | `ListDlpJobs()` | |
| Discovery Configs | `DlpServiceClient` | `ListDiscoveryConfigs()` | |

## Binary Authorization API (`binaryauthorization.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Policy | `BinauthzManagementClient` | `GetPolicy()` | |
| Attestors | `BinauthzManagementClient` | `ListAttestors()` | |

## Container Analysis API (`containeranalysis.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Notes | `GrafeasClient` | `ListNotes()` | |
| Occurrences | `GrafeasClient` | `ListOccurrences()` | |

## Certificate Authority Service API (`privateca.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| CA Pools | `CertificateAuthorityClient` | `ListCaPools()` | |
| Certificate Authorities | `CertificateAuthorityClient` | `ListCertificateAuthorities()` | |
| Certificates | `CertificateAuthorityClient` | `ListCertificates()` | |

## Assured Workloads API (`assuredworkloads.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Workloads | `AssuredWorkloadsClient` | `ListWorkloads()` | |
| Violations | `AssuredWorkloadsClient` | `ListViolations()` | |

## Cloud IDS API (`ids.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Endpoints | `IDSClient` | `ListEndpoints()` | |

## Backup and DR API (`backupdr.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Backup Vaults | `BackupDRClient` | `ListBackupVaults()` | |
| Backup Plans | `BackupDRClient` | `ListBackupPlans()` | |
| Backup Plan Associations | `BackupDRClient` | `ListBackupPlanAssociations()` | |

## Web Security Scanner API (`websecurityscanner.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Scan Configs | `WebSecurityScannerClient` | `ListScanConfigs()` | |
| Scan Runs | `WebSecurityScannerClient` | `ListScanRuns()` | |

## Identity-Aware Proxy API (`iap.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| IAP Settings | `IdentityAwareProxyAdminClient` | `GetIapSettings()` | |
| IAP IAM Policies | `IdentityAwareProxyAdminClient` | `GetIamPolicy()` | |

## Recommender API (`recommender.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| IAM Recommendations | `RecommenderClient` | `ListRecommendations()` | |
| IAM Insights | `RecommenderClient` | `ListInsights()` | |

## Cloud Billing API (`cloudbilling.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Billing Accounts | `CloudBillingClient` | `ListBillingAccounts()` | |
| Project Billing Info | `CloudBillingClient` | `GetProjectBillingInfo()` | |
| Budgets | `BudgetServiceClient` | `ListBudgets()` | |

## Network Management API (`networkmanagement.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Connectivity Tests | `ReachabilityServiceClient` | `ListConnectivityTests()` | |

## Cloud Run API (`run.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Services | `ServicesClient` | `ListServices()` | |
| Revisions | `RevisionsClient` | `ListRevisions()` | |

## Pub/Sub API (`pubsub.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Topics | `PublisherClient` | `ListTopics()` | |
| Subscriptions | `SubscriberClient` | `ListSubscriptions()` | |

## Spanner API (`spanner.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Instances | `InstanceAdminClient` | `ListInstances()` | |
| Databases | `DatabaseAdminClient` | `ListDatabases()` | |

## Bigtable Admin API (`bigtableadmin.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Instances | `BigtableInstanceAdminClient` | `ListInstances()` | |
| Clusters | `BigtableInstanceAdminClient` | `ListClusters()` | |

## Memorystore for Redis API (`redis.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Instances | `CloudRedisClient` | `ListInstances()` | |

## Filestore API (`file.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Instances | `CloudFilestoreManagerClient` | `ListInstances()` | |

## AlloyDB API (`alloydb.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Clusters | `AlloyDBAdminClient` | `ListClusters()` | |

## VPC Access API (`vpcaccess.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| VPC Connectors | `VpcAccessClient` | `ListConnectors()` | :white_check_mark: |

## Summary

**Total: 19/117 (16%)**

| API | Implemented | Total |
|-----|:-----------:|:-----:|
| Compute Engine | 14 | 32 |
| Container | 2 | 2 |
| Resource Manager | 1 | 6 |
| IAM | 2 | 2 |
| Cloud KMS | 0 | 2 |
| API Keys | 0 | 1 |
| Essential Contacts | 0 | 1 |
| Cloud Functions | 0 | 1 |
| Logging | 0 | 4 |
| Monitoring | 0 | 2 |
| DNS | 0 | 2 |
| Access Approval | 0 | 2 |
| Cloud Storage | 0 | 2 |
| Cloud SQL Admin | 0 | 1 |
| BigQuery | 0 | 2 |
| Dataproc | 0 | 1 |
| Service Usage | 0 | 1 |
| Secret Manager | 0 | 1 |
| App Engine Admin | 0 | 2 |
| Security Command Center | 0 | 3 |
| Organization Policy | 0 | 3 |
| Access Context Manager | 0 | 3 |
| Cloud Asset | 0 | 3 |
| Sensitive Data Protection | 0 | 5 |
| Binary Authorization | 0 | 2 |
| Container Analysis | 0 | 2 |
| Certificate Authority | 0 | 3 |
| Assured Workloads | 0 | 2 |
| Cloud IDS | 0 | 1 |
| Backup and DR | 0 | 3 |
| Web Security Scanner | 0 | 2 |
| Identity-Aware Proxy | 0 | 2 |
| Recommender | 0 | 2 |
| Cloud Billing | 0 | 3 |
| Network Management | 0 | 1 |
| Cloud Run | 0 | 2 |
| Pub/Sub | 0 | 2 |
| Spanner | 0 | 2 |
| Bigtable Admin | 0 | 2 |
| Memorystore Redis | 0 | 1 |
| Filestore | 0 | 1 |
| AlloyDB | 0 | 1 |
| VPC Access | 1 | 1 |

See [EXTERNAL_RESOURCES.md](../reference/EXTERNAL_RESOURCES.md) for compliance benchmarks, open source tools, and cloud provider documentation.
