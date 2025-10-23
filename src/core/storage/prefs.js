const ROOT_KEY = 'mss_prefs';

async function getAll() {
  const obj = await chrome.storage.local.get([ROOT_KEY]);
  return obj[ROOT_KEY] || {};
}

async function setAll(root) {
  await chrome.storage.local.set({ [ROOT_KEY]: root });
}

function ensureSite(root, siteKey) {
  if (!root[siteKey]) root[siteKey] = {};
}

export async function getSitePrefs(siteKey) {
  const root = await getAll();
  ensureSite(root, siteKey);
  return root[siteKey];
}

export async function saveSitePrefs(siteKey, partial) {
  const root = await getAll();
  ensureSite(root, siteKey);
  root[siteKey] = { ...root[siteKey], ...partial };
  await setAll(root);
  return root[siteKey];
}
