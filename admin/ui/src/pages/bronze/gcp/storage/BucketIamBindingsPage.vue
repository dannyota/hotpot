<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/storage/bucket-iam-bindings'

const columns: ColumnDef[] = [
  { key: 'role', format: 'bold' },
  { key: 'bucket_name', label: 'Bucket' },
  { key: 'project_id', label: 'Project', format: 'mono' },
]

const filters: FilterDef[] = [
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'role', label: 'Role', mono: true },
  { key: 'members_json', label: 'Members' },
  { key: 'condition_json', label: 'Condition' },
  { key: 'bucket_name', label: 'Bucket' },
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
    search-key="role"
    search-placeholder="Search by role..."
    export-filename="gcp-bucket-iam-bindings.csv"
    drawer-title-key="role"
    title="Bucket IAM Bindings"
  />
</template>
