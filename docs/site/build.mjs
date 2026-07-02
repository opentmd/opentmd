#!/usr/bin/env node
/**
 * Build OpenTMD docs site from markdown sources in ../
 * Output: docs/site/dist/
 */
import { promises as fs } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { marked } from 'marked';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const DOCS_ROOT = path.resolve(__dirname, '..');
const DIST = path.join(__dirname, 'dist');
const ASSETS = path.join(__dirname, 'assets');

const nav = JSON.parse(await fs.readFile(path.join(__dirname, 'nav.json'), 'utf8'));

marked.setOptions({
  gfm: true,
  breaks: false,
});

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function mdToHtml(md, pageSlug) {
  let html = marked.parse(md);
  // Fix relative markdown links: foo.md -> foo.html, architecture/foo.md -> foo.html
  html = html.replace(/href="([^"]+\.md)(#[^"]*)?"/g, (_, href, hash = '') => {
    const base = path.basename(href, '.md');
    const slug = base === 'README' ? 'index' : base;
    return `href="./${slug}.html${hash || ''}"`;
  });
  html = html.replace(/href="\.\.\/([^"]+)"/g, (_, rest) => {
    if (rest.endsWith('.md')) {
      const base = path.basename(rest, '.md');
      return `href="https://github.com/opentmd/opentmd-cli/blob/master/${rest}"`;
    }
    return `href="https://github.com/opentmd/opentmd-cli/blob/master/${rest}"`;
  });
  return html;
}

function renderSidebar(activeSlug) {
  return nav.groups
    .map((g) => {
      const links = g.items
        .map((item) => {
          const active = item.slug === activeSlug ? ' active' : '';
          const aria = item.slug === activeSlug ? ' aria-current="page"' : '';
          return `<a class="dside-link${active}" href="./${item.slug}.html" data-slug="${item.slug}"${aria}>${escapeHtml(item.title)}</a>`;
        })
        .join('\n');
      return `<div class="dside-group"><div class="dside-group-t">${escapeHtml(g.title)}</div>${links}</div>`;
    })
    .join('\n');
}

function renderPage({ title, slug, body, description }) {
  const sidebar = renderSidebar(slug);
  const desc = escapeHtml(description || `${title} — OpenTMD 文档`);
  return `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>${escapeHtml(title)} · OpenTMD 文档</title>
<meta name="description" content="${desc}">
<link rel="canonical" href="https://www.opentmd.com/zh/${slug}.html">
<link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>⚡</text></svg>">
<link rel="stylesheet" href="../assets/docs.css">
<script>(function(){try{if(localStorage.getItem('opentmd_theme')==='light')document.documentElement.classList.add('light')}catch(e){}})();</script>
</head>
<body data-page="${slug}">
<header class="dhdr" id="dhdr">
  <a class="dhdr-logo" href="./index.html">
    <span class="dhdr-mark">⚡</span>
    <span>OpenTMD</span>
    <span class="dhdr-badge">DOCS</span>
  </a>
  <div class="dhdr-right">
    <button class="search-trigger" data-open-search aria-label="搜索文档">
      <span>搜索…</span><span class="kbd">⌘K</span>
    </button>
    <button class="icon-btn" id="themeBtn" aria-label="切换主题"></button>
    <a class="dhdr-link" href="${nav.repoUrl}" target="_blank" rel="noopener">GitHub →</a>
    <button class="icon-btn sb-toggle" id="sbToggle" aria-label="目录">☰</button>
  </div>
</header>
<div class="dlayout">
  <aside class="dside" id="dside">${sidebar}</aside>
  <main class="dmain prose-docs">${body}</main>
</div>
<footer class="dfoot">
  <span>© ${new Date().getFullYear()} OpenTMD · MIT</span>
  <a href="${nav.repoUrl}/issues" target="_blank" rel="noopener">报告问题</a>
</footer>
<div class="search-modal" id="searchModal" aria-hidden="true">
  <div class="search-modal-bg"></div>
  <div class="search-panel">
    <input id="searchInput" type="search" placeholder="搜索文档…" autocomplete="off">
    <div id="searchResults" class="search-results"></div>
  </div>
</div>
<script src="../assets/docs.js"></script>
</body>
</html>`;
}

async function copyDir(src, dest) {
  await fs.mkdir(dest, { recursive: true });
  const entries = await fs.readdir(src, { withFileTypes: true });
  for (const e of entries) {
    const s = path.join(src, e.name);
    const d = path.join(dest, e.name);
    if (e.isDirectory()) await copyDir(s, d);
    else await fs.copyFile(s, d);
  }
}

const searchIndex = [];
const zhOut = path.join(DIST, 'zh');
await fs.rm(DIST, { recursive: true, force: true });
await fs.mkdir(zhOut, { recursive: true });
await copyDir(ASSETS, path.join(DIST, 'assets'));
await fs.copyFile(path.join(__dirname, 'index.html'), path.join(DIST, 'index.html'));
await fs.copyFile(path.join(__dirname, 'CNAME'), path.join(DIST, 'CNAME'));

for (const group of nav.groups) {
  for (const item of group.items) {
    const srcPath = path.join(DOCS_ROOT, item.src);
    const md = await fs.readFile(srcPath, 'utf8');
    const body = mdToHtml(md, item.slug);
    const title = item.title;
    const html = renderPage({
      title,
      slug: item.slug,
      body,
      description: item.description || '',
    });
    await fs.writeFile(path.join(zhOut, `${item.slug}.html`), html);

    const plain = body.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim();
    searchIndex.push({
      slug: item.slug,
      title,
      group: group.title,
      url: `./${item.slug}.html`,
      body: plain.slice(0, 500),
    });
    console.log(`  ✓ zh/${item.slug}.html ← ${item.src}`);
  }
}

await fs.writeFile(
  path.join(DIST, 'search-index.zh.json'),
  JSON.stringify({ pages: searchIndex }, null, 0)
);

console.log(`\nBuilt ${searchIndex.length} pages → ${DIST}`);
