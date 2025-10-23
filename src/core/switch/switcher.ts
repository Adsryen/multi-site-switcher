import { getSiteByKey } from '../sites/registry.js'
import { getAccountById } from '../storage/accounts.js'
import type { SiteAdapter } from '../sites/types.js'

export async function switchAccount(siteKey: string, accountId: string, options: Record<string, any> = {}): Promise<any> {
  const site: SiteAdapter | null = getSiteByKey(siteKey)
  if (!site) throw new Error('site_not_found')
  const account = await getAccountById(siteKey, accountId)
  if (!account) throw new Error('account_not_found')
  if (typeof site.logout === 'function') {
    await site.logout(options)
  }
  const res = await site.login(account, options)
  return res || { ok: true }
}
