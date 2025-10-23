const siteSelect = document.getElementById('siteSelect') as HTMLSelectElement
const metaKey = document.getElementById('metaKey') as HTMLDivElement
const metaName = document.getElementById('metaName') as HTMLDivElement
const metaLogin = document.getElementById('metaLogin') as HTMLAnchorElement

type RpcResponse<T=any> = { ok: boolean; data?: T; error?: string }

function send<T=any>(msg: any): Promise<RpcResponse<T>> {
  return new Promise((resolve) => {
    chrome.runtime.sendMessage(msg, (res: RpcResponse<T>) => resolve(res))
  })
}

async function loadSites(): Promise<any[]> {
  const res = await send<any[]>({ type: 'getSites' })
  if (!res || !res.ok) return []
  const sites = res.data || []
  siteSelect.innerHTML = ''
  for (const s of sites) {
    const opt = document.createElement('option')
    opt.value = s.key
    opt.textContent = s.name || s.key
    siteSelect.appendChild(opt)
  }
  return sites
}

async function loadSiteMeta(siteKey: string): Promise<void> {
  const res = await send<{ key: string; name: string; loginUrl?: string }>({ type: 'getSiteMeta', siteKey })
  if (!res || !res.ok) return
  const { key, name, loginUrl } = (res.data || {}) as any
  metaKey.textContent = key || '-'
  metaName.textContent = name || '-'
  metaLogin.textContent = loginUrl || '-'
  if (loginUrl) metaLogin.href = loginUrl
}

siteSelect.addEventListener('change', async () => {
  const key = siteSelect.value
  if (key) await loadSiteMeta(key)
})

;(async function init() {
  const sites = await loadSites()
  if (sites.length > 0) {
    await loadSiteMeta(sites[0].key)
  }
})()
