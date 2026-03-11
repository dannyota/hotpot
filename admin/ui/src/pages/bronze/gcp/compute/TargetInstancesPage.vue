<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/target-instances'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'zone', transform: shortPath },
  { key: 'instance', transform: shortPath },
  { key: 'nat_policy', label: 'NAT Policy' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'zone' },
  { key: 'nat_policy', label: 'NAT Policy' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'zone', label: 'Zone', transform: shortPath },
  { key: 'instance', label: 'Instance', transform: shortPath },
  { key: 'nat_policy', label: 'NAT Policy' },
  { key: 'network', label: 'Network', transform: shortPath },
  { key: 'security_policy', label: 'Security Policy', transform: shortPath },
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
    export-filename="gcp-compute-target-instances.csv"
    title="Target Instances"
  />
</template>
