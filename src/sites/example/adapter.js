const key = 'example';
const name = 'Example Site';
const loginUrl = 'https://example.com/login';

function matches(url) {
  try { return new URL(url).hostname.includes('example.com'); } catch { return false; }
}

async function logout(opts = {}) {
  // TODO: 执行站点登出逻辑（清理 Cookie、访问登出页等）。此处为占位。
  return { ok: true, step: 'logout_skipped' };
}

async function login(account, opts = {}) {
  // 最小实现：打开或跳转到登录页，后续可注入脚本自动填充
  const target = loginUrl;
  const tabs = await chrome.tabs.query({});
  const existing = tabs.find(t => t.url && matches(t.url));
  if (existing) {
    await chrome.tabs.update(existing.id, { url: target, active: true });
  } else {
    await chrome.tabs.create({ url: target, active: true });
  }
  return { ok: true, step: 'navigated_to_login', loginUrl: target };
}

export default { key, name, loginUrl, matches, login, logout };
