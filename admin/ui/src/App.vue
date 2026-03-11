<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { Moon, Sun } from 'lucide-vue-next'
import AppSidebar from '@/components/app/AppSidebar.vue'
import AppTopbar from '@/components/app/AppTopbar.vue'
import AppToasts from '@/components/app/AppToasts.vue'
import { loadUIConfig, useUIConfig } from '@/composables/useUIConfig'
import { TooltipProvider } from 'reka-ui'

const { loaded, statusLines } = useUIConfig()

const dark = ref(localStorage.getItem('theme-dark') !== 'false')

watch(dark, (v) => {
  localStorage.setItem('theme-dark', String(v))
  document.documentElement.classList.toggle('dark', v)
}, { immediate: true })

onMounted(loadUIConfig)
</script>

<template>
  <!-- Loading screen while waiting for server config -->
  <div v-if="!loaded" class="relative flex items-center justify-center h-screen bg-zinc-50 dark:bg-zinc-950">
    <button
      class="absolute top-4 right-4 p-2 rounded-md text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
      @click="dark = !dark"
    >
      <Sun v-if="dark" class="w-4 h-4" />
      <Moon v-else class="w-4 h-4" />
    </button>
    <div class="w-full max-w-md px-6">
      <h1 class="text-2xl font-semibold text-zinc-700 dark:text-zinc-200 mb-6 text-center">Loading</h1>
      <div class="bg-zinc-900 dark:bg-zinc-900 rounded-lg p-4 font-mono text-xs leading-relaxed h-52 overflow-y-auto">
        <p
          v-for="(line, i) in statusLines"
          :key="i"
          class="text-zinc-400"
          :class="{ 'text-zinc-200': i === statusLines.length - 1 }"
        >{{ line }}</p>
      </div>
    </div>
  </div>

  <!-- Main app once config is loaded -->
  <TooltipProvider v-else :delay-duration="400">
    <div class="flex h-screen overflow-hidden bg-zinc-50 dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100">
      <AppSidebar />
      <div class="flex-1 flex flex-col min-w-0">
        <AppTopbar :dark="dark" @toggle-dark="dark = !dark" />
        <main class="flex-1 overflow-auto">
          <RouterView />
        </main>
      </div>
    </div>
    <AppToasts />
  </TooltipProvider>
</template>
