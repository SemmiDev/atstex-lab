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
      if (target === 'subscriptions' && !subsLoaded) loadSubs();
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
      const el = (id, v) => { const e = document.getElementById(id); if (e) e.textContent = fmtNum(v); };
      el('stat-cv-reviews', s.totalCVReviews);
      el('stat-cover-letters', s.totalCoverLetters);
      el('stat-ats-sims', s.totalATSSimulations);
      el('stat-job-apps', s.totalJobApps);
    } catch (e) {
      console.error('Failed to load stats:', e);
    }
  }
  loadStats();

  // ── Analytics Charts ───────────────────────────────────────
  let chartsLoaded = false;
  async function loadAnalytics() {
    if (chartsLoaded) return;
    chartsLoaded = true;
    try {
      const res = await fetch('/api/admin/analytics');
      if (!res.ok) throw new Error('Failed');
      const data = await res.json();

      const fmtLabel = (iso) => {
        const d = new Date(iso);
        return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short' });
      };

      const chartOpts = {
        responsive: true,
        maintainAspectRatio: true,
        aspectRatio: 2.2,
        devicePixelRatio: 2,
        plugins: { legend: { labels: { font: { family: "'DM Sans', sans-serif", weight: 700, size: 11 }, usePointStyle: true, padding: 16 } } },
        scales: {
          x: { grid: { display: false }, ticks: { font: { family: "'Fira Code', monospace", size: 9 }, maxRotation: 45, minRotation: 30 } },
          y: { beginAtZero: true, ticks: { font: { family: "'Fira Code', monospace", size: 10 }, stepSize: 1 }, grid: { color: 'rgba(0,0,0,0.06)' } }
        }
      };

      // Chart 1: User Registrations (line)
      const regCtx = document.getElementById('chart-registrations');
      if (regCtx) {
        new Chart(regCtx, {
          type: 'line',
          data: {
            labels: data.userRegistrations.map(d => fmtLabel(d.date)),
            datasets: [{
              label: 'Pendaftaran',
              data: data.userRegistrations.map(d => d.count),
              borderColor: '#3b82f6',
              backgroundColor: 'rgba(59,130,246,0.1)',
              borderWidth: 3,
              fill: true,
              tension: 0.35,
              pointRadius: 2,
              pointHoverRadius: 6,
              pointBackgroundColor: '#3b82f6'
            }]
          },
          options: chartOpts
        });
      }

      // Chart 2: AI Feature Usage (stacked bar)
      const aiCtx = document.getElementById('chart-ai-usage');
      if (aiCtx) {
        const labels = data.cvReviews.map(d => fmtLabel(d.date));
        new Chart(aiCtx, {
          type: 'bar',
          data: {
            labels,
            datasets: [
              { label: 'Review CV', data: data.cvReviews.map(d => d.count), backgroundColor: '#5eeb8f', borderWidth: 0, borderRadius: 2 },
              { label: 'Surat Lamaran', data: data.coverLetters.map(d => d.count), backgroundColor: '#b066ff', borderWidth: 0, borderRadius: 2 },
              { label: 'Simulasi ATS', data: data.atsSimulations.map(d => d.count), backgroundColor: '#ff8c3a', borderWidth: 0, borderRadius: 2 }
            ]
          },
          options: {
            ...chartOpts,
            scales: {
              ...chartOpts.scales,
              x: { ...chartOpts.scales.x, stacked: true },
              y: { ...chartOpts.scales.y, stacked: true }
            }
          }
        });
      }
    } catch (e) {
      chartsLoaded = false;
      console.error('Failed to load analytics:', e);
    }
  }
  loadAnalytics();

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
          ${u.role !== 'admin' ? `<button class="btn-reply" style="color:var(--accent);border-color:var(--accent);" onclick="openAssignPlanModal('${u.id}', '${escHtml(u.name)}')"><i class="ph-bold ph-crown"></i> Assign Plan</button>` : ''}
          ${u.role !== 'admin' ? (u.isBlocked
            ? `<button class="btn-reply" style="margin-left:4px;" onclick="adminUnblockUser('${u.id}')"><i class="ph-bold ph-lock-key-open"></i> Unblock</button>`
            : `<button class="btn-reply" style="margin-left:4px;" onclick="adminBlockUser('${u.id}')"><i class="ph-bold ph-prohibit"></i> Block</button>`
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
      if (!res.ok) throw new Error((data.detail || data.error) || 'Failed');
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

  // ── Subscriptions Panel ─────────────────────────────────────
  let subsLoaded = false;
  let allPlans = [];
  const subsCount = document.getElementById('subs-count');
  const subsTbody = document.getElementById('subs-tbody');

  const planModal = document.getElementById('plan-modal');
  const planForm = document.getElementById('plan-form');
  const btnAddPlan = document.getElementById('btn-add-plan');
  const btnClosePlan = document.getElementById('close-plan-modal');
  const btnSavePlan = document.getElementById('btn-save-plan');

  if (btnAddPlan) btnAddPlan.addEventListener('click', () => openPlanModal());
  if (btnClosePlan) {
    btnClosePlan.addEventListener('click', () => { planModal.style.display = 'none'; });
    planModal.addEventListener('click', (e) => { if(e.target === planModal) planModal.style.display = 'none'; });
  }

  async function loadSubs() {
    subsLoaded = true;
    try {
      const res = await fetch('/api/admin/subscription-plans');
      if (!res.ok) throw new Error('Failed');
      const data = await res.json();
      allPlans = data || [];
      subsCount.textContent = `${allPlans.length} paket`;
      renderSubsTable(allPlans);
    } catch (e) {
      console.error('Failed to load subs:', e);
      subsTbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--muted);">Gagal memuat paket</td></tr>';
    }
  }

  function renderSubsTable(plans) {
    if (!plans.length) {
      subsTbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--muted);">Belum ada paket langganan</td></tr>';
      return;
    }
    subsTbody.innerHTML = plans.map(p => {
      const c = (v) => v === -1 ? '∞' : v;
      return `
      <tr>
        <td style="font-weight:700;">${escHtml(p.name)}</td>
        <td>Rp ${fmtNum(p.priceIdr)}</td>
        <td>${p.durationMonths} bln</td>
        <td>CV: ${c(p.maxCvProfiles)} | Rv: ${c(p.maxCvReviews)} | ATS: ${c(p.maxAtsSimulations)} | Cov: ${c(p.maxCoverLetters)}</td>
        <td style="text-align:center;"><span class="role-badge" style="background:var(--accent);color:#fff;">${p.activeUsersCount || 0}</span></td>
        <td>${p.isActive ? '<span class="role-badge" style="background:var(--accent2);color:#fff;">Aktif</span>' : '<span class="role-badge" style="background:var(--muted);color:#fff;">Nonaktif</span>'}</td>
        <td style="white-space:nowrap;">
          <button class="btn-reply" onclick="openPlanModal('${p.id}')"><i class="ph-bold ph-pencil"></i> Edit</button>
          <button class="btn-reply" style="margin-left:4px; ${p.isActive ? 'color:var(--warning);border-color:var(--warning);' : 'color:var(--accent2);border-color:var(--accent2);'}" onclick="togglePlan('${p.id}', ${!p.isActive})"><i class="ph-bold ph-power"></i> ${p.isActive ? 'Matikan' : 'Aktifkan'}</button>
          <button class="btn-reply" style="border-color:var(--error);color:var(--error);margin-left:4px;" onclick="adminDeletePlan('${p.id}', '${escHtml(p.name)}')"><i class="ph-bold ph-trash"></i> Hapus</button>
        </td>
      </tr>`;
    }).join('');
  }

  window.openPlanModal = function(id = null) {
    document.getElementById('plan-id').value = '';
    planForm.reset();
    document.getElementById('plan-modal-title').innerHTML = '<i class="ph-bold ph-crown"></i> Tambah Paket';

    if (id) {
      const plan = allPlans.find(p => p.id === id);
      if (plan) {
        document.getElementById('plan-id').value = plan.id;
        document.getElementById('plan-name').value = plan.name;
        document.getElementById('plan-price').value = plan.priceIdr;
        document.getElementById('plan-duration').value = plan.durationMonths;
        document.getElementById('plan-max-cv').value = plan.maxCvProfiles;
        document.getElementById('plan-max-review').value = plan.maxCvReviews;
        document.getElementById('plan-max-ats').value = plan.maxAtsSimulations;
        document.getElementById('plan-max-cover').value = plan.maxCoverLetters;
        document.getElementById('plan-active').checked = plan.isActive;
        document.getElementById('plan-modal-title').innerHTML = '<i class="ph-bold ph-crown"></i> Edit Paket';
      }
    }
    planModal.style.display = 'flex';
  };

  if (btnSavePlan) {
    btnSavePlan.addEventListener('click', async (e) => {
      e.preventDefault();
      if (!planForm.checkValidity()) {
        planForm.reportValidity();
        return;
      }

      const id = document.getElementById('plan-id').value;
      const payload = {
        name: document.getElementById('plan-name').value,
        priceIdr: parseInt(document.getElementById('plan-price').value),
        durationMonths: parseInt(document.getElementById('plan-duration').value),
        maxCvProfiles: parseInt(document.getElementById('plan-max-cv').value),
        maxCvReviews: parseInt(document.getElementById('plan-max-review').value),
        maxAtsSimulations: parseInt(document.getElementById('plan-max-ats').value),
        maxCoverLetters: parseInt(document.getElementById('plan-max-cover').value),
        isActive: document.getElementById('plan-active').checked
      };

      const method = id ? 'PUT' : 'POST';
      const url = id ? `/api/admin/subscription-plans/${id}` : '/api/admin/subscription-plans';

      btnSavePlan.disabled = true;
      try {
        const res = await fetch(url, {
          method,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        if (!res.ok) throw new Error('Failed');
        planModal.style.display = 'none';
        loadSubs();
      } catch (err) {
        alert('Gagal menyimpan paket!');
      } finally {
        btnSavePlan.disabled = false;
      }
    });
  }

  window.togglePlan = async function(id, isActive) {
    if (!confirm(`Apa Anda yakin ingin ${isActive ? 'mengaktifkan' : 'menonaktifkan'} paket ini?`)) return;
    try {
      const res = await fetch(`/api/admin/subscription-plans/${id}/toggle`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ isActive })
      });
      if (!res.ok) throw new Error('Failed');
      loadSubs();
    } catch (err) {
      alert('Gagal merubah status paket!');
    }
  };

  window.adminDeletePlan = async function(id, name) {
    if (!confirm(`⚠️ PERINGATAN: Apakah Anda yakin ingin MENGHAPUS paket "${name}"?
Jika ada pengguna yang sedang berlangganan paket ini, ini mungkin akan menyebabkan data langganan mereka bermasalah (akan dihapus karena CASCADE).
Tindakan ini tidak dapat dibatalkan.`)) return;
    try {
      const res = await fetch(`/api/admin/subscription-plans/${id}`, {
        method: 'DELETE'
      });
      if (!res.ok) throw new Error('Failed');
      loadSubs();
    } catch (e) {
      alert('Gagal menghapus paket langganan.');
    }
  };

  // ── Assign Plan Modal ───────────────────────────────────────
  const assignPlanModal = document.getElementById('assign-plan-modal');
  const btnCloseAssignPlanModal = document.getElementById('close-assign-plan-modal');
  const btnSubmitAssignPlan = document.getElementById('btn-submit-assign-plan');
  const assignPlanSelect = document.getElementById('assign-plan-select');

  if (btnCloseAssignPlanModal) {
    btnCloseAssignPlanModal.addEventListener('click', () => { assignPlanModal.style.display = 'none'; });
    if (assignPlanModal) {
      assignPlanModal.addEventListener('click', (e) => {
        if (e.target === assignPlanModal) assignPlanModal.style.display = 'none';
      });
    }
  }

  window.openAssignPlanModal = async function(userId, userName) {
    const userIdInput = document.getElementById('assign-plan-user-id');
    const userInfoDiv = document.getElementById('assign-plan-user-info');
    if (!userIdInput || !userInfoDiv || !assignPlanModal || !assignPlanSelect) return;

    userIdInput.value = userId;
    userInfoDiv.innerText = userName;
    assignPlanModal.style.display = 'flex';

    // Fetch and populate plans if empty
    assignPlanSelect.innerHTML = '<option value="">Memuat paket...</option>';
    try {
      if (allPlans.length === 0) {
        const res = await fetch('/api/admin/subscription-plans');
        if (res.ok) {
          allPlans = await res.json() || [];
        }
      }

      assignPlanSelect.innerHTML = '<option value="">-- Pilih Paket --</option>' +
        allPlans.filter(p => p.isActive).map(p => `<option value="${p.id}" data-months="${p.durationMonths}">${escHtml(p.name)} - Rp ${fmtNum(p.priceIdr)} (${p.durationMonths} bln)</option>`).join('');
    } catch (e) {
      assignPlanSelect.innerHTML = '<option value="">Gagal memuat paket</option>';
    }
  };

  if (btnSubmitAssignPlan) {
    btnSubmitAssignPlan.addEventListener('click', async (e) => {
      e.preventDefault();
      const userIdInput = document.getElementById('assign-plan-user-id');
      if (!userIdInput) return;
      const userId = userIdInput.value;
      const planId = assignPlanSelect.value;

      if (!planId) {
        alert("Pilih paket terlebih dahulu.");
        return;
      }

      const option = assignPlanSelect.options[assignPlanSelect.selectedIndex];
      const months = option.getAttribute('data-months') || "1";

      btnSubmitAssignPlan.disabled = true;
      try {
        const res = await fetch(`/api/admin/users/${userId}/subscribe`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ planId, months })
        });
        if (!res.ok) {
          const errData = await res.json().catch(() => ({}));
          throw new Error((errData.detail || errData.error) || "Gagal menetapkan paket.");
        }
        assignPlanModal.style.display = 'none';
        alert('Berhasil menetapkan paket.');
        loadUsers(); // Refresh the user list
      } catch (err) {
        alert(err.message || 'Gagal menetapkan paket!');
      } finally {
        btnSubmitAssignPlan.disabled = false;
      }
    });
  }

})();
