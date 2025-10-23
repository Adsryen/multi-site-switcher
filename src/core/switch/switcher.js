import { getSiteByKey } from '../sites/registry.js';
import { getAccountById } from '../storage/accounts.js';

export async function switchAccount(siteKey, accountId, options = {}) {
  const site = getSiteByKey(siteKey);
  if (!site) throw new Error('site_not_found');
  const account = await getAccountById(siteKey, accountId);
  if (!account) throw new Error('account_not_found');
  if (typeof site.logout === 'function') {
    await site.logout(options);
  }
  const res = await site.login(account, options);
  return res || { ok: true };
}
