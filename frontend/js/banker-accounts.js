// banker-accounts.js — all accounts in the system, with freeze/unfreeze.

let allAccounts = [];

function render() {
  const q = qs('#searchInput').value.trim().toLowerCase();
  const filtered = allAccounts.filter((a) => !q || a.accountNumber.includes(q) || a.maskedNumber.toLowerCase().includes(q));
  const body = qs('#dataBody');
  if (filtered.length === 0) {
    body.innerHTML = `<tr><td colspan="6" class="muted">No accounts match your search.</td></tr>`;
    return;
  }
  body.innerHTML = filtered.map((a) => `
    <tr>
      <td class="mono">${a.maskedNumber}</td>
      <td>${a.customerId}</td>
      <td>${escapeHtml(a.accountType)}</td>
      <td>${formatCurrency(a.balance)}</td>
      <td><span class="badge ${statusBadgeClass(a.status)}">${a.status}</span></td>
      <td><button class="btn btn-sm ${a.status === 'ACTIVE' ? 'btn-danger' : 'btn-secondary'}" data-freeze="${a.id}" data-status="${a.status}">
        ${a.status === 'ACTIVE' ? 'Freeze' : a.status === 'FROZEN' ? 'Unfreeze' : ''}
      </button></td>
    </tr>`).join('');

  qsa('[data-freeze]', body).forEach((btn) => {
    if (!btn.textContent.trim()) return;
    btn.addEventListener('click', async () => {
      const action = btn.dataset.status === 'ACTIVE' ? 'freeze' : 'unfreeze';
      const confirmed = await confirmModal({
        title: action === 'freeze' ? 'Freeze Account' : 'Unfreeze Account',
        message: `Are you sure you want to ${action} account ${btn.dataset.freeze}?`,
        confirmLabel: action === 'freeze' ? 'Freeze' : 'Unfreeze',
        danger: action === 'freeze',
      });
      if (!confirmed) return;
      try {
        await api.post(`/banker/accounts/${btn.dataset.freeze}/${action}`);
        notify.success(action === 'freeze' ? 'Account frozen' : 'Account unfrozen');
        loadAccounts();
      } catch (err) {
        notify.error(err.message);
      }
    });
  });
}

async function loadAccounts() {
  try {
    allAccounts = await api.get('/banker/accounts') || [];
    render();
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  loadAccounts();
  qs('#searchInput').addEventListener('input', debounce(render, 150));
});
