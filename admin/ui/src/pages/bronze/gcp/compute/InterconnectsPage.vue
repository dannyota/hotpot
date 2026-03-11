<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/interconnects'

const opStatusBadge = badgeColors({
  OS_ACTIVE: badge.emerald,
  OS_UNPROVISIONED: badge.amber,
})

const stateBadge = badgeColors({
  ACTIVE: badge.emerald,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'interconnect_type', label: 'Type' },
  { key: 'link_type', label: 'Link' },
  { key: 'operational_status', label: 'Op Status', badge: opStatusBadge },
  { key: 'state', badge: stateBadge },
  { key: 'admin_enabled', label: 'Admin Enabled', format: 'bool' },
  { key: 'location' },
  { key: 'provisioned_link_count', label: 'Links', format: 'number' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'interconnect_type', label: 'Type' },
  { key: 'link_type', label: 'Link' },
  { key: 'operational_status', label: 'Op Status' },
  { key: 'state' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'interconnect_type', label: 'Type' },
  { key: 'link_type', label: 'Link Type' },
  { key: 'operational_status', label: 'Op Status' },
  { key: 'state', label: 'State' },
  { key: 'admin_enabled', label: 'Admin Enabled', format: 'bool' },
  { key: 'location', label: 'Location' },
  { key: 'provisioned_link_count', label: 'Provisioned Links' },
  { key: 'requested_link_count', label: 'Requested Links' },
  { key: 'peer_ip_address', label: 'Peer IP' },
  { key: 'google_ip_address', label: 'Google IP' },
  { key: 'google_reference_id', label: 'Google Reference' },
  { key: 'noc_contact_email', label: 'NOC Contact' },
  { key: 'customer_name', label: 'Customer' },
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
    export-filename="gcp-compute-interconnects.csv"
    title="Interconnects"
  />
</template>
