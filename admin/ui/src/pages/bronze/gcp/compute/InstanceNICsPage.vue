<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/instance-nics'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'network_ip', label: 'IP', format: 'mono' },
  { key: 'network', transform: shortPath },
  { key: 'subnetwork', transform: shortPath },
  { key: 'stack_type', label: 'Stack' },
  { key: 'nic_type', label: 'NIC Type' },
  { key: 'instance_name', label: 'Instance' },
  { key: 'instance_zone', label: 'Zone', transform: shortPath },
  { key: 'project_id', label: 'Project', format: 'mono' },
]

const filters: FilterDef[] = [
  { key: 'stack_type', label: 'Stack Type' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'network_ip', label: 'Network IP' },
  { key: 'network', label: 'Network', transform: shortPath },
  { key: 'subnetwork', label: 'Subnetwork', transform: shortPath },
  { key: 'stack_type', label: 'Stack Type' },
  { key: 'nic_type', label: 'NIC Type' },
  { key: 'instance_name', label: 'Instance' },
  { key: 'instance_zone', label: 'Zone', transform: shortPath },
  { key: 'project_id', label: 'Project ID', mono: true },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-id"
    search-placeholder="Search by name..."
    export-filename="gcp-compute-instance-nics.csv"
    drawer-title-key="name"
    title="Instance NICs"
  />
</template>
