import type { SiteAdapter } from './types.js'
import exampleAdapter from '../../sites/example/adapter.js'

const adapters: SiteAdapter[] = [
  exampleAdapter,
]

export function getSites(): SiteAdapter[] {
  return adapters
}

export function getSiteByKey(key: string): SiteAdapter | null {
  return adapters.find(a => a.key === key) || null
}
