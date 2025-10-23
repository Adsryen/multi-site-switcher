import { generateId } from '../utils/id.js'

export interface Account {
  id?: string
  username?: string
  password?: string
  extra?: Record<string, any>
  [key: string]: any
}

type AccountStore = { accounts: Account[]; activeId: string | null }
type Root = Record<string, AccountStore>

const ROOT_KEY = 'mss_accounts'

async function getAll(): Promise<Root> {
  const obj = await chrome.storage.local.get([ROOT_KEY])
  return (obj?.[ROOT_KEY] as Root) || {}
}

async function setAll(root: Root): Promise<void> {
  await chrome.storage.local.set({ [ROOT_KEY]: root })
}

function ensureSite(root: Root, siteKey: string): void {
  if (!root[siteKey]) {
    root[siteKey] = { accounts: [], activeId: null }
  }
}

export async function listAccounts(siteKey: string): Promise<Account[]> {
  const root = await getAll()
  ensureSite(root, siteKey)
  return root[siteKey].accounts || []
}

export async function getActiveAccountId(siteKey: string): Promise<string | null> {
  const root = await getAll()
  ensureSite(root, siteKey)
  return root[siteKey].activeId || null
}

export async function setActiveAccount(siteKey: string, accountId: string | null): Promise<void> {
  const root = await getAll()
  ensureSite(root, siteKey)
  root[siteKey].activeId = accountId || null
  await setAll(root)
}

export async function saveAccount(siteKey: string, account: Account): Promise<Account> {
  const root = await getAll()
  ensureSite(root, siteKey)
  const list = root[siteKey].accounts || []
  if (account.id) {
    const idx = list.findIndex(a => a.id === account.id)
    if (idx >= 0) {
      list[idx] = { ...list[idx], ...account }
    } else {
      list.push(account)
    }
  } else {
    account.id = generateId('acc')
    list.push(account)
  }
  root[siteKey].accounts = list
  await setAll(root)
  return account
}

export async function deleteAccount(siteKey: string, accountId: string): Promise<void> {
  const root = await getAll()
  ensureSite(root, siteKey)
  const list = root[siteKey].accounts || []
  root[siteKey].accounts = list.filter(a => a.id !== accountId)
  if (root[siteKey].activeId === accountId) {
    root[siteKey].activeId = null
  }
  await setAll(root)
}

export async function getAccountById(siteKey: string, accountId: string): Promise<Account | null> {
  const list = await listAccounts(siteKey)
  return list.find(a => a.id === accountId) || null
}
