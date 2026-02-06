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

### Networking

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Networks | `NetworksClient` | `List()` | ✅ |
| Subnetworks | `SubnetworksClient` | `AggregatedList()` | ✅ |
| Addresses (Regional) | `AddressesClient` | `AggregatedList()` | ✅ |
| Addresses (Global) | `GlobalAddressesClient` | `List()` | ✅ |

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

### VPN

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| VPN Gateways (HA) | `VpnGatewaysClient` | `AggregatedList()` | |
| VPN Gateways (Classic) | `TargetVpnGatewaysClient` | `AggregatedList()` | |
| VPN Tunnels | `VpnTunnelsClient` | `AggregatedList()` | |

## Container API (`container.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| GKE Clusters | `ClusterManagerClient` | `ListClusters()` | ✅ |
| Node Pools | (included in cluster response) | — | ✅ |

## Resource Manager API (`resourcemanager.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| Projects | `ProjectsClient` | `SearchProjects()` | ✅ |

## VPC Access API (`vpcaccess.googleapis.com`)

| Resource | API Client | Method | Status |
|----------|-----------|--------|:------:|
| VPC Connectors | `VpcAccessClient` | `ListConnectors()` | |

## Summary

**Total: 13/24 (54%)**

| API | Implemented | Total |
|-----|:-----------:|:-----:|
| Compute Engine | 11 | 21 |
| Container | 2 | 2 |
| Resource Manager | 1 | 1 |
| VPC Access | 0 | 1 |
