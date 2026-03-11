<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Cloud, Server, Shield, Monitor, Layers, Activity, ShieldAlert, AlertTriangle, Lock, Globe } from 'lucide-vue-next'
import { useApi } from '@/composables/useApi'
import type { StatsOverview, StatValue } from '@/types/api'
import HAlert from '@/components/app/HAlert.vue'

const router = useRouter()
const { data: stats, error, fetch: fetchStats } = useApi<StatsOverview>('/api/v1/stats/overview')

onMounted(fetchStats)

const providers = [
  { key: 'gcp', label: 'GCP', icon: Cloud, color: '#4285F4', path: '/bronze/gcp/compute/instances' },
  { key: 'greennode', label: 'GreenNode', icon: Server, color: '#34A853', path: '/bronze/greennode/compute/servers' },
  { key: 's1', label: 'SentinelOne', icon: Shield, color: '#6C2DC7', path: '/bronze/s1/agents' },
  { key: 'meec', label: 'MEEC', icon: Monitor, color: '#3B82F6', path: '/bronze/meec/inventory/computers' },
  { key: 'vault', label: 'Vault', icon: Lock, color: '#EAB308', path: '/bronze/vault/pki/certificates' },
  { key: 'apicatalog', label: 'API Catalog', icon: Globe, color: '#14B8A6', path: '/bronze/apicatalog/endpoints' },
]

const silverGroups = [
  {
    label: 'Inventory', icon: Layers, gridClass: 'grid-cols-2',
    items: [
      { key: 'machines', label: 'Machines', path: '/silver/inventory/machines' },
      { key: 'k8s_nodes', label: 'K8s Nodes', path: '/silver/inventory/k8s-nodes' },
      { key: 'software', label: 'Software', path: '/silver/inventory/software' },
      { key: 'api_endpoints', label: 'API Endpoints', path: '/silver/inventory/api-endpoints' },
    ],
  },
  {
    label: 'HTTP Traffic', icon: Activity, gridClass: 'grid-cols-3',
    items: [
      { key: 'traffic_5m', label: 'Traffic 5m', path: '/silver/httptraffic/traffic-5m' },
      { key: 'client_ips_5m', label: 'Client IPs 5m', path: '/silver/httptraffic/client-ip-5m' },
      { key: 'user_agents_5m', label: 'User Agents 5m', path: '/silver/httptraffic/user-agent-5m' },
    ],
  },
]

interface GoldRow { label: string; count?: number; delta?: number; color: string }
interface GoldCard {
  label: string; subtitle: string; path: string
  icon: typeof ShieldAlert; iconColor: string; iconBg: string
  total?: number; delta?: number; rows: GoldRow[]
}

const goldCards = computed<GoldCard[]>(() => {
  const g = stats.value?.gold
  return [
    {
      label: 'Software EOL', subtitle: 'expired or end-of-support', path: '/gold/lifecycle/software',
      icon: ShieldAlert, iconColor: 'text-red-500', iconBg: 'bg-red-500/10',
      total: sumCounts(g?.software_eol, g?.software_eoes),
      delta: sumDeltas(g?.software_eol, g?.software_eoes),
      rows: [
        { label: 'EOL Expired', count: g?.software_eol?.count, delta: g?.software_eol?.delta, color: 'text-red-500' },
        { label: 'EOES Expired', count: g?.software_eoes?.count, delta: g?.software_eoes?.delta, color: 'text-amber-500' },
      ],
    },
    {
      label: 'OS EOL', subtitle: 'machines with expired OS', path: '/gold/lifecycle/os',
      icon: Monitor, iconColor: 'text-orange-500', iconBg: 'bg-orange-500/10',
      total: sumCounts(g?.os_eol, g?.os_eoes),
      delta: sumDeltas(g?.os_eol, g?.os_eoes),
      rows: [
        { label: 'EOL Expired', count: g?.os_eol?.count, delta: g?.os_eol?.delta, color: 'text-red-500' },
        { label: 'EOES Expired', count: g?.os_eoes?.count, delta: g?.os_eoes?.delta, color: 'text-amber-500' },
      ],
    },
    {
      label: 'HTTP Anomalies', subtitle: 'detected', path: '/gold/httpmonitor/anomalies',
      icon: AlertTriangle, iconColor: 'text-purple-500', iconBg: 'bg-purple-500/10',
      total: g?.anomalies?.count, delta: g?.anomalies?.delta,
      rows: [
        { label: 'Critical', count: g?.anomalies_critical?.count, delta: g?.anomalies_critical?.delta, color: 'text-red-500' },
        { label: 'High', count: g?.anomalies_high?.count, delta: g?.anomalies_high?.delta, color: 'text-orange-500' },
        { label: 'Medium', count: g?.anomalies_medium?.count, delta: g?.anomalies_medium?.delta, color: 'text-amber-500' },
      ],
    },
  ]
})

function sumCounts(...vals: (StatValue | undefined)[]): number | undefined {
  if (vals.every(v => v == null)) return undefined
  return vals.reduce((sum, v) => sum + (v?.count ?? 0), 0)
}

function sumDeltas(...vals: (StatValue | undefined)[]): number | undefined {
  if (vals.every(v => v?.delta == null)) return undefined
  return vals.reduce((sum, v) => sum + (v?.delta ?? 0), 0)
}

function fmt(n: number | undefined | null): string {
  if (n == null) return '--'
  return n.toLocaleString()
}

function deltaText(d: number | undefined | null): string {
  if (d == null || d === 0) return ''
  return d > 0 ? `+${d.toLocaleString()}` : d.toLocaleString()
}

function deltaClass(d: number | undefined | null, invert = false): string {
  if (d == null || d === 0) return 'text-zinc-500 dark:text-zinc-500'
  const positive = invert ? 'text-red-500' : 'text-emerald-500'
  const negative = invert ? 'text-emerald-500' : 'text-red-500'
  return d > 0 ? positive : negative
}
</script>

<template>
  <div class="p-6 space-y-8 max-w-7xl">
    <HAlert v-if="error" type="error" :message="error" />

    <!-- ═══ Bronze ═══ -->
    <section class="animate-in fade-in duration-300">
      <div class="flex items-center gap-2 mb-3">
        <span class="text-xs font-semibold uppercase tracking-wide text-amber-500">Bronze</span>
        <span class="text-[10px] font-semibold uppercase tracking-wide text-amber-400/80 bg-amber-500/10 px-1.5 py-0.5 rounded">Raw Data</span>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="p in providers"
          :key="p.key"
          class="rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-4 cursor-pointer hover:border-zinc-300 dark:hover:border-zinc-700 hover:shadow-sm transition-all"
          @click="router.push(p.path)"
        >
          <div class="flex items-center gap-2.5 mb-3">
            <div class="w-8 h-8 rounded-lg flex items-center justify-center shrink-0" :style="{ backgroundColor: p.color + '14' }">
              <component :is="p.icon" class="w-4 h-4" :style="{ color: p.color }" />
            </div>
            <span class="text-sm font-semibold text-zinc-800 dark:text-zinc-200">{{ p.label }}</span>
          </div>
          <div
            v-for="(r, i) in stats?.bronze?.[p.key]?.resources ?? []"
            :key="r.label"
            class="flex justify-between items-center py-1.5"
            :class="i > 0 ? 'border-t border-zinc-100 dark:border-zinc-800' : ''"
          >
            <span class="text-[13px] text-zinc-500 dark:text-zinc-400">{{ r.label }}</span>
            <span class="text-[13px] tabular-nums">
              <span class="font-semibold text-zinc-700 dark:text-zinc-300">{{ fmt(r.count) }}</span>
              <span v-if="deltaText(r.delta)" class="ml-1 text-[11px]" :class="deltaClass(r.delta)">{{ deltaText(r.delta) }}</span>
            </span>
          </div>
        </div>
      </div>
    </section>

    <!-- ═══ Silver ═══ -->
    <section class="animate-in fade-in duration-300 delay-100">
      <div class="flex items-center gap-2 mb-3">
        <span class="text-xs font-semibold uppercase tracking-wide text-blue-400">Silver</span>
        <span class="text-[10px] font-semibold uppercase tracking-wide text-blue-400/80 bg-blue-500/10 px-1.5 py-0.5 rounded">Unified</span>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div
          v-for="group in silverGroups"
          :key="group.label"
          class="rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-4"
        >
          <div class="flex items-center gap-2 mb-3">
            <component :is="group.icon" class="w-4 h-4 text-blue-400" />
            <span class="text-sm font-semibold text-zinc-700 dark:text-zinc-300">{{ group.label }}</span>
          </div>
          <div class="grid gap-3" :class="group.gridClass">
            <div
              v-for="item in group.items"
              :key="item.key"
              class="rounded-lg p-3 bg-zinc-50 dark:bg-zinc-950 border border-zinc-200 dark:border-zinc-800 cursor-pointer hover:border-zinc-300 dark:hover:border-zinc-700 transition-all"
              @click="router.push(item.path)"
            >
              <div class="text-sm text-zinc-500 dark:text-zinc-400">{{ item.label }}</div>
              <div class="flex items-baseline gap-1.5 mt-1">
                <span class="text-xl font-bold text-zinc-900 dark:text-zinc-100 tabular-nums">{{ fmt(stats?.silver?.[item.key]?.count) }}</span>
                <span v-if="deltaText(stats?.silver?.[item.key]?.delta)" class="text-xs tabular-nums" :class="deltaClass(stats?.silver?.[item.key]?.delta)">{{ deltaText(stats?.silver?.[item.key]?.delta) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ═══ Gold ═══ -->
    <section class="animate-in fade-in duration-300 delay-200">
      <div class="flex items-center gap-2 mb-3">
        <span class="text-xs font-semibold uppercase tracking-wide text-amber-400">Gold</span>
        <span class="text-[10px] font-semibold uppercase tracking-wide text-amber-400/80 bg-amber-400/10 px-1.5 py-0.5 rounded">Insights</span>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="card in goldCards"
          :key="card.label"
          class="rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-4 cursor-pointer hover:border-zinc-300 dark:hover:border-zinc-700 hover:shadow-sm transition-all"
          @click="router.push(card.path)"
        >
          <div class="flex items-start gap-3">
            <div class="w-10 h-10 rounded-lg flex items-center justify-center shrink-0" :class="card.iconBg">
              <component :is="card.icon" class="w-5 h-5" :class="card.iconColor" />
            </div>
            <div>
              <div class="text-sm font-medium text-zinc-500 dark:text-zinc-400">{{ card.label }}</div>
              <div class="flex items-baseline gap-1.5">
                <span class="text-2xl font-semibold text-zinc-900 dark:text-zinc-100 tabular-nums">{{ fmt(card.total) }}</span>
                <span v-if="deltaText(card.delta)" class="text-xs tabular-nums" :class="deltaClass(card.delta, true)">{{ deltaText(card.delta) }}</span>
              </div>
              <div class="text-xs text-zinc-400 dark:text-zinc-500">{{ card.subtitle }}</div>
            </div>
          </div>
          <template v-if="card.rows.length">
            <div class="h-px bg-zinc-100 dark:bg-zinc-800 my-3" />
            <div
              v-for="(row, i) in card.rows"
              :key="row.label"
              class="flex justify-between items-center py-1.5"
              :class="i > 0 ? 'border-t border-zinc-100 dark:border-zinc-800' : ''"
            >
              <span class="text-[13px] text-zinc-500 dark:text-zinc-400">{{ row.label }}</span>
              <span class="text-[13px] tabular-nums">
                <span class="font-semibold" :class="row.color">{{ fmt(row.count) }}</span>
                <span v-if="deltaText(row.delta)" class="ml-1 text-[11px]" :class="deltaClass(row.delta, true)">{{ deltaText(row.delta) }}</span>
              </span>
            </div>
          </template>
        </div>
      </div>
    </section>
  </div>
</template>
