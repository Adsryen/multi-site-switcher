import { generateId } from '../utils/id.js';

const ROOT_KEY = 'mss_accounts';

async function getAll() {
  const obj = await chrome.storage.local.get([ROOT_KEY]);
  return obj[ROOT_KEY] || {};
}

async function setAll(root) {
  await chrome.storage.local.set({ [ROOT_KEY]: root });
}

function ensureSite(root, siteKey) {
  if (!root[siteKey]) {
    root[siteKey] = { accounts: [], activeId: null };
  }
}

export async function listAccounts(siteKey) {
  const root = await getAll();
  ensureSite(root, siteKey);
  return root[siteKey].accounts || [];
}

export async function getActiveAccountId(siteKey) {
  const root = await getAll();
  ensureSite(root, siteKey);
  return root[siteKey].activeId || null;
}

export async function setActiveAccount(siteKey, accountId) {
  const root = await getAll();
  ensureSite(root, siteKey);
  root[siteKey].activeId = accountId || null;
  await setAll(root);
}

export async function saveAccount(siteKey, account) {
  const root = await getAll();
  ensureSite(root, siteKey);
  const list = root[siteKey].accounts || [];
  if (account.id) {
    const idx = list.findIndex(a => a.id === account.id);
    if (idx >= 0) {
      list[idx] = { ...list[idx], ...account };
    } else {
      list.push(account);
    }
  } else {
    account.id = generateId('acc');
    list.push(account);
  }
  root[siteKey].accounts = list;
  await setAll(root);
  return account;
}

export async function deleteAccount(siteKey, accountId) {
  const root = await getAll();
  ensureSite(root, siteKey);
  const list = root[siteKey].accounts || [];
  root[siteKey].accounts = list.filter(a => a.id !== accountId);
  if (root[siteKey].activeId === accountId) {
    root[siteKey].activeId = null;
  }
  await setAll(root);
}

export async function getAccountById(siteKey, accountId) {
  const list = await listAccounts(siteKey);
  return list.find(a => a.id === accountId) || null;
}
