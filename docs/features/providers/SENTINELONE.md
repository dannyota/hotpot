# SentinelOne

SentinelOne API resource ingestion coverage in the bronze layer.

## 🛡️ Management API (`/web/api/v2.1/`)

### Core Resources

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Accounts | `/accounts` | ✅ |
| Sites | `/sites` | ✅ |
| Groups | `/groups` | ✅ |
| Agents | `/agents` | ✅ |

### Application Management (`/application-management/`)

**Inventory:**

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Application Inventory | `/application-management/inventory` | |
| Endpoint Apps | `/application-management/inventory/applications` | |
| Inventory Endpoints | `/application-management/inventory/endpoints` | |

**Risk & CVEs:**

| Resource | Endpoint | Status |
|----------|----------|:------:|
| CVE Data | `/application-management/risks` | |
| Aggregated App Risk | `/application-management/risks/aggregated-applications` | |
| App Risk | `/application-management/risks/applications` | |
| Application CVEs | `/application-management/risks/cves` | |
| Risk Endpoints | `/application-management/risks/endpoints` | |

### Detection & Response

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Alerts | `/cloud-detection/alerts` | |
| STAR Rules | `/cloud-detection/rules` | |
| Deep Visibility Queries | `/dv/query-status` | |
| IOCs | `/threat-intelligence/iocs` | |

### Policy & Configuration

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Policies | `/policies` | |
| Exclusions | `/exclusions` | |
| Blocklist (Restrictions) | `/restrictions` | |
| Firewall Control | `/firewall-control` | |

### Operations

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Activities (Audit Log) | `/activities` | |
| Remote Scripts | `/remote-scripts` | |

### Identity & Access

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Users | `/users` | |
| RBAC Roles | `/rbac/roles` | |
| API Token Details | `/users/api-token-details` | |
| Service Users | `/service-users` | |

### Network Discovery (`/ranger/`)

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Device Inventory | `/ranger/table-view` | ✅ |
| Gateways | `/ranger/gateways` | ✅ |
| Settings | `/ranger/settings` | ✅ |

### Updates & Packages

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Update Packages | `/update/agent/packages` | |
| Deployment Sites | `/update/agent/deployed-sites` | |

## 📊 Summary

**Total: 7/31 (23%)**

| API | Implemented | Total |
|-----|:-----------:|:-----:|
| Core Resources | 4 | 4 |
| Application Management | 0 | 8 |
| Detection & Response | 0 | 4 |
| Policy & Configuration | 0 | 4 |
| Operations | 0 | 2 |
| Identity & Access | 0 | 4 |
| Network Discovery | 3 | 3 |
| Updates & Packages | 0 | 2 |

See [EXTERNAL_RESOURCES.md](../reference/EXTERNAL_RESOURCES.md) for compliance benchmarks, open source tools, and cloud provider documentation.
