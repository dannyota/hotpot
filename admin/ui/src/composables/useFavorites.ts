import { ref, computed, watch } from 'vue'
import { useUIConfig, type NavItem } from './useUIConfig'

export interface FavoriteLeaf {
  type: 'leaf'
  key: string // same as path
  label: string
  path: string
  context: string
}

export interface FavoriteCategory {
  type: 'category'
  key: string // nodeKey, e.g. "Silver/Inventory"
  label: string
  context: string
  children: { label: string; path: string }[]
}

export type FavoriteEntry = FavoriteLeaf | FavoriteCategory

const keys = ref<string[]>(loadFavorites())

function loadFavorites(): string[] {
  try {
    const raw = localStorage.getItem('hotpot-favorites')
    if (raw) return JSON.parse(raw)
  } catch { /* ignore */ }
  return []
}

watch(keys, (v) => {
  localStorage.setItem('hotpot-favorites', JSON.stringify(v))
}, { deep: true })

export function useFavorites() {
  const { config } = useUIConfig()

  function isFavorite(key: string): boolean {
    return keys.value.includes(key)
  }

  function toggleFavorite(key: string) {
    const i = keys.value.indexOf(key)
    if (i >= 0) {
      keys.value = keys.value.filter((k) => k !== key)
    } else {
      keys.value = [...keys.value, key]
    }
  }

  /** Build context string from breadcrumb: drop layer (first) and label (last). */
  function buildLeafContext(breadcrumb: string[]): string {
    const middle = breadcrumb.slice(1, -1)
    return middle.join(' ')
  }

  /** Collect all leaf pages under a nav item recursively. */
  function collectLeaves(item: NavItem): { label: string; path: string }[] {
    if (!item.children) {
      return item.path ? [{ label: item.label, path: item.path }] : []
    }
    return item.children.flatMap(collectLeaves)
  }

  const favorites = computed<FavoriteEntry[]>(() => {
    const leafMap = new Map<string, FavoriteLeaf>()
    const categoryMap = new Map<string, FavoriteCategory>()

    function walk(items: NavItem[], trail: string[], parentKey?: string) {
      for (const item of items) {
        const nodeKey = parentKey ? `${parentKey}/${item.label}` : item.label
        if (item.children) {
          categoryMap.set(nodeKey, {
            type: 'category',
            key: nodeKey,
            label: item.label,
            context: trail.join(' '),
            children: collectLeaves(item),
          })
          walk(item.children, [...trail, item.label], nodeKey)
        } else if (item.path) {
          leafMap.set(item.path, {
            type: 'leaf',
            key: item.path,
            label: item.label,
            path: item.path,
            context: buildLeafContext([...trail, item.label]),
          })
        }
      }
    }
    walk(config.value.nav, [])

    return keys.value
      .map((k) => leafMap.get(k) ?? categoryMap.get(k))
      .filter((x): x is FavoriteEntry => !!x)
  })

  return { favorites, isFavorite, toggleFavorite }
}
