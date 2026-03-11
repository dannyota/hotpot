<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import HDetailPage from '@/components/app/HDetailPage.vue'
import { shortPath, badge, badgeColors } from '@/composables/formatting'
import type { DetailFieldDef, DetailTabDef } from '@/types/table'

const route = useRoute()
const id = computed(() => route.params.id as string)
const ENDPOINT = computed(() => `/api/v1/bronze/gcp/compute/instances/${id.value}`)
const BASE = '/api/v1/bronze/gcp/compute/instances'

const statusBadge = badgeColors({
  RUNNING: badge.emerald,
  TERMINATED: badge.zinc,
  STOPPED: badge.zinc,
  SUSPENDED: badge.amber,
  STAGING: badge.blue,
})

const fields: DetailFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'status', label: 'Status', badge: statusBadge },
  { key: 'zone', label: 'Zone', transform: shortPath },
  { key: 'machine_type', label: 'Machine Type', mono: true, transform: shortPath },
  { key: 'cpu_platform', label: 'CPU Platform' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'hostname', label: 'Hostname' },
  { key: 'description', label: 'Description' },
  { key: 'deletion_protection', label: 'Delete Protection', format: 'bool' },
  { key: 'can_ip_forward', label: 'Can IP Forward', format: 'bool' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'last_start_timestamp', label: 'Last Start', format: 'date' },
  { key: 'last_stop_timestamp', label: 'Last Stop', format: 'date' },
  { key: 'first_collected_at', label: 'First Seen', format: 'date' },
  { key: 'collected_at', label: 'Last Seen', format: 'relative' },
  { key: 'self_link', label: 'Self Link', mono: true, fullWidth: true },
]

const tabs: DetailTabDef[] = [
  // Core
  {
    key: 'disks', label: 'Disks',
    apiEndpoint: computed(() => `${BASE}/${id.value}/disks`),
    columns: [
      { key: 'device_name', label: 'Device', format: 'bold' },
      { key: 'disk_name', label: 'Disk' },
      { key: 'disk_status', label: 'Status' },
      { key: 'disk_type', label: 'Type', format: 'mono' },
      { key: 'disk_size', label: 'Size (GB)', format: 'number' },
      { key: 'source_image', label: 'Image' },
      { key: 'boot', format: 'bool' },
      { key: 'auto_delete', label: 'Auto Del', format: 'bool' },
    ],
  },
  {
    key: 'nics', label: 'Network', edgeKey: 'edges.nics',
    columns: [
      { key: 'name', format: 'bold' },
      { key: 'network', transform: shortPath },
      { key: 'subnetwork', transform: shortPath },
      { key: 'network_ip', label: 'IP', format: 'mono' },
      { key: 'stack_type', label: 'Stack' },
    ],
  },
  // Security & networking
  {
    key: 'firewalls', label: 'Firewalls',
    apiEndpoint: computed(() => `${BASE}/${id.value}/firewalls`),
    columns: [
      { key: 'name', format: 'bold' },
      { key: 'direction' },
      { key: 'priority', format: 'number' },
      { key: 'disabled', format: 'bool' },
      { key: 'network', transform: shortPath },
      { key: 'creation_timestamp', label: 'Created', format: 'date' },
    ],
  },
  {
    key: 'addresses', label: 'Addresses',
    apiEndpoint: computed(() => `${BASE}/${id.value}/addresses`),
    columns: [
      { key: 'name', format: 'bold' },
      { key: 'address', format: 'mono' },
      { key: 'status' },
      { key: 'address_type', label: 'Type' },
      { key: 'network_tier', label: 'Tier' },
      { key: 'region', transform: shortPath },
      { key: 'creation_timestamp', label: 'Created', format: 'date' },
    ],
  },
  // Identity & config
  {
    key: 'labels', label: 'Labels', edgeKey: 'edges.labels',
    columns: [
      { key: 'key', format: 'bold' },
      { key: 'value' },
    ],
  },
  {
    key: 'tags', label: 'Tags', edgeKey: 'edges.tags',
    columns: [
      { key: 'tag', format: 'bold' },
    ],
  },
  {
    key: 'service_accounts', label: 'Service Accounts', edgeKey: 'edges.service_accounts',
    columns: [
      { key: 'email', format: 'mono' },
    ],
  },
  // Load balancing
  {
    key: 'instance-groups', label: 'Instance Groups',
    apiEndpoint: computed(() => `${BASE}/${id.value}/instance-groups`),
    columns: [
      { key: 'name', format: 'bold' },
      { key: 'zone', transform: shortPath },
      { key: 'size', format: 'number' },
      { key: 'network', transform: shortPath },
      { key: 'creation_timestamp', label: 'Created', format: 'date' },
    ],
  },
  {
    key: 'forwarding-rules', label: 'Forwarding Rules',
    apiEndpoint: computed(() => `${BASE}/${id.value}/forwarding-rules`),
    columns: [
      { key: 'name', format: 'bold' },
      { key: 'ip_address', label: 'IP', format: 'mono' },
      { key: 'ip_protocol', label: 'Protocol' },
      { key: 'ports', label: 'Ports', format: 'mono' },
      { key: 'load_balancing_scheme', label: 'Scheme' },
      { key: 'backend_service', label: 'Backend Service', transform: shortPath },
    ],
  },
  // Storage
  {
    key: 'metadata', label: 'Metadata', edgeKey: 'edges.metadata',
    columns: [
      { key: 'key', format: 'bold' },
      { key: 'value' },
    ],
  },
  {
    key: 'snapshots', label: 'Snapshots',
    apiEndpoint: computed(() => `${BASE}/${id.value}/snapshots`),
    columns: [
      { key: 'name', format: 'bold' },
      { key: 'status' },
      { key: 'snapshot_type', label: 'Type' },
      { key: 'disk_size_gb', label: 'Size (GB)', format: 'number' },
      { key: 'source_disk', label: 'Source Disk', transform: shortPath, format: 'mono' },
      { key: 'creation_timestamp', label: 'Created', format: 'date' },
    ],
  },
]
</script>

<template>
  <HDetailPage
    :endpoint="ENDPOINT"
    back-route="/bronze/gcp/compute/instances"
    back-label="Instances"
    title-key="name"
    :subtitle-fields="['status', 'zone', 'machine_type']"
    :fields="fields"
    :tabs="tabs"
  />
</template>
