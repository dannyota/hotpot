# VNGCloud

VNGCloud resource ingestion coverage in the bronze layer.

## Auth

OAuth2 PKCE flow simulating browser login (no API token mechanism). Not in the OpenAPI spec.

| Setting | Value |
|---------|-------|
| Flow | PKCE authorization code |
| Credentials | Root email + IAM username + password |
| 2FA | Optional TOTP |

## API Gateways

| Gateway | Base URL Pattern | Scope |
|---------|-----------------|-------|
| vServer | `https://{region}.console.vngcloud.vn/vserver/iam-vserver-gateway` | Compute, storage, networking |
| vNetwork | `https://{region}-vnetwork.console.vngcloud.vn/vnetwork-gateway/vnetwork` | VPC endpoints, peering |

Resources scope to **project + region** (`hcm-3`, `han-1`, etc.).

## Pagination

| Gateway | Style | Wrapper |
|---------|-------|---------|
| vServer (most) | `?page=N&size=50` | `{listData, page, pageSize, totalPage, totalItem}` |
| vServer (subnets) | None (returns array) | Direct `[]SubnetDto` |
| vServer (projects) | None | `{success, projects}` |
| vNetwork (endpoints) | JSON `params` query string | `{success, data, page, size, totalPage, total}` |

## vServer API v1 (`/v1/`)

### Projects

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Projects | `GET /v1/projects` | |

### Zones & Regions

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Zones | `GET /v1/{projectId}/zones` | |
| Regions (v2) | `GET /v2/{projectId}/region` | |

### Flavors

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Flavors | `GET /v1/{projectId}/{flavorZoneId}/flavors` | |
| Custom Flavors | `GET /v1/{projectId}/flavors/customs` | |
| Flavor Zones | `GET /v1/{projectId}/flavor_zones/product` | |

### Images

| Resource | Endpoint | Status |
|----------|----------|:------:|
| OS Images | `GET /v1/{projectId}/images/os` | |
| GPU Images | `GET /v1/{projectId}/images/gpu` | |

### Volume Types

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Volume Types | `GET /v1/{projectId}/volume_types` | |
| Volume Type Zones | `GET /v1/{projectId}/volume_type_zones` | |

## vServer API v2 (`/v2/`)

### Compute

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Servers | `GET /v2/{projectId}/servers` | |
| Server Groups | `GET /v2/{projectId}/serverGroups` | |
| Server Group Policies | `GET /v2/{projectId}/serverGroups/policies` | |
| SSH Keys | `GET /v2/{projectId}/sshKeys` | |
| User Images | `GET /v2/{projectId}/user-images` | |

### Networking

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Networks (VPCs) | `GET /v2/{projectId}/networks` | |
| Subnets | `GET /v2/{projectId}/networks/{networkId}/subnets` | |
| Security Groups | `GET /v2/{projectId}/secgroups` | |
| Security Group Rules | `GET /v2/{projectId}/secgroups/{secgroupId}/secGroupRules` | |
| Network ACLs | `GET /v2/{projectId}/network-acl/list` | |
| Network ACL Rules | `GET /v2/{projectId}/network-acl/{uuid}/rules` | |
| Route Tables | `GET /v2/{projectId}/route-table` | |
| Route Table Routes | `GET /v2/{projectId}/route-table/route/{routeTableId}` | |
| DHCP Options | `GET /v2/{projectId}/dhcp_option` | |
| Elastic IPs | `GET /v2/{projectId}/elastic-ips` | |
| Network Interfaces (Elastic) | `GET /v2/{projectId}/network-interfaces-elastic` | |
| Virtual IPs | `GET /v2/{projectId}/virtualIpAddress` | |
| Public VIPs | `GET /v2/{projectId}/public-vips/externalNetworkInterfaces` | |
| Peering | `GET /v2/{projectId}/peering` | |
| Interconnects | `GET /v2/{projectId}/interconnects` | |
| Interconnect Connections | `GET /v2/{projectId}/interconnects/{id}/connections` | |
| WAN IPs | `GET /v2/{projectId}/wanIps` | |

### Block Storage

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Volumes | `GET /v2/{projectId}/volumes` | |
| Persistent Volumes | `GET /v2/{projectId}/persistent-volumes` | |

### Tags & Quotas

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Tags | `GET /v2/{projectId}/tag` | |
| Quotas | `GET /v2/{projectId}/quotas/quotaUsed` | |

### Protocols

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Protocols | `GET /v2/protocols` | |

## vNetwork API (`/vnetwork/v1/`)

Not in the OpenAPI spec.

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Regions | `GET /vnetwork/v1/regions` | |
| VPC Endpoints | `GET /vnetwork/v1/{regionId}/{projectId}/endpoints` | |

## OpenAPI Spec vs Console API

The OpenAPI spec documents the official API. Ingestion uses the web console backend (`console.vngcloud.vn/vserver/iam-vserver-gateway`), which may return different response shapes. Differences to watch for during implementation:

| Resource | Field | OpenAPI Spec | Console API (observed) |
|----------|-------|-------------|----------------------|
| Network | `elasticIps` | `[]ElasticOfNetworkDto` | May be `[]string` |
| Subnet | `secondarySubnets` | `[]SecondarySubnetDto` | May be `[]string` |
| Server.Flavor | `cpu`, `memory` | `number(double)` | Integer values |
| ServerGroup | `servers` | Full `Server` objects | Subset of fields |
| Project | `errorCode` | `integer(int32)` | `string` |

## Summary

**Total: 0/39 (0%)**

| Category | Implemented | Total |
|----------|:-----------:|:-----:|
| Projects | 0 | 1 |
| Zones & Regions | 0 | 2 |
| Flavors | 0 | 3 |
| Images | 0 | 2 |
| Volume Types | 0 | 2 |
| Compute | 0 | 5 |
| Networking | 0 | 17 |
| Block Storage | 0 | 2 |
| Tags & Quotas | 0 | 2 |
| Protocols | 0 | 1 |
| vNetwork | 0 | 2 |

See [EXTERNAL_RESOURCES.md](../reference/EXTERNAL_RESOURCES.md) for compliance benchmarks, open source tools, and cloud provider documentation.
