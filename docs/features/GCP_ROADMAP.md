# GCP Roadmap

Prioritized implementation strategy for remaining GCP resources (57 of 117).

See [GCP.md](./GCP.md) for current coverage details.

## Strategy

Prioritization criteria:

| Factor | Description |
|--------|-------------|
| Security value | Direct security insight for customers |
| Dependencies | Org hierarchy before org-level policies |
| Completion | Finish started services (low effort, shared infra) |
| Usage frequency | Commonly deployed GCP services first |

## Phase 1: Foundation ✅

Complete the org hierarchy and finish partially-done services. Resource Manager provides the org/folder/project tree that org-level security services depend on. Partial services are low effort since client infrastructure already exists.

| Service | Resource | Priority | Status |
|---------|----------|:--------:|:------:|
| **Resource Manager** | Organizations | P0 | ✅ |
| | Folders | P0 | ✅ |
| | Organization IAM Policies | P0 | ✅ |
| | Folder IAM Policies | P0 | ✅ |
| | Project IAM Policies | P0 | ✅ |
| **Logging** | Log Metrics | P1 | ✅ |
| | Log Exclusions | P1 | ✅ |
| **DNS** | DNS Policies | P1 | ✅ |
| **Cloud Storage** | Bucket IAM Policies | P1 | ✅ |
| **Compute** | Interconnects | P2 | ✅ |
| | Packet Mirrorings | P2 | ✅ |
| | Project Metadata | P2 | ✅ |

**Resources:** 12/12
**Depends on:** nothing
**Unlocks:** Phase 2 (org-level security services)

## Phase 2: Security & Governance

Core security services — the product's primary value. SCC findings, org policies, VPC Service Controls, container supply chain, and asset inventory.

| Service | Resource | Priority | Notes |
|---------|----------|:--------:|-------|
| **Security Command Center** | Sources | P0 | ✅ |
| | Findings | P0 | ✅ |
| | Notification Configs | P1 | Alert routing |
| **Organization Policy** | Constraints | P0 | ✅ |
| | Org Policies | P0 | ✅ |
| | Custom Constraints | P1 | Customer-defined rules |
| **Access Context Manager** | Access Policies | P0 | VPC Service Controls root |
| | Access Levels | P0 | Zero-trust conditions |
| | Service Perimeters | P0 | Data exfiltration boundaries |
| **Cloud Asset** | Assets | P1 | Cross-resource inventory |
| | IAM Policy Search | P1 | Who has access to what |
| | Resource Search | P2 | Resource discovery |
| **Binary Authorization** | Policy | P1 | Container deploy policy |
| | Attestors | P1 | Trusted signers |
| **Container Analysis** | Notes | P1 | Vulnerability sources |
| | Occurrences | P1 | Actual vulns in images |
| **Identity-Aware Proxy** | IAP Settings | P2 | App-level access control |
| | IAP IAM Policies | P2 | Who can access what app |

**Resources:** 4/18
**Depends on:** Phase 1 (org hierarchy for org-level services)
**Unlocks:** Compliance rules in gold layer

## Phase 3: Application Platform

Widely-used runtime services. Cloud Run and Functions are commonly audited serverless targets. Monitoring config detects missing alerting.

| Service | Resource | Priority | Notes |
|---------|----------|:--------:|-------|
| **Cloud Run** | Services | P0 | Top serverless platform |
| | Revisions | P1 | Service versions |
| **Cloud Functions** | Functions | P0 | Serverless compute |
| **Pub/Sub** | Topics | P1 | Messaging backbone |
| | Subscriptions | P1 | Message consumers |
| **Monitoring** | Alert Policies | P0 | Detect missing alerts |
| | Uptime Check Configs | P1 | Availability monitoring |
| **App Engine** | Applications | P2 | Legacy app platform |
| | Services | P2 | App Engine services |
| **Service Usage** | Enabled Services | P2 | API surface area |

**Resources:** 10
**Depends on:** nothing (can parallelize with Phase 2)

## Phase 4: Data Services

Data stores hold sensitive information. BigQuery is highest priority — widely used, often misconfigured.

| Service | Resource | Priority | Notes |
|---------|----------|:--------:|-------|
| **BigQuery** | Datasets | P0 | Most common data warehouse |
| | Tables | P1 | Table-level access |
| **Spanner** | Instances | P1 | Global databases |
| | Databases | P1 | Database configs |
| **Bigtable** | Instances | P2 | Wide-column store |
| | Clusters | P2 | Bigtable infra |
| **Dataproc** | Clusters | P2 | Managed Spark/Hadoop |
| **Memorystore Redis** | Instances | P2 | In-memory cache |
| **Filestore** | Instances | P2 | Managed NFS |
| **AlloyDB** | Clusters | P2 | PostgreSQL-compatible |

**Resources:** 10
**Depends on:** nothing (can parallelize with Phase 2-3)

## Phase 5: Compliance & Operations

Specialized compliance, operational, and niche services. Lower urgency but needed for full coverage.

| Service | Resource | Priority | Notes |
|---------|----------|:--------:|-------|
| **Sensitive Data Protection** | Inspect Templates | P1 | DLP scan configs |
| | Deidentify Templates | P1 | Data masking |
| | Job Triggers | P2 | Scheduled scans |
| | DLP Jobs | P2 | Scan results |
| | Discovery Configs | P2 | Auto-discovery |
| **Certificate Authority** | CA Pools | P1 | PKI infrastructure |
| | Certificate Authorities | P1 | CA instances |
| | Certificates | P2 | Issued certs |
| **Backup and DR** | Backup Vaults | P1 | Backup storage |
| | Backup Plans | P1 | Backup schedules |
| | Backup Plan Associations | P2 | What's backed up |
| **Cloud Billing** | Billing Accounts | P2 | Billing hierarchy |
| | Project Billing Info | P2 | Cost attribution |
| | Budgets | P2 | Spend alerts |
| **Assured Workloads** | Workloads | P2 | Compliance environments |
| | Violations | P2 | Compliance violations |
| **Access Approval** | Settings | P2 | Approval config |
| | Approval Requests | P2 | Access requests |
| **Web Security Scanner** | Scan Configs | P2 | App scanning setup |
| | Scan Runs | P2 | Scan results |
| **Recommender** | IAM Recommendations | P2 | Least privilege |
| | IAM Insights | P2 | Over-provisioned access |
| **Cloud IDS** | Endpoints | P2 | Network threat detection |
| **Network Management** | Connectivity Tests | P2 | Network path analysis |
| **API Keys** | API Keys | P2 | Key management |
| **Essential Contacts** | Contacts | P2 | Notification routing |

**Resources:** 27
**Depends on:** nothing

## Execution Order

```
Phase 1 ──→ Phase 2 ──→ ┐
                         ├──→ Phase 5
Phase 3 ────────────────→┤
                         │
Phase 4 ────────────────→┘
```

- Phase 1 → Phase 2 is sequential (org hierarchy required)
- Phases 3 and 4 can run in parallel with Phase 2
- Phase 5 after all others

## Summary

| Phase | Focus | Resources | Key Services |
|-------|-------|:---------:|-------------|
| 1 | Foundation | 12 | Resource Manager, finish partial |
| 2 | Security | 18 | SCC, OrgPolicy, ACM, CloudAsset |
| 3 | Platform | 10 | Cloud Run, Functions, Pub/Sub |
| 4 | Data | 10 | BigQuery, Spanner, Bigtable |
| 5 | Compliance | 27 | DLP, CertAuth, Backup, Billing |
| **Total** | | **77** | |
