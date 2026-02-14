# VNG Cloud

VNG Cloud resource ingestion coverage in the bronze layer.

VNG Cloud wraps OpenStack services behind proprietary API gateways. No standard OpenStack APIs are exposed — all endpoints use custom REST with VNG Cloud-specific conventions.

## Auth

Two authentication flows available:

| Flow | Use Case | Endpoint |
|------|----------|----------|
| OAuth2 PKCE | Browser/console simulation | `signin.vngcloud.vn/ap/auth/iam/login` |
| Service Account | SDK/Terraform (recommended) | `iamapis.vngcloud.vn/accounts-api/v2/auth/token` |

Service Account uses `client_id` + `client_secret` (created via IAM console), returns a bearer token valid for 7200s.

## API Gateways

| Gateway | Console URL | SDK URL | OpenStack Analog |
|---------|-------------|---------|------------------|
| vServer | `{region}.console.vngcloud.vn/vserver/iam-vserver-gateway` | `{region}.api.vngcloud.vn/vserver/vserver-gateway` | Nova + Neutron + Cinder + Glance |
| vNetwork | `{region}-vnetwork.console.vngcloud.vn/vnetwork-gateway/vnetwork` | — | Neutron (extended) |
| vLB | `{region}.console.vngcloud.vn/vserver/iam-vlb-gateway` | `{region}.api.vngcloud.vn/vserver/vlb-gateway` | Octavia |
| vKS | `{region}.console.vngcloud.vn/vserver/iam-vserver-gateway` | `vks.api.vngcloud.vn` | Magnum |
| vDB | — | `vdb-gateway.vngcloud.vn` | Trove |
| vStorage | — | (Swift/S3 compatible) | Swift + S3 |
| vMonitor | `vmonitor.console.vngcloud.vn/vmonitor-api` | — | Ceilometer |
| IAM | — | `iamapis.vngcloud.vn/accounts-api` | Keystone |

Regions: `hcm-3` (Ho Chi Minh City), `han-1` (Hanoi). All resources scoped to `{region}` + `{projectId}`.

## Pagination

| Gateway | Style | Wrapper |
|---------|-------|---------|
| vServer (most) | `?page=N&size=50` | `{listData, page, pageSize, totalPage, totalItem}` |
| vServer (subnets) | None (returns array) | Direct `[]SubnetDto` |
| vServer (projects) | None | `{success, projects}` |
| vNetwork (endpoints) | JSON `params` query string | `{success, data, page, size, totalPage, total}` |
| vMonitor | `?page=N&size=50` | `{lstData, page, pageSize, totalPage, totalItem}` (note: `lstData` not `listData`) |

## 🖥️ vServer — Compute (`iam-vserver-gateway`)

### Compute

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Projects | `GET /v1/projects` | Keystone Projects | |
| Zones | `GET /v1/{projectId}/zones` | Nova AZs | |
| Regions | `GET /v2/{projectId}/region` | Keystone Regions | |
| Servers | `GET /v2/{projectId}/servers` | Nova Servers | |
| Server Groups | `GET /v2/{projectId}/serverGroups` | Nova Server Groups | |
| Server Group Policies | `GET /v2/{projectId}/serverGroups/policies` | Nova SG Policies | |
| SSH Keys | `GET /v2/{projectId}/sshKeys` | Nova Keypairs | |
| OS Images | `GET /v1/{projectId}/images/os` | Glance Images | |
| GPU Images | `GET /v1/{projectId}/images/gpu` | Glance Images | |
| User Images | `GET /v2/{projectId}/user-images` | Glance Images | |
| Flavors | `GET /v1/{projectId}/flavors/families/{family}/platforms/{code}` | Nova Flavors | |
| Flavor Zones | `GET /v1/{projectId}/flavor_zones/product` | Nova AZ + Flavors | |
| Flavor Families | `GET /v1/{projectId}/flavor_zones/families` | — | |
| Quotas | `GET /v2/{projectId}/quotas/quotaUsed` | Nova Quotas | |
| Tags | `GET /v2/{projectId}/tag` | — | |
| Tag Keys | `GET /v2/{projectId}/tag/tag-key` | — | |

### Networking

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Networks (VPCs) | `GET /v2/{projectId}/networks` | Neutron Networks | |
| Subnets | `GET /v2/{projectId}/networks/{networkId}/subnets` | Neutron Subnets | |
| Security Groups | `GET /v2/{projectId}/secgroups` | Neutron SGs | |
| Security Group Rules | `GET /v2/{projectId}/secgroups/{secgroupId}/secGroupRules` | Neutron SG Rules | |
| Network ACLs | `GET /v2/{projectId}/network-acl/list` | — | |
| Network ACL Rules | `GET /v2/{projectId}/network-acl/{uuid}/rules` | — | |
| Route Tables | `GET /v2/{projectId}/route-table` | Neutron Routers | |
| Route Table Routes | `GET /v2/{projectId}/route-table/route/{routeTableId}` | Neutron Routes | |
| DHCP Options | `GET /v2/{projectId}/dhcp_option` | Neutron DHCP | |
| Elastic IPs | `GET /v2/{projectId}/elastic-ips` | Neutron Floating IPs | |
| Network Interfaces | `GET /v2/{projectId}/network-interfaces-elastic` | Neutron Ports | |
| Virtual IPs | `GET /v2/{projectId}/virtualIpAddress` | Neutron VIPs | |
| Public VIPs | `GET /v2/{projectId}/public-vips/externalNetworkInterfaces` | — | |
| Peering | `GET /v2/{projectId}/peering` | — | |
| Interconnects | `GET /v2/{projectId}/interconnects` | — | |
| Interconnect Connections | `GET /v2/{projectId}/interconnects/{id}/connections` | — | |
| WAN IPs | `GET /v2/{projectId}/wanIps` | — | |

### Block Storage

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Volumes | `GET /v2/{projectId}/volumes` | Cinder Volumes | |
| Persistent Volumes | `GET /v2/{projectId}/persistent-volumes` | Cinder Volumes | |
| Volume Types | `GET /v1/{projectId}/volume_types` | Cinder Volume Types | |
| Volume Type Zones | `GET /v1/{projectId}/volume_type_zones` | Cinder AZ Types | |
| Encryption Types | `GET /v1/{projectId}/volumes/encryption_types` | Cinder Encryption | |

### Other

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Protocols | `GET /v2/protocols` | — | |

## 🌐 vNetwork (`vnetwork-gateway`)

Uses different host: `{region}-vnetwork.console.vngcloud.vn`. Requires `regionId` (MongoDB ObjectId, not region name).

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Regions | `GET /vnetwork/v1/regions` | Keystone Regions | |
| VPC Endpoints | `GET /vnetwork/v1/{regionId}/{projectId}/endpoints` | — | |

## ⚖️ vLB — Load Balancer (`iam-vlb-gateway`)

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Load Balancers | `GET /v2/{projectId}/loadBalancers` | Octavia LBs | |
| Pools | `GET /v2/{projectId}/loadBalancers/{lbId}/pools` | Octavia Pools | |
| Listeners | `GET /v2/{projectId}/loadBalancers/{lbId}/listeners` | Octavia Listeners | |
| L7 Policies | `GET /v2/{projectId}/loadBalancers/{lbId}/l7policies` | Octavia L7Policies | |
| Certificates | `GET /v2/{projectId}/certificates` | Barbican | |
| LB Packages | `GET /v2/{projectId}/packages` | — | |

## 🚢 vKS — Kubernetes (`iam-vserver-gateway`)

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Clusters | `GET /v2/{projectId}/clusters` | Magnum Clusters | |
| Cluster Node Groups | `GET /v2/{projectId}/clusters/{clusterId}/nodeGroups` | Magnum Node Groups | |

## 🗄️ vDB — Database (`vdb-gateway.vngcloud.vn`)

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Relational DBs | `GET /v1/{projectId}/relational/databases` | Trove Instances | |
| Relational Backups | `GET /v1/{projectId}/relational/backups` | Trove Backups | |
| Relational Config Groups | `GET /v1/{projectId}/relational/config-groups` | Trove Configurations | |
| Memstore DBs | `GET /v1/{projectId}/memstore/databases` | Trove Instances | |
| Memstore Backups | `GET /v1/{projectId}/memstore/backups` | Trove Backups | |
| Memstore Config Groups | `GET /v1/{projectId}/memstore/config-groups` | Trove Configurations | |
| DB Packages | `GET /v1/{projectId}/packages` | — | |

## 📦 vStorage — Object Storage

S3-compatible and Swift-compatible protocols. Separate service, not behind vServer gateway.

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Containers | Swift API: `GET /v1/{account}` | Swift Containers | |
| Objects | Swift API: `GET /v1/{account}/{container}` | Swift Objects | |
| Buckets | S3 API: `GET /` | S3 ListBuckets | |

## 📊 vMonitor (`vmonitor.console.vngcloud.vn`)

Uses different host pattern. Note: `lstData` wrapper (not `listData`).

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Dashboards | `GET /vmonitor-api/api/v1/dashboards` | Grafana | |
| Configurations | `GET /vmonitor-api/api/v1/configurations/key/{key}` | — | |
| User Info | `GET /user-api/v1/userinfo` | — | |
| Quota Status | `POST /billing-api/v1/introspect-quotas` | — | |

## 💰 Billing (`iam-billing-gateway`)

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| User Info | `GET /v1/users/info` | — | |

## 🔑 IAM (`iamapis.vngcloud.vn`)

| Resource | Endpoint | OpenStack | Status |
|----------|----------|-----------|:------:|
| Token | `POST /accounts-api/v2/auth/token` | Keystone Token | |
| User Info | `GET /accounts-api/v1/auth/userinfo` | Keystone User | |

## 📊 Summary

**Total: 0/74 (0%)**

| Category | Implemented | Total |
|----------|:-----------:|:-----:|
| Compute | 0 | 16 |
| Networking | 0 | 17 |
| Block Storage | 0 | 5 |
| vNetwork | 0 | 2 |
| Load Balancer | 0 | 6 |
| Kubernetes | 0 | 2 |
| Database | 0 | 7 |
| Object Storage | 0 | 3 |
| Monitoring | 0 | 4 |
| Billing | 0 | 1 |
| IAM | 0 | 2 |
| Other | 0 | 1 |

## References

- [VNG Cloud API Docs](https://docs.api.vngcloud.vn/)
- [VNG Cloud Help Center](https://docs.vngcloud.vn/)
- [Terraform Provider](https://registry.terraform.io/providers/vngcloud/vngcloud/latest/docs) (52 resources)
- [Go SDK](https://github.com/vngcloud/vngcloud-go-sdk) (v2.18.4)
