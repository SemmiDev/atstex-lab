// ══════════════════════════════════════════════════════════════
// ATSTEXLAB — Admin Dashboard JS
// ══════════════════════════════════════════════════════════════

(function () {
  'use strict';

  // ── Panel switching ─────────────────────────────────────────
  const sidebarLinks = document.querySelectorAll('.sidebar-link');
  const panels = document.querySelectorAll('.admin-panel');

  sidebarLinks.forEach(btn => {
    btn.addEventListener('click', () => {
      const target = btn.dataset.panel;
      sidebarLinks.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      panels.forEach(p => {
        p.classList.toggle('active', p.id === `panel-${target}`);
      });
      // Lazy load
      if (target === 'users' && !usersLoaded) loadUsers();
    });
  });

  // ── Dashboard Stats ─────────────────────────────────────────
  async function loadStats() {
    try {
      const res = await fetch('/api/admin/stats');
      if (!res.ok) throw new Error('Failed');
      const s = await res.json();
      document.getElementById('stat-users').textContent = fmtNum(s.totalUsers);
      document.getElementById('stat-admins').textContent = fmtNum(s.totalAdmins);
      document.getElementById('stat-ai-tokens').textContent = fmtNum(s.totalAITokens);
      document.getElementById('stat-biodata').textContent = fmtNum(s.totalBiodata);
      document.getElementById('stat-sessions').textContent = fmtNum(s.totalSessions);
    } catch (e) {
      console.error('Failed to load stats:', e);
    }
  }
  loadStats();

  // ── Users Panel ─────────────────────────────────────────────
  let usersLoaded = false;
  let currentPage = 1;
  let currentSort = 'created_at';
  let currentOrder = 'desc';
  let searchTimeout = null;

  const userSearch = document.getElementById('user-search');
  const usersTbody = document.getElementById('users-tbody');
  const usersCount = document.getElementById('users-count');
  const pagination = document.getElementById('pagination');

  // Search with debounce
  userSearch.addEventListener('input', () => {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      currentPage = 1;
      loadUsers();
    }, 300);
  });

  // Column sort
  document.querySelectorAll('.sortable').forEach(th => {
    th.addEventListener('click', () => {
      const col = th.dataset.sort;
      if (currentSort === col) {
        currentOrder = currentOrder === 'asc' ? 'desc' : 'asc';
      } else {
        currentSort = col;
        currentOrder = 'asc';
      }
      currentPage = 1;
      loadUsers();
    });
  });

  async function loadUsers() {
    usersLoaded = true;
    const search = userSearch.value.trim();
    const params = new URLSearchParams({
      page: currentPage,
      per_page: 20,
      sort: currentSort,
      order: currentOrder,
    });
    if (search) params.set('search', search);

    try {
      const res = await fetch(`/api/admin/users?${params}`);
      if (!res.ok) throw new Error('Failed');
      const data = await res.json();

      usersCount.textContent = `${data.total} user${data.total !== 1 ? 's' : ''}`;
      renderUsersTable(data.users || []);
      renderPagination(data.total, data.page, data.perPage);
    } catch (e) {
      console.error('Failed to load users:', e);
      usersTbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--muted);">Failed to load users</td></tr>';
    }
  }

  function renderUsersTable(users) {
    if (!users.length) {
      usersTbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--muted);">No users found</td></tr>';
      return;
    }

    usersTbody.innerHTML = users.map(u => `
      <tr>
        <td>
          <div class="user-cell">
            <img src="${escHtml(u.picture || '')}" alt="" onerror="this.style.display='none'">
            <span>${escHtml(u.name)}</span>
          </div>
        </td>
        <td>${escHtml(u.email)}</td>
        <td><span class="role-badge ${u.role}">${u.role}</span></td>
        <td>${u.biodataCount}</td>
        <td>${fmtNum(u.aiTokensUsed)}</td>
        <td>${fmtDate(u.createdAt)}</td>
      </tr>
    `).join('');
  }

  function renderPagination(total, page, perPage) {
    const totalPages = Math.ceil(total / perPage);
    if (totalPages <= 1) {
      pagination.innerHTML = '';
      return;
    }

    let html = '';
    html += `<button ${page <= 1 ? 'disabled' : ''} data-page="${page - 1}">← Prev</button>`;

    const maxButtons = 7;
    let start = Math.max(1, page - Math.floor(maxButtons / 2));
    let end = Math.min(totalPages, start + maxButtons - 1);
    if (end - start < maxButtons - 1) start = Math.max(1, end - maxButtons + 1);

    for (let i = start; i <= end; i++) {
      html += `<button class="${i === page ? 'active' : ''}" data-page="${i}">${i}</button>`;
    }

    html += `<button ${page >= totalPages ? 'disabled' : ''} data-page="${page + 1}">Next →</button>`;
    pagination.innerHTML = html;

    pagination.querySelectorAll('button[data-page]').forEach(btn => {
      btn.addEventListener('click', () => {
        const p = parseInt(btn.dataset.page);
        if (p >= 1 && p <= totalPages) {
          currentPage = p;
          loadUsers();
        }
      });
    });
  }

  // ── Helpers ─────────────────────────────────────────────────
  function fmtNum(n) {
    if (n == null) return '0';
    return Number(n).toLocaleString();
  }

  function fmtDate(iso) {
    if (!iso) return '–';
    const d = new Date(iso);
    return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  function escHtml(str) {
    const el = document.createElement('span');
    el.textContent = str || '';
    return el.innerHTML;
  }
})();
