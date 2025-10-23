const ROOT_KEY = 'mss_prefs'

type Root = Record<string, Record<string, any>>

async function getAll(): Promise<Root> {
  const obj = await chrome.storage.local.get([ROOT_KEY])
  return (obj?.[ROOT_KEY] as Root) || {}
}

async function setAll(root: Root): Promise<void> {
  await chrome.storage.local.set({ [ROOT_KEY]: root })
}

function ensureSite(root: Root, siteKey: string): void {
  if (!root[siteKey]) root[siteKey] = {}
}

export async function getSitePrefs(siteKey: string): Promise<Record<string, any>> {
  const root = await getAll()
  ensureSite(root, siteKey)
  return root[siteKey]
}

export async function saveSitePrefs(siteKey: string, partial: Record<string, any>): Promise<Record<string, any>> {
  const root = await getAll()
  ensureSite(root, siteKey)
  root[siteKey] = { ...root[siteKey], ...partial }
  await setAll(root)
  return root[siteKey]
}
