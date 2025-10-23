import type { SiteAdapter } from '../../core/sites/types.js'

const key = 'example'
const name = 'Example Site'
const loginUrl = 'https://example.com/login'

function matches(url: string): boolean {
  try { return new URL(url).hostname.includes('example.com') } catch { return false }
}

async function logout(opts: Record<string, any> = {}): Promise<any> {
  // TODO: implement site-specific logout (clear cookies, call logout URL, etc.)
  return { ok: true, step: 'logout_skipped' }
}

async function login(account: Record<string, any>, opts: Record<string, any> = {}): Promise<any> {
  const target = loginUrl
  const tabs = await chrome.tabs.query({})
  const existing = tabs.find((t: any) => t.url && matches(t.url))
  if (existing) {
    await chrome.tabs.update(existing.id, { url: target, active: true })
  } else {
    await chrome.tabs.create({ url: target, active: true })
  }
  return { ok: true, step: 'navigated_to_login', loginUrl: target }
}

const adapter: SiteAdapter = { key, name, loginUrl, matches, login, logout }
export default adapter
