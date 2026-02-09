# DigitalOcean

DigitalOcean API resource ingestion coverage in the bronze layer.

## 🌐 API v2 (`/v2/`)

### Account & Billing

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Account | `/account` | |
| Balance | `/customers/my/balance` | |
| Billing History | `/customers/my/billing_history` | |
| Invoices | `/customers/my/invoices` | |
| Billing Insights | `/billing/{account_urn}/insights` | |

### Droplets

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Droplets | `/droplets` | |
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
| Clusters | `/kubernetes/clusters` | |
| Node Pools | `/kubernetes/clusters/{id}/node_pools` | |
| Cluster Credentials | `/kubernetes/clusters/{id}/credentials` | |
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
| Database Clusters | `/databases` | |
| Databases | `/databases/{id}/dbs` | |
| Database Users | `/databases/{id}/users` | |
| Database Replicas | `/databases/{id}/replicas` | |
| Database Connection Pools | `/databases/{id}/pools` | |
| Database Backups | `/databases/{id}/backups` | |
| Database Config | `/databases/{id}/config` | |
| Database Firewall Rules | `/databases/{id}/firewall` | |
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
| Volumes | `/volumes` | |
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
| VPCs | `/vpcs` | |
| VPC Members | `/vpcs/{id}/members` | |
| VPC Peerings | `/vpc_peerings` | |
| VPC NAT Gateways | `/vpc_nat_gateways` | |
| Domains | `/domains` | |
| Domain Records | `/domains/{name}/records` | |
| Firewalls | `/firewalls` | |
| Load Balancers | `/load_balancers` | |
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
| Projects | `/projects` | |
| Project Resources | `/projects/{id}/resources` | |

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
| SSH Keys | `/account/keys` | |

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

## 📊 Summary

**Total: 0/113 (0%)**

| API | Implemented | Total |
|-----|:-----------:|:-----:|
| Account & Billing | 0 | 5 |
| Droplets | 0 | 10 |
| Kubernetes | 0 | 6 |
| App Platform | 0 | 7 |
| Functions | 0 | 2 |
| Databases | 0 | 17 |
| Block Storage | 0 | 2 |
| NFS | 0 | 2 |
| Spaces | 0 | 1 |
| Snapshots | 0 | 1 |
| Networking | 0 | 15 |
| Container Registry | 0 | 7 |
| Images | 0 | 1 |
| Projects | 0 | 2 |
| Monitoring | 0 | 3 |
| Uptime | 0 | 3 |
| SSH Keys | 0 | 1 |
| Tags | 0 | 1 |
| Actions | 0 | 1 |
| Sizes & Regions | 0 | 2 |
| 1-Click Applications | 0 | 1 |
| Add-Ons | 0 | 2 |
| GradientAI Platform | 0 | 12 |

See [EXTERNAL_RESOURCES.md](../reference/EXTERNAL_RESOURCES.md) for compliance benchmarks, open source tools, and cloud provider documentation.
