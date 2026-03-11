<script setup lang="ts">
import { ref, computed } from 'vue'
import { Bell, X, Check, CheckCheck, Trash2, AlertCircle, AlertTriangle, Info } from 'lucide-vue-next'
import {
  useNotifications,
  shortenSource,
  formatNotificationTime,
  type Notification,
} from '@/composables/useNotifications'

const { notifications, unreadCount, markRead, markAllRead, remove, clearAll } = useNotifications()

const open = ref(false)

const typeIcon = { error: AlertCircle, warning: AlertTriangle, info: Info }
const typeColor = {
  error: 'text-red-500',
  warning: 'text-amber-500',
  info: 'text-blue-500',
}

const hasNotifications = computed(() => notifications.value.length > 0)
</script>

<template>
  <div class="relative">
    <!-- Bell button -->
    <button
      class="p-2 rounded-md text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors relative"
      @click.stop="open = !open"
    >
      <Bell class="w-4 h-4" />
      <span
        v-if="unreadCount > 0"
        class="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] flex items-center justify-center rounded-full bg-red-500 text-white text-[10px] font-medium px-1"
      >
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
    </button>

    <!-- Backdrop -->
    <Teleport to="body">
      <div v-if="open" class="fixed inset-0 z-40" @click="open = false" />
    </Teleport>

    <!-- Dropdown -->
    <div
      v-if="open"
      class="absolute right-0 top-full mt-1 z-50 w-96 rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 shadow-xl"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-zinc-200 dark:border-zinc-700">
        <span class="text-sm font-medium text-zinc-900 dark:text-zinc-100">Notifications</span>
        <div v-if="hasNotifications" class="flex items-center gap-1">
          <button
            v-if="unreadCount > 0"
            class="p-1 rounded text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
            title="Mark all read"
            @click="markAllRead"
          >
            <CheckCheck class="w-4 h-4" />
          </button>
          <button
            class="p-1 rounded text-zinc-400 hover:text-red-500 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
            title="Clear all"
            @click="clearAll"
          >
            <Trash2 class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- List -->
      <div class="max-h-96 overflow-y-auto">
        <template v-if="hasNotifications">
          <div
            v-for="n in notifications"
            :key="n.id"
            class="flex items-start gap-3 px-4 py-3 border-b border-zinc-100 dark:border-zinc-800 last:border-b-0 cursor-pointer transition-colors"
            :class="n.read
              ? 'bg-transparent hover:bg-zinc-50 dark:hover:bg-zinc-800/50'
              : 'bg-zinc-50 dark:bg-zinc-800/30 hover:bg-zinc-100 dark:hover:bg-zinc-800/60'"
            @click="!n.read && markRead(n.id)"
          >
            <!-- Icon -->
            <component :is="typeIcon[n.type]" class="w-4 h-4 mt-0.5 shrink-0" :class="typeColor[n.type]" />

            <!-- Content -->
            <div class="flex-1 min-w-0">
              <p class="text-sm text-zinc-900 dark:text-zinc-100 leading-snug">{{ n.message }}</p>
              <div class="flex items-center gap-2 mt-1">
                <span class="text-[11px] text-zinc-400 dark:text-zinc-500 truncate">{{ shortenSource(n.source) }}</span>
                <span class="text-[11px] text-zinc-400 dark:text-zinc-500 shrink-0">{{ formatNotificationTime(n.timestamp) }}</span>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-0.5 shrink-0">
              <button
                v-if="!n.read"
                class="p-1 rounded text-zinc-300 hover:text-zinc-500 dark:text-zinc-600 dark:hover:text-zinc-400 transition-colors"
                title="Mark read"
                @click.stop="markRead(n.id)"
              >
                <Check class="w-3.5 h-3.5" />
              </button>
              <button
                class="p-1 rounded text-zinc-300 hover:text-red-500 dark:text-zinc-600 dark:hover:text-red-400 transition-colors"
                title="Dismiss"
                @click.stop="remove(n.id)"
              >
                <X class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </template>

        <!-- Empty state -->
        <div v-else class="px-4 py-10 text-center">
          <Bell class="w-8 h-8 mx-auto text-zinc-300 dark:text-zinc-600 mb-2" />
          <p class="text-sm text-zinc-400 dark:text-zinc-500">No notifications</p>
        </div>
      </div>
    </div>
  </div>
</template>
