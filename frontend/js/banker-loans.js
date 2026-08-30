// banker-loans.js — read-only overview of every loan application.

async function loadLoans() {
  try {
    const list = await api.get('/banker/loans') || [];
    const body = qs('#dataBody');
    if (list.length === 0) {
      body.innerHTML = `<tr><td colspan="6" class="muted">No loan applications.</td></tr>`;
      return;
    }
    body.innerHTML = list.map((l) => `
      <tr>
        <td>${l.customerId}</td>
        <td>${formatCurrency(l.principalAmount)}</td>
        <td>${l.interestRate}%</td>
        <td>${l.tenureMonths} mo</td>
        <td>${formatCurrency(l.emiAmount)}</td>
        <td><span class="badge ${statusBadgeClass(l.status)}">${l.status}</span></td>
      </tr>`).join('');
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  loadLoans();
  const toolbar = qs('#searchInput');
  if (toolbar) toolbar.closest('.table-toolbar').classList.add('hidden');
});
