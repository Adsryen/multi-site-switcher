import { getSites, getSiteByKey } from '../core/sites/registry.js'
import { listAccounts, saveAccount, deleteAccount, getActiveAccountId, setActiveAccount } from '../core/storage/accounts.js'
import { switchAccount } from '../core/switch/switcher.js'

type Msg =
  | { type: 'getSites' }
  | { type: 'getAccounts'; siteKey: string }
  | { type: 'saveAccount'; siteKey: string; account: any }
  | { type: 'deleteAccount'; siteKey: string; accountId: string }
  | { type: 'switchAccount'; siteKey: string; accountId: string; options?: Record<string, any> }
  | { type: 'getSiteMeta'; siteKey: string }
  | { type: string; [k: string]: any }

chrome.runtime.onMessage.addListener((message: Msg, sender, sendResponse) => {
  ;(async () => {
    try {
      if (message && message.type === 'getSites') {
        const sites = getSites().map(s => ({ key: s.key, name: s.name }))
        sendResponse({ ok: true, data: sites })
        return
      }
      if (message && message.type === 'getAccounts') {
        const { siteKey } = message
        const accounts = await listAccounts(siteKey)
        const activeId = await getActiveAccountId(siteKey)
        sendResponse({ ok: true, data: { accounts, activeId } })
        return
      }
      if (message && message.type === 'saveAccount') {
        const { siteKey, account } = message
        const saved = await saveAccount(siteKey, account)
        sendResponse({ ok: true, data: saved })
        return
      }
      if (message && message.type === 'deleteAccount') {
        const { siteKey, accountId } = message
        await deleteAccount(siteKey, accountId)
        sendResponse({ ok: true })
        return
      }
      if (message && message.type === 'switchAccount') {
        const { siteKey, accountId, options } = message
        await setActiveAccount(siteKey, accountId)
        const result = await switchAccount(siteKey, accountId, options || {})
        sendResponse({ ok: true, data: result })
        return
      }
      if (message && message.type === 'getSiteMeta') {
        const { siteKey } = message
        const site = getSiteByKey(siteKey)
        if (!site) {
          sendResponse({ ok: false, error: 'site_not_found' })
          return
        }
        sendResponse({ ok: true, data: { key: site.key, name: site.name, loginUrl: site.loginUrl || '' } })
        return
      }
      sendResponse({ ok: false, error: 'unknown_message' })
    } catch (e: any) {
      sendResponse({ ok: false, error: String(e && e.message ? e.message : e) })
    }
  })()
  return true
})
