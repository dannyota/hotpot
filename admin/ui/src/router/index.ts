import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

// ---------------------------------------------------------------------------
// Auto-discover page components via glob and build a normalized lookup.
// Key: lowercase path with hyphens stripped (e.g. "bronze/gcp/compute/vpngateways")
// Value: lazy import function
// ---------------------------------------------------------------------------

const pageModules = import.meta.glob<any>('../pages/**/*Page.vue')

const moduleLookup = new Map<string, () => Promise<any>>()
for (const key of Object.keys(pageModules)) {
  const normalized = key.replace('../pages/', '').replace(/Page\.vue$/, '').toLowerCase()
  moduleLookup.set(normalized, pageModules[key])
}

// Overrides for routes whose path doesn't match the file location.
const componentOverrides: Record<string, string> = {
  '/bronze/meec/inventory/computers':          'bronze/meec/computers',
  '/bronze/meec/inventory/software':           'bronze/meec/software',
  '/bronze/meec/inventory/installed-software':  'bronze/meec/installedsoftware',
  '/bronze/greennode/network/secgroups':        'bronze/greennode/network/securitygroups',
  '/bronze/greennode/loadbalancer/lbs':         'bronze/greennode/loadbalancer/loadbalancers',
}

function resolveComponent(routePath: string): () => Promise<any> {
  const normalized = componentOverrides[routePath]
    ?? routePath.slice(1).replace(/-/g, '').toLowerCase()
  const mod = moduleLookup.get(normalized)
  if (!mod) throw new Error(`No page component for route: ${routePath} (looked up "${normalized}")`)
  return mod
}

// ---------------------------------------------------------------------------
// Route definitions — compact tuples: [path, ...breadcrumb]
// Component and name are auto-derived from the path.
// ---------------------------------------------------------------------------

// prettier-ignore
const pageRoutes: [string, ...string[]][] = [
  // Bronze: GCP Compute — Instances
  ['/bronze/gcp/compute/instances',                'Bronze', 'GCP', 'Compute', 'Instances', 'Instances'],
  ['/bronze/gcp/compute/instance-groups',          'Bronze', 'GCP', 'Compute', 'Instances', 'Instance Groups'],
  ['/bronze/gcp/compute/target-instances',         'Bronze', 'GCP', 'Compute', 'Instances', 'Target Instances'],
  ['/bronze/gcp/compute/target-pools',             'Bronze', 'GCP', 'Compute', 'Instances', 'Target Pools'],
  ['/bronze/gcp/compute/instance-disks',           'Bronze', 'GCP', 'Compute', 'Instances', 'Instance Disks'],
  ['/bronze/gcp/compute/instance-nics',            'Bronze', 'GCP', 'Compute', 'Instances', 'Instance NICs'],
  ['/bronze/gcp/compute/instance-service-accounts','Bronze', 'GCP', 'Compute', 'Instances', 'Instance Service Accounts'],
  ['/bronze/gcp/compute/instance-group-members',   'Bronze', 'GCP', 'Compute', 'Instances', 'Instance Group Members'],

  // Bronze: GCP Compute — Storage
  ['/bronze/gcp/compute/disks',              'Bronze', 'GCP', 'Compute', 'Storage', 'Disks'],
  ['/bronze/gcp/compute/snapshots',          'Bronze', 'GCP', 'Compute', 'Storage', 'Snapshots'],
  ['/bronze/gcp/compute/images',             'Bronze', 'GCP', 'Compute', 'Storage', 'Images'],

  // Bronze: GCP Compute — Networking
  ['/bronze/gcp/compute/networks',                      'Bronze', 'GCP', 'Compute', 'Networking', 'Networks'],
  ['/bronze/gcp/compute/subnetworks',                   'Bronze', 'GCP', 'Compute', 'Networking', 'Subnetworks'],
  ['/bronze/gcp/compute/addresses',                     'Bronze', 'GCP', 'Compute', 'Networking', 'Addresses'],
  ['/bronze/gcp/compute/global-addresses',              'Bronze', 'GCP', 'Compute', 'Networking', 'Global Addresses'],
  ['/bronze/gcp/compute/firewalls',                     'Bronze', 'GCP', 'Compute', 'Networking', 'Firewalls'],
  ['/bronze/gcp/compute/routers',                       'Bronze', 'GCP', 'Compute', 'Networking', 'Routers'],
  ['/bronze/gcp/compute/firewall-rules',                'Bronze', 'GCP', 'Compute', 'Networking', 'Firewall Allow Rules'],
  ['/bronze/gcp/compute/firewall-deny-rules',           'Bronze', 'GCP', 'Compute', 'Networking', 'Firewall Deny Rules'],
  ['/bronze/gcp/compute/network-peerings',              'Bronze', 'GCP', 'Compute', 'Networking', 'Network Peerings'],
  ['/bronze/gcp/compute/subnetwork-secondary-ranges',   'Bronze', 'GCP', 'Compute', 'Networking', 'Secondary Ranges'],

  // Bronze: GCP Compute — Interconnect & VPN
  ['/bronze/gcp/compute/interconnects',      'Bronze', 'GCP', 'Compute', 'Interconnect & VPN', 'Interconnects'],
  ['/bronze/gcp/compute/vpn-gateways',       'Bronze', 'GCP', 'Compute', 'Interconnect & VPN', 'VPN Gateways'],
  ['/bronze/gcp/compute/target-vpn-gateways','Bronze', 'GCP', 'Compute', 'Interconnect & VPN', 'Target VPN Gateways'],
  ['/bronze/gcp/compute/vpn-tunnels',        'Bronze', 'GCP', 'Compute', 'Interconnect & VPN', 'VPN Tunnels'],
  ['/bronze/gcp/compute/packet-mirrorings',  'Bronze', 'GCP', 'Compute', 'Interconnect & VPN', 'Packet Mirrorings'],

  // Bronze: GCP Compute — Load Balancing
  ['/bronze/gcp/compute/backend-services',         'Bronze', 'GCP', 'Compute', 'Load Balancing', 'Backend Services'],
  ['/bronze/gcp/compute/backend-service-backends', 'Bronze', 'GCP', 'Compute', 'Load Balancing', 'Backend Service Backends'],
  ['/bronze/gcp/compute/forwarding-rules',       'Bronze', 'GCP', 'Compute', 'Load Balancing', 'Forwarding Rules'],
  ['/bronze/gcp/compute/global-forwarding-rules','Bronze', 'GCP', 'Compute', 'Load Balancing', 'Global Forwarding Rules'],
  ['/bronze/gcp/compute/health-checks',          'Bronze', 'GCP', 'Compute', 'Load Balancing', 'Health Checks'],
  ['/bronze/gcp/compute/negs',                   'Bronze', 'GCP', 'Compute', 'Load Balancing', 'NEGs'],
  ['/bronze/gcp/compute/ssl-policies',           'Bronze', 'GCP', 'Compute', 'Load Balancing', 'SSL Policies'],
  ['/bronze/gcp/compute/target-http-proxies',    'Bronze', 'GCP', 'Compute', 'Load Balancing', 'Target HTTP Proxies'],
  ['/bronze/gcp/compute/target-https-proxies',   'Bronze', 'GCP', 'Compute', 'Load Balancing', 'Target HTTPS Proxies'],
  ['/bronze/gcp/compute/target-ssl-proxies',     'Bronze', 'GCP', 'Compute', 'Load Balancing', 'Target SSL Proxies'],
  ['/bronze/gcp/compute/target-tcp-proxies',     'Bronze', 'GCP', 'Compute', 'Load Balancing', 'Target TCP Proxies'],
  ['/bronze/gcp/compute/url-maps',               'Bronze', 'GCP', 'Compute', 'Load Balancing', 'URL Maps'],

  // Bronze: GCP Compute — Security & Config
  ['/bronze/gcp/compute/security-policies',  'Bronze', 'GCP', 'Compute', 'Security & Config', 'Security Policies'],
  ['/bronze/gcp/compute/project-metadata',   'Bronze', 'GCP', 'Compute', 'Security & Config', 'Project Metadata'],

  // Bronze: GCP Container
  ['/bronze/gcp/container/clusters',         'Bronze', 'GCP', 'Container', 'Clusters'],
  ['/bronze/gcp/container/node-pools',       'Bronze', 'GCP', 'Container', 'Node Pools'],

  // Bronze: GCP IAM
  ['/bronze/gcp/iam/service-accounts',       'Bronze', 'GCP', 'IAM', 'Service Accounts'],
  ['/bronze/gcp/iam/service-account-keys',   'Bronze', 'GCP', 'IAM', 'Service Account Keys'],

  // Bronze: GCP IAM & Resource Manager
  ['/bronze/gcp/resourcemanager/project-iam-bindings', 'Bronze', 'GCP', 'Resource Manager', 'Project IAM Bindings'],
  ['/bronze/gcp/resourcemanager/folder-iam-bindings',  'Bronze', 'GCP', 'Resource Manager', 'Folder IAM Bindings'],
  ['/bronze/gcp/resourcemanager/org-iam-bindings',     'Bronze', 'GCP', 'Resource Manager', 'Org IAM Bindings'],

  // Bronze: GCP Storage
  ['/bronze/gcp/storage/buckets',              'Bronze', 'GCP', 'Storage', 'Buckets'],
  ['/bronze/gcp/storage/bucket-iam-bindings',  'Bronze', 'GCP', 'Storage', 'Bucket IAM Bindings'],

  // Bronze: GCP Cloud SQL
  ['/bronze/gcp/sql/instances',              'Bronze', 'GCP', 'Cloud SQL', 'Instances'],

  // Bronze: GreenNode Compute
  ['/bronze/greennode/compute/servers',       'Bronze', 'GreenNode', 'Compute', 'Servers'],
  ['/bronze/greennode/compute/server-groups', 'Bronze', 'GreenNode', 'Compute', 'Server Groups'],
  ['/bronze/greennode/compute/ssh-keys',      'Bronze', 'GreenNode', 'Compute', 'SSH Keys'],
  ['/bronze/greennode/compute/os-images',     'Bronze', 'GreenNode', 'Compute', 'OS Images'],
  ['/bronze/greennode/compute/user-images',   'Bronze', 'GreenNode', 'Compute', 'User Images'],

  // Bronze: GreenNode Network
  ['/bronze/greennode/network/vpcs',          'Bronze', 'GreenNode', 'Network', 'VPCs'],
  ['/bronze/greennode/network/secgroups',     'Bronze', 'GreenNode', 'Network', 'Security Groups'],
  ['/bronze/greennode/network/subnets',       'Bronze', 'GreenNode', 'Network', 'Subnets'],
  ['/bronze/greennode/network/endpoints',     'Bronze', 'GreenNode', 'Network', 'Endpoints'],
  ['/bronze/greennode/network/interconnects', 'Bronze', 'GreenNode', 'Network', 'Interconnects'],
  ['/bronze/greennode/network/peerings',      'Bronze', 'GreenNode', 'Network', 'Peerings'],
  ['/bronze/greennode/network/route-tables',  'Bronze', 'GreenNode', 'Network', 'Route Tables'],

  // Bronze: GreenNode Load Balancer
  ['/bronze/greennode/loadbalancer/lbs',          'Bronze', 'GreenNode', 'Load Balancer', 'Load Balancers'],
  ['/bronze/greennode/loadbalancer/certificates', 'Bronze', 'GreenNode', 'Load Balancer', 'Certificates'],
  ['/bronze/greennode/loadbalancer/packages',     'Bronze', 'GreenNode', 'Load Balancer', 'Packages'],

  // Bronze: GreenNode Volume
  ['/bronze/greennode/volume/block-volumes',  'Bronze', 'GreenNode', 'Volume', 'Block Volumes'],
  ['/bronze/greennode/volume/volume-types',   'Bronze', 'GreenNode', 'Volume', 'Volume Types'],
  ['/bronze/greennode/volume/snapshots',      'Bronze', 'GreenNode', 'Volume', 'Snapshots'],

  // Bronze: GreenNode Portal
  ['/bronze/greennode/portal/regions',        'Bronze', 'GreenNode', 'Portal', 'Regions'],
  ['/bronze/greennode/portal/zones',          'Bronze', 'GreenNode', 'Portal', 'Zones'],
  ['/bronze/greennode/portal/quotas',         'Bronze', 'GreenNode', 'Portal', 'Quotas'],

  // Bronze: GreenNode DNS
  ['/bronze/greennode/dns/hosted-zones',      'Bronze', 'GreenNode', 'DNS', 'Hosted Zones'],
  ['/bronze/greennode/dns/records',           'Bronze', 'GreenNode', 'DNS', 'Records'],

  // Bronze: GreenNode GLB
  ['/bronze/greennode/glb/load-balancers',    'Bronze', 'GreenNode', 'GLB', 'Load Balancers'],
  ['/bronze/greennode/glb/packages',          'Bronze', 'GreenNode', 'GLB', 'Packages'],
  ['/bronze/greennode/glb/regions',           'Bronze', 'GreenNode', 'GLB', 'Regions'],

  // Bronze: SentinelOne
  ['/bronze/s1/agents',              'Bronze', 'SentinelOne', 'Agents'],
  ['/bronze/s1/accounts',            'Bronze', 'SentinelOne', 'Accounts'],
  ['/bronze/s1/sites',               'Bronze', 'SentinelOne', 'Sites'],
  ['/bronze/s1/groups',              'Bronze', 'SentinelOne', 'Groups'],
  ['/bronze/s1/app-inventory',       'Bronze', 'SentinelOne', 'App Inventory'],
  ['/bronze/s1/endpoint-apps',       'Bronze', 'SentinelOne', 'Endpoint Apps'],
  ['/bronze/s1/network-discoveries', 'Bronze', 'SentinelOne', 'Network Discoveries'],
  ['/bronze/s1/ranger-devices',      'Bronze', 'SentinelOne', 'Ranger Devices'],
  ['/bronze/s1/ranger-gateways',     'Bronze', 'SentinelOne', 'Ranger Gateways'],
  ['/bronze/s1/ranger-settings',     'Bronze', 'SentinelOne', 'Ranger Settings'],

  // Bronze: MEEC
  ['/bronze/meec/inventory/computers',          'Bronze', 'MEEC', 'Computers'],
  ['/bronze/meec/inventory/software',           'Bronze', 'MEEC', 'Software'],
  ['/bronze/meec/inventory/installed-software', 'Bronze', 'MEEC', 'Installed Software'],

  // Silver: Inventory
  ['/silver/inventory/machines',       'Silver', 'Inventory', 'Machines'],
  ['/silver/inventory/k8s-nodes',      'Silver', 'Inventory', 'K8s Nodes'],
  ['/silver/inventory/software',       'Silver', 'Inventory', 'Software'],
  ['/silver/inventory/api-endpoints',  'Silver', 'Inventory', 'API Endpoints'],

  // Silver: HTTP Traffic
  ['/silver/httptraffic/traffic-5m',    'Silver', 'HTTP Traffic', 'Traffic 5m'],
  ['/silver/httptraffic/client-ip-5m',  'Silver', 'HTTP Traffic', 'Client IP 5m'],
  ['/silver/httptraffic/user-agent-5m', 'Silver', 'HTTP Traffic', 'User Agent 5m'],

  // Gold: Lifecycle
  ['/gold/lifecycle/software', 'Gold', 'Lifecycle', 'Software EOL'],
  ['/gold/lifecycle/os',       'Gold', 'Lifecycle', 'OS EOL'],

  // Gold: HTTP Monitor
  ['/gold/httpmonitor/anomalies', 'Gold', 'HTTP Monitor', 'Anomalies'],
]

// ---------------------------------------------------------------------------
// Build route records
// ---------------------------------------------------------------------------

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/dashboard' },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('@/pages/dashboard/DashboardPage.vue'),
    meta: { breadcrumb: ['Dashboard'] },
  },

  // Expand compact page routes.
  ...pageRoutes.map(([path, ...breadcrumb]): RouteRecordRaw => ({
    path,
    name: path.slice(1).replace(/\//g, '-'),
    component: resolveComponent(path),
    meta: { breadcrumb },
  })),

  // Detail routes (parameterized — can't use compact tuple format)
  {
    path: '/bronze/gcp/compute/instances/:id',
    name: 'bronze-gcp-compute-instance-detail',
    component: () => import('@/pages/bronze/gcp/compute/InstanceDetailPage.vue'),
    meta: { breadcrumb: ['Bronze', 'GCP', 'Compute', 'Instances', 'Detail'] },
  },

  // Catch-all — used for dynamically registered routes (GenericTablePage).
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/pages/GenericTablePage.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router

/**
 * Register dynamic routes from the UI config nav tree.
 * Called after loadUIConfig() completes.
 * Only adds routes for nav items with an `api` field that don't already
 * have a specific route defined above.
 */
export function registerNavRoutes(nav: import('@/composables/useUIConfig').NavItem[]) {
  function walk(items: import('@/composables/useUIConfig').NavItem[], breadcrumb: string[]) {
    for (const item of items) {
      if (item.children) {
        walk(item.children, [...breadcrumb, item.label])
      } else if (item.path && item.api) {
        // Skip if a named route already handles this path.
        const existing = router.resolve(item.path)
        if (existing.name === 'not-found' || existing.matched.length === 0) {
          router.addRoute({
            path: item.path,
            component: () => import('@/pages/GenericTablePage.vue'),
            meta: {
              api: item.api,
              label: item.label,
              breadcrumb: [...breadcrumb, item.label],
            },
          })
        }
      }
    }
  }
  walk(nav, [])
}
