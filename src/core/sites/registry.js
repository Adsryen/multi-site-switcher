import exampleAdapter from '../../sites/example/adapter.js';

const adapters = [
  exampleAdapter,
];

export function getSites() {
  return adapters;
}

export function getSiteByKey(key) {
  return adapters.find(a => a.key === key) || null;
}
