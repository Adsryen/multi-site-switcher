const siteSelect = document.getElementById('siteSelect') as HTMLSelectElement
const accountsList = document.getElementById('accountsList') as HTMLUListElement
const refreshBtn = document.getElementById('refreshAccounts') as HTMLButtonElement
const openOptions = document.getElementById('openOptions') as HTMLAnchorElement

const form = document.getElementById('addAccountForm') as HTMLFormElement
const fieldId = document.getElementById('accountId') as HTMLInputElement
const fieldUser = document.getElementById('username') as HTMLInputElement
const fieldPass = document.getElementById('password') as HTMLInputElement
const resetFormBtn = document.getElementById('resetForm') as HTMLButtonElement

type RpcResponse<T=any> = { ok: boolean; data?: T; error?: string }

type Account = { id?: string; username?: string; password?: string; [k: string]: any }

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

async function loadAccounts(): Promise<void> {
  const siteKey = siteSelect.value
  if (!siteKey) return
  const res = await send<{ accounts: Account[]; activeId: string | null }>({ type: 'getAccounts', siteKey })
  if (!res || !res.ok) return
  const { accounts, activeId } = res.data as any
  renderAccounts(accounts || [], activeId || null)
}

function renderAccounts(accounts: Account[], activeId: string | null): void {
  accountsList.innerHTML = ''
  accounts.forEach(acc => {
    const li = document.createElement('li')
    li.className = 'account-item' + (acc.id === activeId ? ' active' : '')

    const info = document.createElement('div')
    info.className = 'info'
    info.textContent = acc.username || '(未命名)'

    const actions = document.createElement('div')
    actions.className = 'actions'

    const useBtn = document.createElement('button')
    useBtn.textContent = '切换'
    useBtn.addEventListener('click', async () => {
      const siteKey = siteSelect.value
      await send({ type: 'switchAccount', siteKey, accountId: acc.id })
      await loadAccounts()
    })

    const editBtn = document.createElement('button')
    editBtn.textContent = '编辑'
    editBtn.addEventListener('click', () => {
      fieldId.value = acc.id || ''
      fieldUser.value = acc.username || ''
      fieldPass.value = acc.password || ''
      fieldUser.focus()
    })

    const delBtn = document.createElement('button')
    delBtn.textContent = '删除'
    delBtn.addEventListener('click', async () => {
      const siteKey = siteSelect.value
      await send({ type: 'deleteAccount', siteKey, accountId: acc.id })
      if (fieldId.value === acc.id) clearForm()
      await loadAccounts()
    })

    actions.appendChild(useBtn)
    actions.appendChild(editBtn)
    actions.appendChild(delBtn)

    li.appendChild(info)
    li.appendChild(actions)
    accountsList.appendChild(li)
  })
}

function clearForm(): void {
  fieldId.value = ''
  fieldUser.value = ''
  fieldPass.value = ''
}

form.addEventListener('submit', async (e: Event) => {
  e.preventDefault()
  const siteKey = siteSelect.value
  const account: Account = {
    id: fieldId.value || undefined,
    username: fieldUser.value.trim(),
    password: fieldPass.value,
  }
  if (!account.username) return
  const res = await send<Account>({ type: 'saveAccount', siteKey, account })
  if (res && res.ok) {
    clearForm()
    await loadAccounts()
  }
})

resetFormBtn.addEventListener('click', clearForm)
refreshBtn.addEventListener('click', loadAccounts)

openOptions.addEventListener('click', async (e: Event) => {
  e.preventDefault()
  chrome.runtime.openOptionsPage()
})

siteSelect.addEventListener('change', loadAccounts)

;(async function init() {
  const sites = await loadSites()
  if (sites.length > 0) {
    await loadAccounts()
  }
})()
