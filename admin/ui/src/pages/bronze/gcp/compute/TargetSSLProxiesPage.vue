<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/target-ssl-proxies'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'service', transform: shortPath },
  { key: 'ssl_policy', label: 'SSL Policy', transform: shortPath },
  { key: 'proxy_header', label: 'Proxy Header' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'service', label: 'Service', transform: shortPath },
  { key: 'ssl_policy', label: 'SSL Policy', transform: shortPath },
  { key: 'proxy_header', label: 'Proxy Header' },
  { key: 'certificate_map', label: 'Certificate Map' },
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
    export-filename="gcp-compute-target-ssl-proxies.csv"
    title="Target SSL Proxies"
  />
</template>
