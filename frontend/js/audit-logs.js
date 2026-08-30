// audit-logs.js — every recorded action, most recent first.

const ACTOR_LABEL = { CUSTOMER: '👤 Customer', BANKER: '👨‍💼 Banker', SYSTEM: '⚙ System' };

async function loadAuditLogs() {
  try {
    const list = await api.get('/banker/audit-logs') || [];
    const body = qs('#dataBody');
    if (list.length === 0) {
      body.innerHTML = `<tr><td colspan="5" class="muted">No audit log entries yet.</td></tr>`;
      return;
    }
    body.innerHTML = list.map((a) => `
      <tr>
        <td><span class="badge badge-neutral">${escapeHtml(a.action)}</span></td>
        <td>${ACTOR_LABEL[a.actorType] || a.actorType}</td>
        <td class="muted">${escapeHtml(a.entity)}${a.entityId ? ' #' + a.entityId : ''}</td>
        <td>${escapeHtml(a.description)}</td>
        <td class="muted">${formatDate(a.createdAt)}</td>
      </tr>`).join('');
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  loadAuditLogs();
  const toolbar = qs('#searchInput');
  if (toolbar) toolbar.closest('.table-toolbar').classList.add('hidden');
});
