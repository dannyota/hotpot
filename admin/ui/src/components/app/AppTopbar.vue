<script setup lang="ts">
import { Moon, Sun } from 'lucide-vue-next'
import BreadcrumbNav from './BreadcrumbNav.vue'
import PageSearch from './PageSearch.vue'
import AppNotifications from './AppNotifications.vue'
import AppTimezone from './AppTimezone.vue'

defineProps<{ dark: boolean }>()
defineEmits<{ 'toggle-dark': [] }>()
</script>

<template>
  <div class="h-14 border-b border-zinc-200 dark:border-zinc-800 flex items-center px-4 gap-4 bg-white dark:bg-zinc-950">
    <BreadcrumbNav />
    <PageSearch />
    <div class="flex items-center gap-1 shrink-0">
      <AppTimezone />
      <AppNotifications />
      <button
        class="p-2 rounded-md text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
        @click="$emit('toggle-dark')"
      >
        <Sun v-if="dark" class="w-4 h-4" />
        <Moon v-else class="w-4 h-4" />
      </button>
    </div>
  </div>
</template>
