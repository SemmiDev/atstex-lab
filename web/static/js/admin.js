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
      if (target === 'feedback' && !fbLoaded) loadFeedbacks();
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
      usersTbody.innerHTML = '<tr><td colspan="9" style="text-align:center;padding:32px;color:var(--muted);">No users found</td></tr>';
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
        <td>${u.username ? `<a href="/u/${escHtml(u.username)}" target="_blank" style="color:var(--accent);font-weight:700;font-family:var(--font-mono);font-size:0.8rem;text-decoration:none;">/u/${escHtml(u.username)}</a>` : '<span style="color:var(--muted);">–</span>'}</td>
        <td><span class="role-badge ${u.role}">${u.role}</span></td>
        <td>${u.isBlocked ? '<span class="role-badge" style="background:var(--error, #ef4444);color:#fff;">Blocked</span>' : '<span class="role-badge" style="background:var(--accent2);color:#fff;">Active</span>'}</td>
        <td>${u.biodataCount}</td>
        <td>${fmtNum(u.aiTokensUsed)}</td>
        <td>${fmtDate(u.createdAt)}</td>
        <td style="white-space:nowrap;">
          ${u.role !== 'admin' ? (u.isBlocked
            ? `<button class="btn-reply" onclick="adminUnblockUser('${u.id}')"><i class="ph-bold ph-lock-key-open"></i> Unblock</button>`
            : `<button class="btn-reply" onclick="adminBlockUser('${u.id}')"><i class="ph-bold ph-prohibit"></i> Block</button>`
          ) : ''}
          ${u.role !== 'admin' ? `<button class="btn-reply" style="color:var(--accent2);border-color:var(--accent2);margin-left:4px;" onclick="adminMakeUserAdmin('${u.id}', '${escHtml(u.name)}')"><i class="ph-bold ph-shield-star"></i> Admin</button>` : ''}
          ${(u.role === 'admin' && u.id !== window.currentUserId) ? `<button class="btn-reply" style="color:var(--warning, #fbbf24);border-color:var(--warning, #fbbf24);margin-left:4px;" onclick="adminRevokeUserAdmin('${u.id}', '${escHtml(u.name)}')"><i class="ph-bold ph-shield-minus"></i> Revoke Admin</button>` : ''}
          ${u.role !== 'admin' ? `<button class="btn-reply" style="border-color:var(--error);color:var(--error);margin-left:4px;" onclick="adminDeleteUser('${u.id}', '${escHtml(u.name)}')"><i class="ph-bold ph-trash"></i></button>` : ''}
        </td>
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

  // ── Feedback Panel ──────────────────────────────────────────
  let fbLoaded = false;
  let fbPage = 1;
  let fbSearchTimeout = null;
  let currentReplyId = null;

  const fbSearch = document.getElementById('fb-search');
  const fbTbody = document.getElementById('fb-tbody');
  const fbCount = document.getElementById('fb-count');
  const fbPagination = document.getElementById('fb-pagination');
  const replyModal = document.getElementById('reply-modal');
  const closeReplyModal = document.getElementById('close-reply-modal');
  const replyText = document.getElementById('reply-text');
  const replyFbInfo = document.getElementById('reply-fb-info');
  const replyStatus = document.getElementById('reply-status');
  const btnSendReply = document.getElementById('btn-send-reply');

  if (fbSearch) {
    fbSearch.addEventListener('input', () => {
      clearTimeout(fbSearchTimeout);
      fbSearchTimeout = setTimeout(() => {
        fbPage = 1;
        loadFeedbacks();
      }, 300);
    });
  }

  if (closeReplyModal) {
    closeReplyModal.addEventListener('click', () => {
      replyModal.style.display = 'none';
    });
    replyModal.addEventListener('click', (e) => {
      if (e.target === replyModal) replyModal.style.display = 'none';
    });
  }

  if (btnSendReply) {
    btnSendReply.addEventListener('click', async () => {
      const reply = replyText.value.trim();
      if (!reply || !currentReplyId) return;

      btnSendReply.disabled = true;
      replyStatus.textContent = 'Sending…';
      replyStatus.style.color = 'var(--muted)';

      try {
        const res = await fetch(`/api/admin/feedbacks/${currentReplyId}/reply`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ reply }),
        });
        if (!res.ok) throw new Error('Failed');
        replyStatus.textContent = '✓ Reply sent!';
        replyStatus.style.color = 'var(--accent2)';
        setTimeout(() => {
          replyModal.style.display = 'none';
          replyText.value = '';
          replyStatus.textContent = '';
          currentReplyId = null;
          loadFeedbacks();
        }, 800);
      } catch (e) {
        replyStatus.textContent = '✗ Failed to send';
        replyStatus.style.color = 'var(--error, #ff3333)';
      } finally {
        btnSendReply.disabled = false;
      }
    });
  }

  async function loadFeedbacks() {
    fbLoaded = true;
    const search = fbSearch ? fbSearch.value.trim() : '';
    const params = new URLSearchParams({ page: fbPage, per_page: 20 });
    if (search) params.set('search', search);

    try {
      const res = await fetch(`/api/admin/feedbacks?${params}`);
      if (!res.ok) throw new Error('Failed');
      const data = await res.json();

      fbCount.textContent = `${data.total} feedback${data.total !== 1 ? 's' : ''}`;
      renderFbTable(data.feedbacks || []);
      renderFbPagination(data.total, data.page, data.perPage);
    } catch (e) {
      console.error('Failed to load feedbacks:', e);
      fbTbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--muted);">Failed to load feedback</td></tr>';
    }
  }

  function renderFbTable(feedbacks) {
    if (!feedbacks.length) {
      fbTbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--muted);">No feedback found</td></tr>';
      return;
    }

    fbTbody.innerHTML = feedbacks.map(fb => {
      const hasReply = fb.adminReply && fb.adminReply.length > 0;
      const statusBadge = hasReply
        ? '<span class="role-badge admin">Answered</span>'
        : '<span class="role-badge user">Pending</span>';
      const actionBtn = hasReply
        ? `<button class="btn-reply" onclick="openReplyModal('${fb.id}', this)" data-subject="${escHtml(fb.subject)}" data-reply="${escHtml(fb.adminReply || '')}">Edit Reply</button>`
        : `<button class="btn-reply primary" onclick="openReplyModal('${fb.id}', this)" data-subject="${escHtml(fb.subject)}" data-reply="">Reply</button>`;
      const deleteBtn = `<button class="btn-reply" style="border-color:var(--error);color:var(--error);margin-left:4px;" onclick="adminDeleteFeedback('${fb.id}')"><i class="ph-bold ph-trash"></i></button>`;
      const msgPreview = (fb.message || '').length > 80 ? fb.message.substring(0, 80) + '…' : fb.message;

      return `
        <tr>
          <td>
            <div class="user-cell">
              <img src="${escHtml(fb.userPicture || '')}" alt="" onerror="this.style.display='none'">
              <div>
                <div style="font-weight:700;">${escHtml(fb.userName)}</div>
                <div style="font-size:0.7rem;color:var(--muted);">${escHtml(fb.userEmail)}</div>
              </div>
            </div>
          </td>
          <td style="font-weight:700;">${escHtml(fb.subject)}</td>
          <td style="max-width:250px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;" title="${escHtml(fb.message)}">${escHtml(msgPreview)}</td>
          <td>${statusBadge}</td>
          <td>${fmtDate(fb.createdAt)}</td>
          <td style="white-space:nowrap;">${actionBtn} ${deleteBtn}</td>
        </tr>`;
    }).join('');
  }

  function renderFbPagination(total, page, perPage) {
    const totalPages = Math.ceil(total / perPage);
    if (totalPages <= 1) {
      fbPagination.innerHTML = '';
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
    fbPagination.innerHTML = html;

    fbPagination.querySelectorAll('button[data-page]').forEach(btn => {
      btn.addEventListener('click', () => {
        const p = parseInt(btn.dataset.page);
        if (p >= 1 && p <= totalPages) {
          fbPage = p;
          loadFeedbacks();
        }
      });
    });
  }

  // Make openReplyModal available globally
  window.openReplyModal = function(id, btn) {
    currentReplyId = id;
    const subject = btn.dataset.subject || '';
    const existingReply = btn.dataset.reply || '';
    replyFbInfo.innerHTML = `<div style="font-weight:800;text-transform:uppercase;font-size:0.85rem;margin-bottom:4px;">Re: ${subject}</div>`;
    replyText.value = existingReply;
    replyStatus.textContent = '';
    replyModal.style.display = 'flex';
    replyText.focus();
  };

  // ── Admin Action Helpers ────────────────────────────────────

  window.adminBlockUser = async function(id) {
    if (!confirm('Are you sure you want to block this user? They will be logged out and unable to access the platform.')) return;
    try {
      const res = await fetch(`/api/admin/users/${id}/block`, { method: 'POST' });
      if (!res.ok) throw new Error('Failed');
      loadUsers();
    } catch (e) {
      alert('Failed to block user.');
    }
  };

  window.adminUnblockUser = async function(id) {
    if (!confirm('Unblock this user? They will be able to log in again.')) return;
    try {
      const res = await fetch(`/api/admin/users/${id}/unblock`, { method: 'POST' });
      if (!res.ok) throw new Error('Failed');
      loadUsers();
    } catch (e) {
      alert('Failed to unblock user.');
    }
  };

  window.adminDeleteUser = async function(id, name) {
    if (!confirm(`⚠️ PERMANENTLY DELETE user "${name}" and ALL their data? This cannot be undone.`)) return;
    if (!confirm('Are you ABSOLUTELY sure? This action is irreversible.')) return;
    try {
      const res = await fetch(`/api/admin/users/${id}`, { method: 'DELETE' });
      if (!res.ok) throw new Error('Failed');
      loadUsers();
    } catch (e) {
      alert('Failed to delete user.');
    }
  };

  window.adminMakeUserAdmin = async function(id, name) {
    if (!confirm(`Are you sure you want to promote "${name}" to Admin? They will have full access to this dashboard.`)) return;
    try {
      const res = await fetch(`/api/admin/users/${id}/make-admin`, { method: 'POST' });
      if (!res.ok) throw new Error('Failed');
      loadUsers();
    } catch (e) {
      alert('Failed to promote user to Admin.');
    }
  };

  window.adminRevokeUserAdmin = async function(id, name) {
    if (!confirm(`⚠️ Are you sure you want to REVOKE Admin access for "${name}"? They will lose access to this dashboard.`)) return;
    try {
      const res = await fetch(`/api/admin/users/${id}/revoke-admin`, { method: 'POST' });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || 'Failed');
      loadUsers();
    } catch (e) {
      alert('Failed to revoke Admin access: ' + e.message);
    }
  };

  window.adminDeleteFeedback = async function(id) {
    if (!confirm('Delete this feedback entry? This cannot be undone.')) return;
    try {
      const res = await fetch(`/api/admin/feedbacks/${id}`, { method: 'DELETE' });
      if (!res.ok) throw new Error('Failed');
      loadFeedbacks();
    } catch (e) {
      alert('Failed to delete feedback.');
    }
  };
})();
