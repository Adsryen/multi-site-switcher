const siteSelect = document.getElementById('siteSelect');
const metaKey = document.getElementById('metaKey');
const metaName = document.getElementById('metaName');
const metaLogin = document.getElementById('metaLogin');

function send(msg) {
  return new Promise((resolve) => {
    chrome.runtime.sendMessage(msg, (res) => resolve(res));
  });
}

async function loadSites() {
  const res = await send({ type: 'getSites' });
  if (!res || !res.ok) return [];
  const sites = res.data || [];
  siteSelect.innerHTML = '';
  for (const s of sites) {
    const opt = document.createElement('option');
    opt.value = s.key;
    opt.textContent = s.name || s.key;
    siteSelect.appendChild(opt);
  }
  return sites;
}

async function loadSiteMeta(siteKey) {
  const res = await send({ type: 'getSiteMeta', siteKey });
  if (!res || !res.ok) return;
  const { key, name, loginUrl } = res.data || {};
  metaKey.textContent = key || '-';
  metaName.textContent = name || '-';
  metaLogin.textContent = loginUrl || '-';
  if (loginUrl) metaLogin.href = loginUrl;
}

siteSelect.addEventListener('change', async () => {
  const key = siteSelect.value;
  if (key) await loadSiteMeta(key);
});

(async function init() {
  const sites = await loadSites();
  if (sites.length > 0) {
    await loadSiteMeta(sites[0].key);
  }
})();
