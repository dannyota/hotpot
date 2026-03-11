<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/glb/regions'

const columns: ColumnDef[] = [
  { key: 'resource_id', label: 'Resource ID', format: 'mono' },
  { key: 'name', format: 'bold' },
  { key: 'vserver_endpoint', label: 'VServer Endpoint', format: 'mono' },
  { key: 'vlb_endpoint', label: 'VLB Endpoint', format: 'mono' },
  { key: 'ui_server_endpoint', label: 'UI Server Endpoint', format: 'mono' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'vserver_endpoint', label: 'VServer Endpoint', mono: true },
  { key: 'vlb_endpoint', label: 'VLB Endpoint', mono: true },
  { key: 'ui_server_endpoint', label: 'UI Server Endpoint', mono: true },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'first_collected_at', label: 'First Seen', format: 'date' },
  { key: 'collected_at', label: 'Last Seen', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    default-sort="-collected_at"
    export-filename="greennode-glb-regions.csv"
    title="GLB Regions"
  />
</template>
