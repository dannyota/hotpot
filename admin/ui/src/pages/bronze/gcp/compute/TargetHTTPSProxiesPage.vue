<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/target-https-proxies'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'url_map', label: 'URL Map', transform: shortPath },
  { key: 'ssl_policy', label: 'SSL Policy', transform: shortPath },
  { key: 'quic_override', label: 'QUIC' },
  { key: 'region', transform: shortPath },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'quic_override', label: 'QUIC' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'url_map', label: 'URL Map', transform: shortPath },
  { key: 'ssl_policy', label: 'SSL Policy', transform: shortPath },
  { key: 'quic_override', label: 'QUIC Override' },
  { key: 'region', label: 'Region', transform: shortPath },
  { key: 'server_tls_policy', label: 'Server TLS Policy' },
  { key: 'authorization_policy', label: 'Authorization Policy' },
  { key: 'certificate_map', label: 'Certificate Map' },
  { key: 'tls_early_data', label: 'TLS Early Data' },
  { key: 'proxy_bind', label: 'Proxy Bind', format: 'bool' },
  { key: 'http_keep_alive_timeout_sec', label: 'HTTP Keep-Alive Timeout', transform: (v) => v != null ? `${v}s` : '' },
  { key: 'fingerprint', label: 'Fingerprint' },
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
    export-filename="gcp-compute-target-https-proxies.csv"
    title="Target HTTPS Proxies"
  />
</template>
