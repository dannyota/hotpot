<script setup lang="ts">
import { computed, provide } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  LayoutDashboard, Database, Server, ShieldAlert, Cloud, Shield,
  HardDrive, Globe, Key, Container, Monitor, Folder, Layers,
  ChevronDown, ChevronLeft, ChevronRight, Star,
} from 'lucide-vue-next'
import { useUIConfig, type NavItem } from '@/composables/useUIConfig'
import { useSidebar } from '@/composables/useSidebar'
import { useFavorites } from '@/composables/useFavorites'
import NavTreeNode from './NavTreeNode.vue'

const route = useRoute()
const router = useRouter()
const { config } = useUIConfig()
const { open, expanded, togglePin, toggleGroup, closeTemp } = useSidebar()
const { favorites, isFavorite, toggleFavorite } = useFavorites()

provide('nav-expanded', expanded)

const collapsed = computed(() => !open.value)

const iconMap: Record<string, any> = {
  'layout-dashboard': LayoutDashboard,
  'database': Database,
  'server': Server,
  'shield-alert': ShieldAlert,
  'cloud': Cloud,
  'shield': Shield,
  'hard-drive': HardDrive,
  'globe': Globe,
  'key': Key,
  'container': Container,
  'monitor': Monitor,
  'folder': Folder,
  'layers': Layers,
}

function resolveIcon(name?: string) {
  if (!name) return null
  return iconMap[name] ?? Folder
}

function leafCount(item: NavItem): number {
  if (!item.children) return 1
  return item.children.reduce((n, c) => n + leafCount(c), 0)
}

function isActive(path?: string) {
  return path === route.path
}

function navigate(path?: string) {
  if (path) router.push(path)
}

const isImageIcon = computed(() => {
  const icon = config.value.icon
  return icon.startsWith('http://') || icon.startsWith('https://') || icon.startsWith('/')
})

const isLetterIcon = computed(() => {
  const icon = config.value.icon
  return icon.length === 1 && /^[a-zA-Z0-9]$/.test(icon)
})
</script>

<template>
  <div
    class="flex flex-col border-r border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-950 transition-all duration-200 shrink-0"
    :class="collapsed ? 'w-14' : 'w-72 max-w-80'"
    @mouseleave="closeTemp"
  >
    <!-- Logo -->
    <div class="flex items-center gap-3 px-4 h-14 border-b border-zinc-200 dark:border-zinc-800">
      <div v-if="isImageIcon" class="w-8 h-8 rounded-lg overflow-hidden shrink-0">
        <img :src="config.icon" :alt="config.name" class="w-full h-full object-cover" />
      </div>
      <div
        v-else-if="isLetterIcon"
        class="w-8 h-8 rounded-lg flex items-center justify-center text-white text-sm font-bold shrink-0"
        :style="config.color ? { backgroundColor: config.color } : {}"
        :class="!config.color && 'bg-gradient-to-br from-orange-500 to-red-600'"
      >
        {{ config.icon }}
      </div>
      <span v-else class="text-2xl leading-none shrink-0">{{ config.icon }}</span>
      <div v-if="!collapsed" class="min-w-0">
        <span class="font-semibold text-base text-zinc-900 dark:text-zinc-100 tracking-tight block truncate">{{ config.name }}</span>
        <span v-if="config.description" class="text-[11px] text-zinc-400 dark:text-zinc-500 block truncate leading-tight">{{ config.description }}</span>
      </div>
    </div>

    <!-- Nav -->
    <nav class="flex-1 overflow-y-auto px-2 py-3 space-y-0.5">
      <!-- Favorites section -->
      <template v-if="!collapsed && favorites.length">
        <div class="mb-1">
          <div class="flex items-center gap-2 px-3 py-1.5 text-[11px] font-medium uppercase tracking-wider text-zinc-400 dark:text-zinc-500">
            <Star class="w-3 h-3" fill="currentColor" />
            <span>Favorites</span>
          </div>
          <template v-for="fav in favorites" :key="fav.key">
            <!-- Leaf favorite -->
            <button
              v-if="fav.type === 'leaf'"
              class="group flex items-center gap-2 w-full rounded-md text-sm transition-colors px-3 py-1.5"
              :class="isActive((fav as any).path)
                ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
                : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:text-zinc-100 dark:hover:bg-zinc-800'"
              @click="navigate((fav as any).path)"
            >
              <span class="truncate">{{ fav.label }}</span>
              <span
                v-if="fav.context"
                class="text-[11px] shrink-0 text-zinc-400 dark:text-zinc-500"
              >({{ fav.context }})</span>
              <span
                class="shrink-0 ml-auto p-0.5 rounded opacity-0 group-hover:opacity-100 transition-colors text-zinc-400 hover:text-amber-400"
                @click.stop="toggleFavorite(fav.key)"
              >
                <Star class="w-3 h-3" fill="currentColor" />
              </span>
            </button>

            <!-- Category favorite -->
            <div v-else>
              <div
                class="group flex items-center gap-2 w-full rounded-md text-[13px] px-3 py-1.5 text-zinc-500 dark:text-zinc-400"
              >
                <span class="truncate font-medium">{{ fav.label }}</span>
                <span
                  v-if="fav.context"
                  class="text-[11px] shrink-0 text-zinc-400 dark:text-zinc-500"
                >({{ fav.context }})</span>
                <span
                  class="shrink-0 ml-auto p-0.5 rounded opacity-0 group-hover:opacity-100 transition-colors text-zinc-400 hover:text-amber-400"
                  @click.stop="toggleFavorite(fav.key)"
                >
                  <Star class="w-3 h-3" fill="currentColor" />
                </span>
              </div>
              <button
                v-for="child in (fav as any).children"
                :key="child.path"
                class="flex items-center w-full rounded-md text-sm transition-colors pl-6 pr-3 py-1"
                :class="isActive(child.path)
                  ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
                  : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:text-zinc-100 dark:hover:bg-zinc-800'"
                @click="navigate(child.path)"
              >
                <span class="truncate">{{ child.label }}</span>
              </button>
            </div>
          </template>
        </div>
        <div class="mx-3 border-t border-zinc-200 dark:border-zinc-800 mb-1" />
      </template>

      <template v-for="item in config.nav" :key="item.label">
        <!-- Top-level leaf (e.g., Dashboard) -->
        <button
          v-if="!item.children"
          class="group flex items-center gap-2 w-full rounded-md text-sm transition-colors px-3 py-2"
          :class="isActive(item.path)
            ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
            : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:text-zinc-100 dark:hover:bg-zinc-800'"
          @click="navigate(item.path)"
        >
          <component :is="resolveIcon(item.icon)" v-if="item.icon" class="w-4 h-4 shrink-0" />
          <template v-if="!collapsed">
            <span class="truncate flex-1 text-left">{{ item.label }}</span>
            <span
              class="shrink-0 p-0.5 rounded transition-colors"
              :class="item.path && isFavorite(item.path)
                ? 'text-amber-400 opacity-100'
                : 'opacity-0 group-hover:opacity-100 text-zinc-400 hover:text-amber-400'"
              @click.stop="item.path && toggleFavorite(item.path)"
            >
              <Star class="w-3 h-3" :fill="item.path && isFavorite(item.path) ? 'currentColor' : 'none'" />
            </span>
          </template>
        </button>

        <!-- Top-level group (e.g., Bronze, Silver, Gold) -->
        <div v-else>
          <button
            class="group flex items-center gap-2 w-full rounded-md text-sm transition-colors px-3 py-2 text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:text-zinc-100 dark:hover:bg-zinc-800"
            @click="toggleGroup(item.label)"
          >
            <component :is="resolveIcon(item.icon)" v-if="item.icon" class="w-4 h-4 shrink-0" />
            <template v-if="!collapsed">
              <span class="truncate flex-1 text-left">{{ item.label }}</span>
              <span class="text-[11px] text-zinc-400 dark:text-zinc-500 tabular-nums">{{ leafCount(item) }}</span>
              <span
                class="shrink-0 p-0.5 rounded transition-colors"
                :class="isFavorite(item.label)
                  ? 'text-amber-400 opacity-100'
                  : 'opacity-0 group-hover:opacity-100 text-zinc-400 hover:text-amber-400'"
                @click.stop="toggleFavorite(item.label)"
              >
                <Star class="w-3 h-3" :fill="isFavorite(item.label) ? 'currentColor' : 'none'" />
              </span>
              <ChevronDown
                class="w-3.5 h-3.5 transition-transform shrink-0"
                :class="expanded.has(item.label) && 'rotate-180'"
              />
            </template>
          </button>

          <div v-if="!collapsed && expanded.has(item.label)" class="mt-0.5 space-y-0.5">
            <NavTreeNode
              v-for="child in item.children"
              :key="child.label"
              :item="child"
              :depth="1"
              :parent-key="item.label"
            />
          </div>
        </div>
      </template>
    </nav>

    <!-- Collapse toggle -->
    <div class="border-t border-zinc-200 dark:border-zinc-800 p-2">
      <button
        class="flex items-center justify-center w-full p-2 rounded-md text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
        @click="togglePin"
      >
        <ChevronLeft v-if="!collapsed" class="w-4 h-4" />
        <ChevronRight v-else class="w-4 h-4" />
      </button>
    </div>
  </div>
</template>
