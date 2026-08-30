// suspicious-transactions.js — transactions flagged MEDIUM/HIGH by the
// simple rule-based simulation in the backend. Not real fraud detection.

async function loadSuspicious() {
  try {
    const list = await api.get('/banker/suspicious-transactions') || [];
    const body = qs('#dataBody');
    if (list.length === 0) {
      body.innerHTML = `<tr><td colspan="6" class="muted">No flagged transactions.</td></tr>`;
      return;
    }
    body.innerHTML = list.map((t) => `
      <tr>
        <td class="mono faint">${t.transactionRef}</td>
        <td><span class="badge badge-neutral">${t.type}</span></td>
        <td class="risk-${t.riskLevel}">${t.riskLevel}</td>
        <td>${escapeHtml(t.description || '')}</td>
        <td class="muted">${formatDate(t.createdAt)}</td>
        <td style="text-align:right;">${formatCurrency(t.amount)}</td>
      </tr>`).join('');
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  loadSuspicious();
  qs('#searchInput').closest('.table-toolbar').classList.add('hidden');
});
