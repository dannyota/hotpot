import { ref, readonly } from 'vue'
import router, { registerNavRoutes } from '@/router'

export interface NavItem {
  label: string
  icon?: string
  path?: string
  api?: string
  children?: NavItem[]
}

export interface UIConfig {
  name: string
  description: string
  title: string
  icon: string
  color: string
  nav: NavItem[]
}

const config = ref<UIConfig>({
  name: 'Hotpot',
  description: '',
  title: 'Hotpot',
  icon: '🍲',
  color: '',
  nav: [],
})
const loaded = ref(false)
const statusLines = ref<string[]>([])

function addStatus(msg: string) {
  const ts = new Date().toLocaleTimeString()
  statusLines.value = [...statusLines.value.slice(-9), `[${ts}] ${msg}`]
}

async function applyConfig(data: UIConfig) {
  config.value = data
  document.title = data.title

  if (data.icon && !data.icon.startsWith('http') && !data.icon.startsWith('/')) {
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><text y=".9em" font-size="90">${data.icon}</text></svg>`
    let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
    if (!link) {
      link = document.createElement('link')
      link.rel = 'icon'
      document.head.appendChild(link)
    }
    link.href = `data:image/svg+xml,${encodeURIComponent(svg)}`
  }

  if (data.nav?.length) {
    registerNavRoutes(data.nav)
    // The initial navigation resolved before dynamic routes were registered,
    // so the current route may be the catch-all. Re-resolve it now and wait
    // for the navigation to complete before setting loaded (which gates
    // <RouterView>), so the page mounts with the correct route meta.
    const current = router.currentRoute.value
    if (current.name === 'not-found') {
      await router.replace(current.fullPath)
    }
  }
}

/** Fetch UI config from the backend with retry. Called once on app mount. */
export async function loadUIConfig() {
  document.title = 'Loading'
  addStatus('Connecting to server...')

  while (true) {
    try {
      const res = await fetch('/api/v1/admin/ui-config')
      if (res.ok) {
        const data: UIConfig = await res.json()
        addStatus('Configuration loaded.')
        await applyConfig(data)
        loaded.value = true
        return
      }
      addStatus(`Server returned ${res.status}. Retrying in 3s...`)
    } catch {
      addStatus('Cannot reach server. Retrying in 3s...')
    }
    await new Promise((r) => setTimeout(r, 3000))
  }
}

/** Reactive UI config. */
export function useUIConfig() {
  return {
    config: readonly(config),
    loaded: readonly(loaded),
    statusLines: readonly(statusLines),
  }
}
