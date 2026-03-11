<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/vpn-tunnels'

const statusBadge = badgeColors({
  ESTABLISHED: badge.emerald,
  PROVISIONING: badge.amber,
  WAITING_FOR_FULL_CONFIG: badge.amber,
  FIRST_HANDSHAKE: badge.blue,
  NO_INCOMING_PACKETS: badge.red,
  AUTHORIZATION_ERROR: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'region', transform: shortPath },
  { key: 'peer_ip', label: 'Peer IP', format: 'mono' },
  { key: 'ike_version', label: 'IKE', format: 'number' },
  { key: 'vpn_gateway', label: 'VPN Gateway', transform: shortPath },
  { key: 'router', transform: shortPath },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'region' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'status', label: 'Status' },
  { key: 'detailed_status', label: 'Detailed Status' },
  { key: 'region', label: 'Region', transform: shortPath },
  { key: 'peer_ip', label: 'Peer IP', mono: true },
  { key: 'ike_version', label: 'IKE Version' },
  { key: 'vpn_gateway', label: 'VPN Gateway', transform: shortPath },
  { key: 'target_vpn_gateway', label: 'Target VPN Gateway', transform: shortPath },
  { key: 'vpn_gateway_interface', label: 'VPN Gateway Interface' },
  { key: 'router', label: 'Router', transform: shortPath },
  { key: 'peer_external_gateway', label: 'Peer External Gateway', transform: shortPath },
  { key: 'peer_external_gateway_interface', label: 'Peer External Gateway Interface' },
  { key: 'peer_gcp_gateway', label: 'Peer GCP Gateway', transform: shortPath },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'self_link', label: 'Self Link', mono: true },
  { key: 'creation_timestamp', label: 'Created' },
  { key: 'first_collected_at', label: 'First Seen', format: 'date' },
  { key: 'collected_at', label: 'Last Seen', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-creation_timestamp"
    export-filename="gcp-compute-vpn-tunnels.csv"
    title="VPN Tunnels"
  />
</template>
