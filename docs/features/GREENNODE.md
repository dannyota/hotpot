# GreenNode

GreenNode (formerly VNG Cloud) resource ingestion coverage in the bronze layer.

GreenNode wraps OpenStack services behind proprietary API gateways. No standard OpenStack APIs are exposed — all endpoints use custom REST with GreenNode-specific conventions.

Go SDK: `danny.vn/greennode`

## Auth

Two authentication flows available:

| Flow | Use Case | Endpoint |
|------|----------|----------|
| OAuth2 PKCE | Browser/console simulation | `signin.vngcloud.vn/ap/auth/iam/login` |
| Service Account | SDK/Terraform (recommended) | `iamapis.vngcloud.vn/accounts-api/v2/auth/token` |

Service Account uses `client_id` + `client_secret` (created via IAM console), returns a bearer token valid for 7200s.

## API Gateways

| Gateway | SDK Client | Host Pattern |
|---------|-----------|--------------|
| vServer | `Compute` | `{region}.api.vngcloud.vn/vserver/vserver-gateway` |
| Portal | `Portal`, `PortalV1` | `{region}.api.vngcloud.vn/vserver/vserver-gateway` |
| vNetwork | `Network`, `NetworkV1` | `{region}-vnetwork.console.vngcloud.vn/vnetwork-gateway` |
| vLB | `LoadBalancer` | `{region}.api.vngcloud.vn/vserver/vlb-gateway` |
| GLB | `GLB` | Global load balancer |
| Volume | `Volume`, `VolumeV1` | `{region}.api.vngcloud.vn/vserver/vserver-gateway` |
| DNS | `DNS` | DNS service |
| IAM | `Identity` | `iamapis.vngcloud.vn/accounts-api` |

Regions: `hcm-3` (Ho Chi Minh City), `han-1` (Hanoi). All resources scoped to `{region}` + `{projectId}`.

## Pagination

| Gateway | Style | Wrapper |
|---------|-------|---------|
| vServer (most) | `?page=N&size=50` | `{listData, page, pageSize, totalPage, totalItem}` |
| vServer (subnets) | None (returns array) | Direct `[]SubnetDto` |
| vServer (projects) | None | `{success, projects}` |
| vNetwork (endpoints) | JSON `params` query string | `{success, data, page, size, totalPage, total}` |
| vMonitor | `?page=N&size=50` | `{lstData, page, pageSize, totalPage, totalItem}` (note: `lstData` not `listData`) |

## Portal API (`Portal`)

| Resource | SDK Method | REST Endpoint | Status |
|----------|-----------|---------------|:------:|
| Regions | `Portal.ListRegions()` | `GET /v2/{projectId}/region` | ✅ |
| Quotas | `Portal.ListAllQuotaUsed()` | `GET /v2/{projectId}/quotas/quotaUsed` | ✅ |
| Projects | `PortalV1.ListProjects()` | `GET /v1/projects` | |
| Zones | `PortalV1.ListZones()` | `GET /v1/{projectId}/zones` | |

## Compute API (`Compute`)

| Resource | SDK Method | REST Endpoint | Status |
|----------|-----------|---------------|:------:|
| Servers | `Compute.ListServers()` | `GET /v2/{projectId}/servers` | ✅ |
| Server Groups | `Compute.ListServerGroups()` | `GET /v2/{projectId}/serverGroups` | ✅ |
| Server Group Policies | `Compute.ListServerGroupPolicies()` | `GET /v2/{projectId}/serverGroups/policies` | |
| SSH Keys | `Compute.ListSSHKeys()` | `GET /v2/{projectId}/sshKeys` | ✅ |

### Not in SDK

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| OS Images | `GET /v1/{projectId}/images/os` | Glance Images | |
| GPU Images | `GET /v1/{projectId}/images/gpu` | Glance Images | |
| User Images | `GET /v2/{projectId}/user-images` | Glance Images | |
| Flavors | `GET /v1/{projectId}/flavors/families/{family}/platforms/{code}` | Nova Flavors | |
| Flavor Zones | `GET /v1/{projectId}/flavor_zones/product` | Nova AZ + Flavors | |
| Flavor Families | `GET /v1/{projectId}/flavor_zones/families` | — | |
| Tags | `GET /v2/{projectId}/tag` | — | |
| Tag Keys | `GET /v2/{projectId}/tag/tag-key` | — | |

## Network API (`Network`)

| Resource | SDK Method | REST Endpoint | Status |
|----------|-----------|---------------|:------:|
| Security Groups | `Network.ListSecgroup()` | `GET /v2/{projectId}/secgroups` | |
| Security Group Rules | `Network.ListSecgroupRulesBySecgroupID()` | `GET /v2/{projectId}/secgroups/{id}/secGroupRules` | |

### Not in SDK

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| Networks (VPCs) | `GET /v2/{projectId}/networks` | Neutron Networks | |
| Subnets | `GET /v2/{projectId}/networks/{networkId}/subnets` | Neutron Subnets | |
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

## vNetwork API (`NetworkV1`)

| Resource | SDK Method | REST Endpoint | Status |
|----------|-----------|---------------|:------:|
| VPC Endpoints | `NetworkV1.ListEndpoints()` | `GET /vnetwork/v1/{regionId}/{projectId}/endpoints` | |

### Not in SDK

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| Regions | `GET /vnetwork/v1/regions` | Keystone Regions | |

## Volume API (`Volume`)

| Resource | SDK Method | REST Endpoint | Status |
|----------|-----------|---------------|:------:|
| Block Volumes | `Volume.ListBlockVolumes()` | `GET /v2/{projectId}/volumes` | |
| Snapshots | `Volume.ListSnapshotsByBlockVolumeID()` | `GET /v2/{projectId}/volumes/{id}/snapshots` | |
| Volume Types | `VolumeV1.GetListVolumeTypes()` | `GET /v1/{projectId}/volume_types` | |
| Volume Type Zones | `VolumeV1.GetVolumeTypeZones()` | `GET /v1/{projectId}/volume_type_zones` | |

### Not in SDK

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| Persistent Volumes | `GET /v2/{projectId}/persistent-volumes` | Cinder Volumes | |
| Encryption Types | `GET /v1/{projectId}/volumes/encryption_types` | Cinder Encryption | |

## Load Balancer API (`LoadBalancer`)

| Resource | SDK Method | REST Endpoint | Status |
|----------|-----------|---------------|:------:|
| Load Balancers | `LoadBalancer.ListLoadBalancers()` | `GET /v2/{projectId}/loadBalancers` | |
| Listeners | `LoadBalancer.ListListenersByLoadBalancerID()` | `GET /v2/{projectId}/loadBalancers/{id}/listeners` | |
| Pools | `LoadBalancer.ListPoolsByLoadBalancerID()` | `GET /v2/{projectId}/loadBalancers/{id}/pools` | |
| Pool Members | `LoadBalancer.ListPoolMembers()` | `GET /v2/{projectId}/loadBalancers/{id}/pools/{id}/members` | |
| L7 Policies | `LoadBalancer.ListPolicies()` | `GET /v2/{projectId}/loadBalancers/{id}/l7policies` | |
| Certificates | `LoadBalancer.ListCertificates()` | `GET /v2/{projectId}/certificates` | |
| LB Packages | `LoadBalancer.ListLoadBalancerPackages()` | `GET /v2/{projectId}/packages` | |

## Global Load Balancer API (`GLB`)

| Resource | SDK Method | REST Endpoint | Status |
|----------|-----------|---------------|:------:|
| Global Load Balancers | `GLB.ListGlobalLoadBalancers()` | `GET /v1/loadBalancers` | |
| Global Listeners | `GLB.ListGlobalListeners()` | `GET /v1/loadBalancers/{id}/listeners` | |
| Global Pools | `GLB.ListGlobalPools()` | `GET /v1/loadBalancers/{id}/pools` | |
| Global Pool Members | `GLB.ListGlobalPoolMembers()` | `GET /v1/loadBalancers/{id}/pools/{id}/members` | |
| Global Packages | `GLB.ListGlobalPackages()` | `GET /v1/packages` | |
| Global Regions | `GLB.ListGlobalRegions()` | `GET /v1/regions` | |

## DNS API (`DNS`)

| Resource | SDK Method | REST Endpoint | Status |
|----------|-----------|---------------|:------:|
| Hosted Zones | `DNS.ListHostedZones()` | DNS hosted zones | |
| DNS Records | `DNS.ListRecords()` | DNS records | |

## IAM API (`Identity`)

| Resource | SDK Method | REST Endpoint | Status |
|----------|-----------|---------------|:------:|
| Access Token | `Identity.GetAccessToken()` | `POST /accounts-api/v2/auth/token` | |

### Not in SDK

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| User Info | `GET /accounts-api/v1/auth/userinfo` | Keystone User | |

## Kubernetes (vKS) — No SDK

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| Clusters | `GET /v2/{projectId}/clusters` | Magnum Clusters | |
| Cluster Node Groups | `GET /v2/{projectId}/clusters/{id}/nodeGroups` | Magnum Node Groups | |

## Database (vDB) — No SDK

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| Relational DBs | `GET /v1/{projectId}/relational/databases` | Trove Instances | |
| Relational Backups | `GET /v1/{projectId}/relational/backups` | Trove Backups | |
| Relational Config Groups | `GET /v1/{projectId}/relational/config-groups` | Trove Configurations | |
| Memstore DBs | `GET /v1/{projectId}/memstore/databases` | Trove Instances | |
| Memstore Backups | `GET /v1/{projectId}/memstore/backups` | Trove Backups | |
| Memstore Config Groups | `GET /v1/{projectId}/memstore/config-groups` | Trove Configurations | |
| DB Packages | `GET /v1/{projectId}/packages` | — | |

## Object Storage (vStorage) — No SDK

S3-compatible and Swift-compatible protocols. Separate service, not behind vServer gateway.

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| Containers | Swift: `GET /v1/{account}` | Swift Containers | |
| Objects | Swift: `GET /v1/{account}/{container}` | Swift Objects | |
| Buckets | S3: `GET /` | S3 ListBuckets | |

## Monitoring (vMonitor) — No SDK

Uses different host: `vmonitor.console.vngcloud.vn`. Note: `lstData` wrapper (not `listData`).

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| Dashboards | `GET /vmonitor-api/api/v1/dashboards` | Grafana | |
| Configurations | `GET /vmonitor-api/api/v1/configurations/key/{key}` | — | |
| User Info | `GET /user-api/v1/userinfo` | — | |
| Quota Status | `POST /billing-api/v1/introspect-quotas` | — | |

## Billing — No SDK

| Resource | REST Endpoint | OpenStack Analog | Status |
|----------|---------------|------------------|:------:|
| User Info | `GET /v1/users/info` | — | |

## Summary

**Total: 5/74 (7%)**

| Category | SDK Client | Implemented | Total |
|----------|-----------|:-----------:|:-----:|
| Portal | `Portal` | 2 | 4 |
| Compute | `Compute` | 3 | 12 |
| Network | `Network` | 0 | 17 |
| vNetwork | `NetworkV1` | 0 | 2 |
| Volume | `Volume` | 0 | 6 |
| Load Balancer | `LoadBalancer` | 0 | 7 |
| Global LB | `GLB` | 0 | 6 |
| DNS | `DNS` | 0 | 2 |
| IAM | `Identity` | 0 | 2 |
| Kubernetes | — | 0 | 2 |
| Database | — | 0 | 7 |
| Object Storage | — | 0 | 3 |
| Monitoring | — | 0 | 4 |
| Billing | — | 0 | 1 |
| Other | — | 0 | 1 |

## References

- [GreenNode API Docs](https://docs.api.vngcloud.vn/)
- [GreenNode Help Center](https://docs.vngcloud.vn/)
- [Terraform Provider](https://registry.terraform.io/providers/vngcloud/vngcloud/latest/docs) (52 resources)
- Go SDK: `danny.vn/greennode`
