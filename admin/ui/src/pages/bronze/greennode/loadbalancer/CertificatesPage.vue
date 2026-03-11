<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/loadbalancer/certificates'

const certTypeBadge = badgeColors({
  IMPORTED: badge.blue,
  MANAGED: badge.emerald,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'certificate_type', label: 'Type', badge: certTypeBadge },
  { key: 'domain_name', label: 'Domain', format: 'mono' },
  { key: 'subject' },
  { key: 'in_use', label: 'In Use', format: 'bool' },
  { key: 'key_algorithm', label: 'Key Algorithm' },
  { key: 'not_after', label: 'Expires', format: 'date' },
  { key: 'region' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'certificate_type', label: 'Type' },
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'certificate_type', label: 'Certificate Type' },
  { key: 'domain_name', label: 'Domain' },
  { key: 'subject', label: 'Subject' },
  { key: 'issuer', label: 'Issuer' },
  { key: 'in_use', label: 'In Use', format: 'bool' },
  { key: 'key_algorithm', label: 'Key Algorithm' },
  { key: 'signature_algorithm', label: 'Signature Algorithm' },
  { key: 'serial', label: 'Serial', mono: true },
  { key: 'not_before', label: 'Not Before', format: 'date' },
  { key: 'not_after', label: 'Not After', format: 'date' },
  { key: 'expired_at', label: 'Expired At', format: 'date' },
  { key: 'imported_at', label: 'Imported At', format: 'date' },
  { key: 'region', label: 'Region' },
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
    :filters="filters"
    default-sort="-collected_at"
    search-key="q"
    export-filename="greennode-certificates.csv"
    title="Certificates"
  />
</template>
