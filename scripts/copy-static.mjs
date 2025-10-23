// Copy static assets (manifest, HTML/CSS/JS/images, public/) into dist
// Keep folder structure so manifest paths remain valid.

import { promises as fs } from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const repoRoot = path.resolve(__dirname, '..');
const distDir = path.join(repoRoot, 'dist');

const exts = new Set(['.html', '.css', '.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg', '.ico']);

async function ensureDir(p) {
  await fs.mkdir(p, { recursive: true });
}

async function copyFile(src, dest) {
  await ensureDir(path.dirname(dest));
  await fs.copyFile(src, dest);
}

async function copyDirRecursive(srcDir, destDir, filterExts = null) {
  const entries = await fs.readdir(srcDir, { withFileTypes: true });
  for (const entry of entries) {
    const srcPath = path.join(srcDir, entry.name);
    const destPath = path.join(destDir, entry.name);
    if (entry.isDirectory()) {
      await copyDirRecursive(srcPath, destPath, filterExts);
    } else if (entry.isFile()) {
      const ext = path.extname(entry.name).toLowerCase();
      if (!filterExts || filterExts.has(ext)) {
        await copyFile(srcPath, destPath);
      }
    }
  }
}

async function main() {
  await ensureDir(distDir);
  // manifest.json -> dist/
  const manifestSrc = path.join(repoRoot, 'manifest.json');
  try {
    await copyFile(manifestSrc, path.join(distDir, 'manifest.json'));
  } catch {
    // ignore if missing
  }

  // Copy static assets from src/ (no TS/JS)
  const srcDir = path.join(repoRoot, 'src');
  try {
    await copyDirRecursive(srcDir, path.join(distDir, 'src'), exts);
  } catch {
    // ignore if missing
  }

  // Copy public/ if exists
  const publicDir = path.join(repoRoot, 'public');
  try {
    const stat = await fs.stat(publicDir);
    if (stat.isDirectory()) {
      await copyDirRecursive(publicDir, path.join(distDir, 'public'));
    }
  } catch {
    // ignore if missing
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
