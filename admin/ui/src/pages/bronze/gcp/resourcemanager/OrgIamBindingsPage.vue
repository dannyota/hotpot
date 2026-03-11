<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/resourcemanager/org-iam-bindings'

const columns: ColumnDef[] = [
  { key: 'role', format: 'bold' },
  { key: 'org_name', label: 'Organization' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'role', label: 'Role', mono: true },
  { key: 'members_json', label: 'Members' },
  { key: 'condition_json', label: 'Condition' },
  { key: 'org_name', label: 'Organization' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    default-sort="-id"
    search-key="role"
    search-placeholder="Search by role..."
    export-filename="gcp-org-iam-bindings.csv"
    drawer-title-key="role"
    title="Org IAM Bindings"
  />
</template>
