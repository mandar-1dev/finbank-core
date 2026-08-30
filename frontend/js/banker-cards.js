// banker-cards.js — read-only overview of every card in the system.

async function loadCards() {
  try {
    const list = await api.get('/banker/cards') || [];
    const body = qs('#dataBody');
    if (list.length === 0) {
      body.innerHTML = `<tr><td colspan="4" class="muted">No cards issued.</td></tr>`;
      return;
    }
    body.innerHTML = list.map((c) => `
      <tr>
        <td class="mono">${c.maskedNumber}</td>
        <td>${escapeHtml(c.cardHolder)}</td>
        <td><span class="badge ${statusBadgeClass(c.status)}">${c.status}</span></td>
        <td>${formatCurrency(c.spendingLimit)}</td>
      </tr>`).join('');
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  loadCards();
  const toolbar = qs('#searchInput');
  if (toolbar) toolbar.closest('.table-toolbar').classList.add('hidden');
});
