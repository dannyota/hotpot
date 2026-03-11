# DigitalOcean

DigitalOcean API resource ingestion coverage in the bronze layer.

## 🔑 API Token Setup

Create a **custom-scoped** personal access token with read-only permissions.
Do NOT use "Full Access" or "Read Only" (grants all read scopes including
credentials).

### Required scopes

| Scope | Grants |
|-------|--------|
| `account:read` | View account details |
| `database:read` | View managed databases |
| `domain:read` | View domains and records |
| `droplet:read` | View Droplets |
| `firewall:read` | View cloud firewalls |
| `kubernetes:read` | View clusters (no credentials) |
| `load_balancer:read` | View load balancers |
| `project:read` | View projects |
| `block_storage:read` | View volumes |
| `ssh_key:read` | View SSH keys |
| `vpc:read` | View VPCs |

### Do NOT grant

These scopes expose secrets and are not needed for ingestion:

| Scope | Risk |
|-------|------|
| `database:view_credentials` | Exposes database passwords, certificates, and connection URIs |
| `kubernetes:access_cluster` | Generates and downloads kubeconfig with client certificates |

The `database:read` scope returns cluster metadata (engine, version, status,
firewall rules, configs) without connection strings or passwords.
The `kubernetes:read` scope returns cluster metadata (version, HA, node pools)
without kubeconfig credentials.

## 🌐 API v2 (`/v2/`)

### Account & Billing

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Account | `/account` | ✅ |
| Balance | `/customers/my/balance` | |
| Billing History | `/customers/my/billing_history` | |
| Invoices | `/customers/my/invoices` | |
| Billing Insights | `/billing/{account_urn}/insights` | |

### Droplets

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Droplets | `/droplets` | ✅ |
| Droplet Backups | `/droplets/{id}/backups` | |
| Droplet Backup Policies | `/droplets/backups/policies` | |
| Droplet Firewalls | `/droplets/{id}/firewalls` | |
| Droplet Kernels | `/droplets/{id}/kernels` | |
| Droplet Neighbors | `/droplets/{id}/neighbors` | |
| Droplet Snapshots | `/droplets/{id}/snapshots` | |
| Droplet Autoscale Pools | `/droplets/autoscale` | |
| Droplet Autoscale Members | `/droplets/autoscale/{id}/members` | |
| Droplet Autoscale History | `/droplets/autoscale/{id}/history` | |

### Kubernetes

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Clusters | `/kubernetes/clusters` | ✅ |
| Node Pools | `/kubernetes/clusters/{id}/node_pools` | ✅ |
| Cluster Credentials | `/kubernetes/clusters/{id}/credentials` | ⚠️ |
| Cluster Upgrades | `/kubernetes/clusters/{id}/upgrades` | |
| Cluster Lint Results | `/kubernetes/clusters/{id}/clusterlint` | |
| Options | `/kubernetes/options` | |

### App Platform

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Apps | `/apps` | |
| App Deployments | `/apps/{id}/deployments` | |
| App Alerts | `/apps/{id}/alerts` | |
| App Instances | `/apps/{id}/instances` | |
| App Health | `/apps/{id}/health` | |
| App Regions | `/apps/regions` | |
| App Instance Sizes | `/apps/tiers/instance_sizes` | |

### Functions

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Namespaces | `/functions/namespaces` | |
| Triggers | `/functions/namespaces/{id}/triggers` | |

### Databases

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Database Clusters | `/databases` | ✅ |
| Databases | `/databases/{id}/dbs` | |
| Database Users | `/databases/{id}/users` | ✅ |
| Database Replicas | `/databases/{id}/replicas` | ✅ |
| Database Connection Pools | `/databases/{id}/pools` | ✅ |
| Database Backups | `/databases/{id}/backups` | ✅ |
| Database Config | `/databases/{id}/config` | ✅ |
| Database Firewall Rules | `/databases/{id}/firewall` | ✅ |
| Database Eviction Policy | `/databases/{id}/eviction_policy` | |
| Database SQL Mode | `/databases/{id}/sql_mode` | |
| Database Topics (Kafka) | `/databases/{id}/topics` | |
| Database Indexes | `/databases/{id}/indexes` | |
| Database Log Sinks | `/databases/{id}/logsink` | |
| Database Autoscale | `/databases/{id}/autoscale` | |
| Database Events | `/databases/{id}/events` | |
| Database Options | `/databases/options` | |
| Database Metrics Credentials | `/databases/metrics/credentials` | |

### Block Storage

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Volumes | `/volumes` | ✅ |
| Volume Snapshots | `/volumes/{id}/snapshots` | |

### NFS

| Resource | Endpoint | Status |
|----------|----------|:------:|
| NFS Shares | `/nfs` | |
| NFS Snapshots | `/nfs/snapshots` | |

### Spaces

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Spaces Keys | `/spaces/keys` | |

### Snapshots

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Snapshots | `/snapshots` | |

### Networking

| Resource | Endpoint | Status |
|----------|----------|:------:|
| VPCs | `/vpcs` | ✅ |
| VPC Members | `/vpcs/{id}/members` | |
| VPC Peerings | `/vpc_peerings` | |
| VPC NAT Gateways | `/vpc_nat_gateways` | |
| Domains | `/domains` | ✅ |
| Domain Records | `/domains/{name}/records` | ✅ |
| Firewalls | `/firewalls` | ✅ |
| Load Balancers | `/load_balancers` | ✅ |
| Reserved IPs | `/reserved_ips` | |
| Reserved IPv6 | `/reserved_ipv6` | |
| Floating IPs | `/floating_ips` | |
| CDN Endpoints | `/cdn/endpoints` | |
| Certificates | `/certificates` | |
| BYOIP Prefixes | `/byoip_prefixes` | |
| Partner Network Connect | `/partner_network_connect/attachments` | |

### Container Registry

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Registries | `/registries` | |
| Registry Repositories | `/registries/{name}/repositoriesV2` | |
| Repository Digests | `/registries/{name}/repositories/{repo}/digests` | |
| Repository Tags | `/registries/{name}/repositories/{repo}/tags` | |
| Registry Garbage Collections | `/registries/{name}/garbage-collections` | |
| Registry Subscription | `/registries/subscription` | |
| Registry Options | `/registries/options` | |

### Images

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Images | `/images` | |

### Projects

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Projects | `/projects` | ✅ |
| Project Resources | `/projects/{id}/resources` | ✅ |

### Monitoring

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Alert Policies | `/monitoring/alerts` | |
| Monitoring Sinks | `/monitoring/sinks` | |
| Monitoring Sink Destinations | `/monitoring/sinks/destinations` | |

### Uptime

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Uptime Checks | `/uptime/checks` | |
| Uptime Check State | `/uptime/checks/{id}/state` | |
| Uptime Alerts | `/uptime/checks/{id}/alerts` | |

### SSH Keys

| Resource | Endpoint | Status |
|----------|----------|:------:|
| SSH Keys | `/account/keys` | ✅ |

### Tags

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Tags | `/tags` | |

### Actions

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Actions | `/actions` | |

### Sizes & Regions

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Sizes | `/sizes` | |
| Regions | `/regions` | |

### 1-Click Applications

| Resource | Endpoint | Status |
|----------|----------|:------:|
| 1-Click Apps | `/1-clicks` | |

### Add-Ons

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Add-On Apps | `/add-ons/apps` | |
| SaaS Add-Ons | `/add-ons/saas` | |

### GradientAI Platform

| Resource | Endpoint | Status |
|----------|----------|:------:|
| AI Agents | `/gen-ai/agents` | |
| AI Knowledge Bases | `/gen-ai/knowledge_bases` | |
| AI Models | `/gen-ai/models` | |
| AI Model API Keys | `/gen-ai/models/api_keys` | |
| AI Workspaces | `/gen-ai/workspaces` | |
| AI Indexing Jobs | `/gen-ai/indexing_jobs` | |
| AI Evaluation Test Cases | `/gen-ai/evaluation_test_cases` | |
| AI Evaluation Runs | `/gen-ai/evaluation_runs` | |
| AI Evaluation Metrics | `/gen-ai/evaluation_metrics` | |
| AI Regions | `/gen-ai/regions` | |
| OpenAI Keys | `/gen-ai/openai/keys` | |
| Anthropic Keys | `/gen-ai/anthropic/keys` | |

## 🗺️ Ingestion Roadmap

No CIS benchmark exists for DigitalOcean. Phases are ordered by security posture
value — prioritizing data protection, access control, and encryption visibility
before operational/inventory resources.

### Phase 1 — Databases (data protection)

Managed databases are the highest-risk surface: public access, authentication,
encryption, and backup configuration all live here.

| Resource | Endpoint | Parent | Security Signal |
|----------|----------|--------|-----------------|
| Database Clusters | `/databases` | — | Engine, version, public access, region, encryption |
| Database Firewall Rules | `/databases/{id}/firewall` | Cluster | Trusted sources allowlist |
| Database Users | `/databases/{id}/users` | Cluster | Authentication, role grants |
| Database Replicas | `/databases/{id}/replicas` | Cluster | Replication targets, cross-region exposure |
| Database Backups | `/databases/{id}/backups` | Cluster | Backup retention, recovery readiness |
| Database Config | `/databases/{id}/config` | Cluster | Engine-specific security settings |
| Database Connection Pools | `/databases/{id}/pools` | Cluster | Connection management, user mapping |

Deferred (low security value or engine-specific):
Databases (logical DBs), Eviction Policy, SQL Mode, Topics (Kafka), Indexes,
Log Sinks, Autoscale, Events, Options, Metrics Credentials.

### Phase 2 — Kubernetes (container security)

DOKS clusters are high-value: node pool sizing, version currency, and cluster
configuration control blast radius.

| Resource | Endpoint | Parent | Security Signal |
|----------|----------|--------|-----------------|
| Clusters | `/kubernetes/clusters` | — | Version, HA, surge upgrades, auto-upgrade |
| Node Pools | `/kubernetes/clusters/{id}/node_pools` | Cluster | Size, count, taints, labels |
| Cluster Credentials | `/kubernetes/clusters/{id}/credentials` | Cluster | Certificate expiry, token rotation |

Credentials skipped: `GetCredentials` *generates* new credentials (takes
`ExpirySeconds`), making it unsafe for read-only ingestion. Key security signals
(version, HA, auto-upgrade, control plane firewall) are on the cluster object.

Deferred: Upgrades (point-in-time), Lint Results (on-demand), Options (static).

### Phase 3 — Networking gaps (access control)

Fill remaining network visibility gaps — IP allocation, TLS posture, VPC topology.

| Resource | Endpoint | Parent | Security Signal |
|----------|----------|--------|-----------------|
| Certificates | `/certificates` | — | Expiration, type (custom vs Let's Encrypt) |
| Reserved IPs | `/reserved_ips` | — | Allocation, droplet attachment |
| VPC Members | `/vpcs/{id}/members` | VPC | Resource isolation boundaries |
| VPC Peerings | `/vpc_peerings` | — | Cross-VPC connectivity |

Deferred: Reserved IPv6, Floating IPs (legacy), CDN Endpoints, BYOIP Prefixes,
Partner Network Connect, NAT Gateways.

### Phase 4 — Container Registry (supply chain)

Image provenance and access control for private registries.

| Resource | Endpoint | Parent | Security Signal |
|----------|----------|--------|-----------------|
| Registries | `/registries` | — | Storage usage, region, subscription tier |
| Registry Repositories | `/registries/{name}/repositoriesV2` | Registry | Image inventory |
| Repository Tags | `/registries/{name}/repositories/{repo}/tags` | Repository | Tag mutability, latest versions |

Deferred: Digests (detail level), Garbage Collections (ops), Subscription (billing),
Options (static).

### Phase 5 — Data snapshots (backup & recovery)

Snapshots and backup policies for disaster recovery posture.

| Resource | Endpoint | Parent | Security Signal |
|----------|----------|--------|-----------------|
| Snapshots | `/snapshots` | — | Type, size, region, age |
| Volume Snapshots | `/volumes/{id}/snapshots` | Volume | Backup coverage for block storage |
| Droplet Backups | `/droplets/{id}/backups` | Droplet | Backup existence and recency |
| Droplet Backup Policies | `/droplets/backups/policies` | — | Automatic backup scheduling |

### Phase 6 — Monitoring & observability (detection gaps)

Alerting coverage determines whether security events are noticed.

| Resource | Endpoint | Parent | Security Signal |
|----------|----------|--------|-----------------|
| Alert Policies | `/monitoring/alerts` | — | What's monitored, thresholds |
| Uptime Checks | `/uptime/checks` | — | Availability monitoring targets |
| Uptime Alerts | `/uptime/checks/{id}/alerts` | Check | Notification configuration |

Deferred: Uptime Check State (ephemeral), Monitoring Sinks/Destinations.

### Phase 7 — Images & metadata (inventory)

Foundational inventory for drift detection and governance.

| Resource | Endpoint | Parent | Security Signal |
|----------|----------|--------|-----------------|
| Images | `/images` | — | Custom images, public flag, distribution |
| Tags | `/tags` | — | Governance labeling, resource counts |
| Spaces Keys | `/spaces/keys` | — | S3-compatible access credentials |

### Phase 8 — App Platform & Functions (compute inventory)

Lower priority — these are PaaS/FaaS surfaces with less configurable security.

| Resource | Endpoint | Parent | Security Signal |
|----------|----------|--------|-----------------|
| Apps | `/apps` | — | Deployment config, domains, env vars |
| App Deployments | `/apps/{id}/deployments` | App | Deployment history |
| Namespaces | `/functions/namespaces` | — | Serverless function inventory |
| Triggers | `/functions/namespaces/{id}/triggers` | Namespace | Invocation sources |

Deferred: App Alerts, Instances, Health, Regions, Instance Sizes.

### Not planned

Low security value — billing, reference data, marketplace, and AI platform
resources. Revisit if compliance requirements change.

- Account & Billing: Balance, Billing History, Invoices, Billing Insights
- Droplets: Kernels, Neighbors, Autoscale (Pools/Members/History)
- NFS: Shares, Snapshots
- Actions (audit log — high volume, better via event stream)
- Sizes, Regions (static reference data)
- 1-Click Applications, Add-Ons, GradientAI Platform

## 📊 Summary

**Total: 20/113 (18%)**

| API | Implemented | Total |
|-----|:-----------:|:-----:|
| Account & Billing | 1 | 5 |
| Droplets | 1 | 10 |
| Kubernetes | 2 | 6 |
| App Platform | 0 | 7 |
| Functions | 0 | 2 |
| Databases | 7 | 17 |
| Block Storage | 1 | 2 |
| NFS | 0 | 2 |
| Spaces | 0 | 1 |
| Snapshots | 0 | 1 |
| Networking | 5 | 15 |
| Container Registry | 0 | 7 |
| Images | 0 | 1 |
| Projects | 2 | 2 |
| Monitoring | 0 | 3 |
| Uptime | 0 | 3 |
| SSH Keys | 1 | 1 |
| Tags | 0 | 1 |
| Actions | 0 | 1 |
| Sizes & Regions | 0 | 2 |
| 1-Click Applications | 0 | 1 |
| Add-Ons | 0 | 2 |
| GradientAI Platform | 0 | 12 |

See [EXTERNAL_RESOURCES.md](../reference/EXTERNAL_RESOURCES.md) for compliance benchmarks, open source tools, and cloud provider documentation.
