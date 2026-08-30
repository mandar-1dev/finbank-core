// banker-transactions.js — full transaction feed across the system.

let allTxns = [];

function render() {
  const q = qs('#searchInput').value.trim().toLowerCase();
  const filtered = allTxns.filter((t) =>
    !q || (t.description || '').toLowerCase().includes(q) || t.transactionRef.toLowerCase().includes(q));
  const body = qs('#dataBody');
  if (filtered.length === 0) {
    body.innerHTML = `<tr><td colspan="6" class="muted">No transactions match your search.</td></tr>`;
    return;
  }
  body.innerHTML = filtered.map((t) => `
    <tr>
      <td class="mono faint">${t.transactionRef}</td>
      <td><span class="badge badge-neutral">${t.type}</span></td>
      <td>${escapeHtml(t.description || '')}</td>
      <td class="muted">${formatDate(t.createdAt)}</td>
      <td><span class="badge ${statusBadgeClass(t.status)}">${t.status}</span></td>
      <td style="text-align:right;">${formatCurrency(t.amount)}</td>
    </tr>`).join('');
}

async function loadTransactions() {
  try {
    allTxns = await api.get('/banker/transactions') || [];
    render();
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  loadTransactions();
  qs('#searchInput').addEventListener('input', debounce(render, 150));
});
