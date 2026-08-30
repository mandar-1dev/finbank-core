// customer-details.js (banker) — full picture of one customer: accounts
// (with freeze/unfreeze), cards, loans, and recent transactions.

async function toggleFreeze(accountId, status, customerId) {
  const action = status === 'ACTIVE' ? 'freeze' : 'unfreeze';
  const confirmed = await confirmModal({
    title: action === 'freeze' ? 'Freeze Account' : 'Unfreeze Account',
    message: `Are you sure you want to ${action} this account? This will be recorded in the audit log.`,
    confirmLabel: action === 'freeze' ? 'Freeze' : 'Unfreeze',
    danger: action === 'freeze',
  });
  if (!confirmed) return;
  try {
    await api.post(`/banker/accounts/${accountId}/${action}`);
    notify.success(action === 'freeze' ? 'Account frozen' : 'Account unfrozen');
    loadDetails(customerId);
  } catch (err) {
    notify.error(err.message);
  }
}

async function loadDetails(customerId) {
  try {
    const data = await api.get(`/banker/customers/${customerId}`);
    const c = data.customer;

    qs('#custName').textContent = c.fullName;
    qs('#custMeta').textContent = `${c.customerNumber} · ${c.email} · ${c.phone}`;

    qs('#acctGrid').innerHTML = (data.accounts || []).map((a) => `
      <div class="card">
        <p class="faint" style="text-transform:uppercase; font-size:0.75rem;">${escapeHtml(a.accountType)}</p>
        <p class="mono muted">${a.maskedNumber}</p>
        <h3 style="margin:10px 0;">${formatCurrency(a.availableBalance)}</h3>
        <span class="badge ${statusBadgeClass(a.status)}">${a.status}</span>
        <button class="btn btn-sm ${a.status === 'ACTIVE' ? 'btn-danger' : 'btn-secondary'} btn-block" style="margin-top:12px;" data-freeze="${a.id}" data-status="${a.status}">
          ${a.status === 'ACTIVE' ? 'Freeze Account' : 'Unfreeze Account'}
        </button>
      </div>`).join('') || `<p class="muted">No accounts.</p>`;

    qsa('[data-freeze]', qs('#acctGrid')).forEach((btn) => {
      btn.addEventListener('click', () => toggleFreeze(btn.dataset.freeze, btn.dataset.status, customerId));
    });

    qs('#cardGrid').innerHTML = (data.cards || []).map((card) => `
      <div class="card">
        <p class="mono muted">${card.maskedNumber}</p>
        <p style="margin-top:6px;">${escapeHtml(card.cardHolder)}</p>
        <span class="badge ${statusBadgeClass(card.status)}" style="margin-top:8px; display:inline-block;">${card.status}</span>
      </div>`).join('') || `<p class="muted">No cards.</p>`;

    qs('#loanList').innerHTML = (data.loans || []).map((l) => `
      <div class="card" style="margin-bottom:10px; display:flex; justify-content:space-between; align-items:center;">
        <div><strong>${formatCurrency(l.principalAmount)}</strong> <span class="muted">at ${l.interestRate}% · ${l.tenureMonths}mo</span></div>
        <span class="badge ${statusBadgeClass(l.status)}">${l.status}</span>
      </div>`).join('') || `<p class="muted">No loans.</p>`;

    const body = qs('#txnBody');
    const txns = data.transactions || [];
    body.innerHTML = txns.length ? txns.map((t) => `
      <tr>
        <td class="mono faint">${t.transactionRef}</td>
        <td><span class="badge badge-neutral">${t.type}</span></td>
        <td>${escapeHtml(t.description || '')}</td>
        <td class="muted">${formatDate(t.createdAt)}</td>
        <td style="text-align:right;">${formatCurrency(t.amount)}</td>
      </tr>`).join('') : `<tr><td colspan="5" class="muted">No transactions.</td></tr>`;
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const id = new URLSearchParams(window.location.search).get('id');
  if (!id) { window.location.href = 'customers.html'; return; }
  loadDetails(id);
});
