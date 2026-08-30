// accounts.js — lists every account belonging to the selected customer
// with a link through to that account's transaction history.

async function loadAccounts() {
  const customerId = requireSelectedCustomer();
  if (!customerId) return;
  const grid = qs('#accountsGrid');
  try {
    const accounts = await api.get(`/customers/${customerId}/accounts`);
    if (!accounts || accounts.length === 0) {
      grid.innerHTML = `<div class="empty-state"><p>No accounts found.</p></div>`;
      return;
    }
    grid.innerHTML = accounts.map((a) => `
      <div class="card card-hover" data-account-id="${a.id}">
        <div style="display:flex; justify-content:space-between; align-items:flex-start;">
          <div>
            <p class="faint" style="text-transform:uppercase; font-size:0.75rem; margin-bottom:4px;">${escapeHtml(a.accountType)}</p>
            <p class="mono muted">${a.maskedNumber}</p>
          </div>
          <span class="badge ${statusBadgeClass(a.status)}">${a.status}</span>
        </div>
        <h2 style="margin-top:16px;">${formatCurrency(a.availableBalance)}</h2>
        <p class="faint" style="font-size:0.78rem;">Balance: ${formatCurrency(a.balance)}</p>
      </div>`).join('');

    qsa('[data-account-id]', grid).forEach((card) => {
      card.addEventListener('click', () => {
        window.location.href = `account-details.html?id=${card.getAttribute('data-account-id')}`;
      });
    });
  } catch (err) {
    notify.error(err.message);
    grid.innerHTML = `<div class="empty-state"><p>${escapeHtml(err.message)}</p></div>`;
  }
}

document.addEventListener('DOMContentLoaded', loadAccounts);
