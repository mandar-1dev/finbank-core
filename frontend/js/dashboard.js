// dashboard.js — loads the selected customer's summary: total balance,
// account cards, and the 5 most recent transactions.

async function loadDashboard() {
  const customerId = requireSelectedCustomer();
  if (!customerId) return;

  try {
    const [customer, accounts, transactions] = await Promise.all([
      api.get(`/customers/${customerId}`),
      api.get(`/customers/${customerId}/accounts`),
      api.get(`/customers/${customerId}/transactions`),
    ]);

    qs('#greeting').textContent = `Hello, ${customer.fullName.split(' ')[0]} 👋`;

    const total = (accounts || []).reduce((sum, a) => sum + a.availableBalance, 0);
    qs('#totalBalance').textContent = formatCurrency(total);

    const grid = qs('#accountsGrid');
    grid.innerHTML = (accounts || []).map((a) => `
      <div class="card">
        <p class="faint" style="text-transform:uppercase; font-size:0.75rem; margin-bottom:4px;">${escapeHtml(a.accountType)} Account</p>
        <p class="mono muted" style="margin-bottom:10px;">${a.maskedNumber}</p>
        <h2 style="margin-bottom:10px;">${formatCurrency(a.availableBalance)}</h2>
        <span class="badge ${statusBadgeClass(a.status)}">${a.status}</span>
      </div>`).join('') || `<div class="empty-state"><p>No accounts found.</p></div>`;

    const recent = (transactions || []).slice(0, 5);
    const body = qs('#recentTxnBody');
    if (recent.length === 0) {
      body.innerHTML = `<tr><td colspan="3" class="muted">No transactions yet.</td></tr>`;
      return;
    }
    body.innerHTML = recent.map((t) => {
      const isCredit = t.destinationAccountId && (accounts || []).some(a => a.id === t.destinationAccountId);
      const sign = isCredit ? '+' : '−';
      const cls = isCredit ? 'amount-in' : 'amount-out';
      return `
        <tr>
          <td>${escapeHtml(t.description || t.type)}</td>
          <td class="muted">${formatDate(t.createdAt)}</td>
          <td style="text-align:right;" class="${cls}">${sign}${formatCurrency(t.amount)}</td>
        </tr>`;
    }).join('');
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', loadDashboard);
