# GCP Roadmap

Prioritized implementation strategy for remaining GCP resources (23 of 117).

See [GCP.md](./GCP.md) for current coverage details.

## Completed Batch: Phases 2+3+4 ✅ (34 resources, 20 services)

All 34 resources across 20 services implemented in parallel. Full build passes.

### Execution Steps

1. **Schemas** — 20 agents create bronze + bronze history ent schemas in parallel ✅
2. **Generate** — Single `go generate` after all schemas land ✅
3. **Service layer** — 20 agents create service files (client, converter, diff, history, service, activities, workflows, register) in parallel ✅
4. **Wire up** — Add new Go module dependencies, update parent register/workflows, build and fix ✅
5. **Docs** — Update GCP.md counts and roadmap status ✅

### Agent Assignments

| # | Service | Resources | Phase |
|---|---------|:---------:|:-----:|
| 1 | SCC Notification Configs | 1 | 2 |
| 2 | OrgPolicy Custom Constraints | 1 | 2 |
| 3 | Access Context Manager | 3 | 2 |
| 4 | Cloud Asset | 3 | 2 |
| 5 | Binary Authorization | 2 | 2 |
| 6 | Container Analysis | 2 | 2 |
| 7 | Identity-Aware Proxy | 2 | 2 |
| 8 | Cloud Run | 2 | 3 |
| 9 | Cloud Functions | 1 | 3 |
| 10 | Pub/Sub | 2 | 3 |
| 11 | Monitoring | 2 | 3 |
| 12 | App Engine | 2 | 3 |
| 13 | Service Usage | 1 | 3 |
| 14 | BigQuery | 2 | 4 |
| 15 | Spanner | 2 | 4 |
| 16 | Bigtable | 2 | 4 |
| 17 | Dataproc | 1 | 4 |
| 18 | Memorystore Redis | 1 | 4 |
| 19 | Filestore | 1 | 4 |
| 20 | AlloyDB | 1 | 4 |

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

## Phase 2: Security & Governance ✅

Core security services — the product's primary value. SCC findings, org policies, VPC Service Controls, container supply chain, and asset inventory.

| Service | Resource | Priority | Status |
|---------|----------|:--------:|:------:|
| **Security Command Center** | Sources | P0 | ✅ |
| | Findings | P0 | ✅ |
| | Notification Configs | P1 | ✅ |
| **Organization Policy** | Constraints | P0 | ✅ |
| | Org Policies | P0 | ✅ |
| | Custom Constraints | P1 | ✅ |
| **Access Context Manager** | Access Policies | P0 | ✅ |
| | Access Levels | P0 | ✅ |
| | Service Perimeters | P0 | ✅ |
| **Cloud Asset** | Assets | P1 | ✅ |
| | IAM Policy Search | P1 | ✅ |
| | Resource Search | P2 | ✅ |
| **Binary Authorization** | Policy | P1 | ✅ |
| | Attestors | P1 | ✅ |
| **Container Analysis** | Notes | P1 | ✅ |
| | Occurrences | P1 | ✅ |
| **Identity-Aware Proxy** | IAP Settings | P2 | ✅ |
| | IAP IAM Policies | P2 | ✅ |

**Resources:** 18/18
**Depends on:** Phase 1 (org hierarchy for org-level services)
**Unlocks:** Compliance rules in gold layer

## Phase 3: Application Platform ✅

Widely-used runtime services. Cloud Run and Functions are commonly audited serverless targets. Monitoring config detects missing alerting.

| Service | Resource | Priority | Status |
|---------|----------|:--------:|:------:|
| **Cloud Run** | Services | P0 | ✅ |
| | Revisions | P1 | ✅ |
| **Cloud Functions** | Functions | P0 | ✅ |
| **Pub/Sub** | Topics | P1 | ✅ |
| | Subscriptions | P1 | ✅ |
| **Monitoring** | Alert Policies | P0 | ✅ |
| | Uptime Check Configs | P1 | ✅ |
| **App Engine** | Applications | P2 | ✅ |
| | Services | P2 | ✅ |
| **Service Usage** | Enabled Services | P2 | ✅ |

**Resources:** 10/10
**Depends on:** nothing (can parallelize with Phase 2)

## Phase 4: Data Services ✅

Data stores hold sensitive information. BigQuery is highest priority — widely used, often misconfigured.

| Service | Resource | Priority | Status |
|---------|----------|:--------:|:------:|
| **BigQuery** | Datasets | P0 | ✅ |
| | Tables | P1 | ✅ |
| **Spanner** | Instances | P1 | ✅ |
| | Databases | P1 | ✅ |
| **Bigtable** | Instances | P2 | ✅ |
| | Clusters | P2 | ✅ |
| **Dataproc** | Clusters | P2 | ✅ |
| **Memorystore Redis** | Instances | P2 | ✅ |
| **Filestore** | Instances | P2 | ✅ |
| **AlloyDB** | Clusters | P2 | ✅ |

**Resources:** 10/10
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
| 1 | Foundation | 12 ✅ | Resource Manager, finish partial |
| 2 | Security | 18 ✅ | SCC, OrgPolicy, ACM, CloudAsset |
| 3 | Platform | 10 ✅ | Cloud Run, Functions, Pub/Sub |
| 4 | Data | 10 ✅ | BigQuery, Spanner, Bigtable |
| 5 | Compliance | 27 | DLP, CertAuth, Backup, Billing |
| **Total** | | **77** | |
