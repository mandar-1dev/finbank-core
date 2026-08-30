// customer-selection.js — lists every seeded customer with their primary
// account balance and lets the user pick one to "enter" as.

async function loadCustomers() {
  const grid = qs('#customerGrid');
  try {
    const customers = await api.get('/customers');
    if (!customers || customers.length === 0) {
      grid.innerHTML = `<div class="empty-state"><div class="icon">🏦</div><p>No demo customers found. Did you run database/seed.sql?</p></div>`;
      return;
    }

    // Fetch each customer's accounts in parallel so we can show a primary
    // account + balance on every card.
    const withAccounts = await Promise.all(customers.map(async (c) => {
      const accounts = await api.get(`/customers/${c.id}/accounts`).catch(() => []);
      return { customer: c, accounts };
    }));

    grid.innerHTML = withAccounts.map(({ customer, accounts }) => {
      const primary = accounts[0];
      return `
        <div class="card card-hover" data-customer-id="${customer.id}">
          <div class="emoji" style="font-size:1.8rem;">👤</div>
          <h3>${escapeHtml(customer.fullName)}</h3>
          ${primary ? `
            <p class="muted" style="margin-bottom:2px;">${escapeHtml(primary.accountType.charAt(0) + primary.accountType.slice(1).toLowerCase())} Account</p>
            <p class="mono muted" style="margin-bottom:14px;">${primary.maskedNumber}</p>
            <p class="faint" style="font-size:0.78rem; text-transform:uppercase; margin-bottom:2px;">Available Balance</p>
            <h2>${formatCurrency(primary.availableBalance)}</h2>
          ` : `<p class="faint">No accounts on file</p>`}
          <button class="btn btn-primary btn-block" style="margin-top:14px;">Open Dashboard</button>
        </div>`;
    }).join('');

    qsa('[data-customer-id]', grid).forEach((card) => {
      card.addEventListener('click', () => {
        setSelectedCustomer(card.getAttribute('data-customer-id'));
        window.location.href = 'dashboard.html';
      });
    });
  } catch (err) {
    grid.innerHTML = `<div class="empty-state"><div class="icon">⚠</div><p>${escapeHtml(err.message)}</p></div>`;
  }
}

document.addEventListener('DOMContentLoaded', loadCustomers);
