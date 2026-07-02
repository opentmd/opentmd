(function () {
  'use strict';

  const html = document.documentElement;

  function applyTheme(light) {
    html.classList.toggle('light', light);
    const btn = document.getElementById('themeBtn');
    if (btn) {
      btn.textContent = light ? '☀' : '☾';
      btn.setAttribute('aria-label', light ? '深色模式' : '浅色模式');
    }
    try {
      localStorage.setItem('opentmd_theme', light ? 'light' : 'dark');
    } catch (_) {}
  }

  function readTheme() {
    try {
      const s = localStorage.getItem('opentmd_theme');
      if (s === 'light' || s === 'dark') return s === 'light';
    } catch (_) {}
    return window.matchMedia('(prefers-color-scheme: light)').matches;
  }

  applyTheme(readTheme());

  document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('themeBtn')?.addEventListener('click', () => {
      applyTheme(!html.classList.contains('light'));
    });

    const sb = document.getElementById('dside');
    const toggle = document.getElementById('sbToggle');
    if (sb && toggle) {
      toggle.addEventListener('click', () => {
        const open = sb.classList.toggle('open');
        toggle.textContent = open ? '✕' : '☰';
      });
    }

    initSearch();
  });

  let searchData = null;

  async function loadSearch() {
    if (searchData) return searchData;
    try {
      const r = await fetch('../search-index.zh.json', { cache: 'no-cache' });
      if (!r.ok) throw new Error(r.status);
      searchData = await r.json();
    } catch (e) {
      console.warn('[docs] search index:', e);
      searchData = { pages: [] };
    }
    return searchData;
  }

  function initSearch() {
    const modal = document.getElementById('searchModal');
    const input = document.getElementById('searchInput');
    const results = document.getElementById('searchResults');
    if (!modal || !input || !results) return;

    let sel = 0;
    let items = [];

    function close() {
      modal.classList.remove('open');
      modal.setAttribute('aria-hidden', 'true');
    }

    function open() {
      modal.classList.add('open');
      modal.setAttribute('aria-hidden', 'false');
      loadSearch().then(() => {
        input.value = '';
        render('');
        input.focus();
      });
    }

    function render(q) {
      const data = searchData?.pages || [];
      const query = q.trim().toLowerCase();
      items = query
        ? data.filter((p) => {
            const hay = `${p.title} ${p.group} ${p.body}`.toLowerCase();
            return hay.includes(query);
          })
        : data.slice(0, 12);
      sel = 0;
      if (!items.length) {
        results.innerHTML = '<div class="search-item">没有匹配结果</div>';
        return;
      }
      results.innerHTML = items
        .map(
          (p, i) =>
            `<a class="search-item${i === sel ? ' active' : ''}" href="${p.url}"><strong>${esc(p.title)}</strong><small>${esc(p.group)}</small></a>`
        )
        .join('');
    }

    function esc(s) {
      return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
    }

    document.querySelectorAll('[data-open-search]').forEach((el) => el.addEventListener('click', open));
    modal.querySelector('.search-modal-bg')?.addEventListener('click', close);

    input.addEventListener('input', () => render(input.value));
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') close();
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        sel = Math.min(sel + 1, items.length - 1);
        render(input.value);
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        sel = Math.max(sel - 1, 0);
        render(input.value);
      }
      if (e.key === 'Enter' && items[sel]) {
        location.href = items[sel].url;
      }
    });

    document.addEventListener('keydown', (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        open();
      }
    });
  }
})();
