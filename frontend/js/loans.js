// loans.js — EMI calculator (calls the backend so the formula lives in
// one place), loan application, and the customer's existing loans list.

let customerId;
let myAccounts = [];

async function calculate() {
  const principal = Number(qs('#calcPrincipal').value);
  const rate = Number(qs('#calcRate').value);
  const tenure = Number(qs('#calcTenure').value);
  try {
    const estimate = await api.post('/loans/estimate', { principal, interestRate: rate, tenureMonths: tenure });
    qs('#calcResult').innerHTML = `
      <div class="modal-row"><span class="muted">Monthly EMI</span><strong>${formatCurrency(estimate.monthlyEmi)}</strong></div>
      <div class="modal-row"><span class="muted">Total Interest</span><strong>${formatCurrency(estimate.totalInterest)}</strong></div>
      <div class="modal-row"><span class="muted">Total Repayment</span><strong>${formatCurrency(estimate.totalRepayment)}</strong></div>`;
  } catch (err) {
    notify.error(err.message);
  }
}

async function applyForLoan() {
  if (myAccounts.length === 0) {
    notify.warning('You need at least one account to apply for a loan.');
    return;
  }
  const principal = Number(qs('#calcPrincipal').value);
  const rate = Number(qs('#calcRate').value);
  const tenure = Number(qs('#calcTenure').value);

  const confirmed = await confirmModal({
    title: 'Apply for Simulated Loan',
    message: `Apply for a ${formatCurrency(principal)} loan at ${rate}% over ${tenure} months? This is a simulation only.`,
    confirmLabel: 'Apply',
  });
  if (!confirmed) return;

  try {
    await api.post('/loans/apply', {
      customerId,
      accountId: myAccounts[0].id,
      principal, interestRate: rate, tenureMonths: tenure,
    });
    notify.success('Loan application submitted (simulation)');
    loadLoans();
  } catch (err) {
    notify.error(err.message);
  }
}

async function loadLoans() {
  const list = qs('#loansList');
  try {
    const loans = await api.get(`/customers/${customerId}/loans`);
    if (!loans || loans.length === 0) {
      list.innerHTML = `<div class="empty-state"><div class="icon">🏦</div><p>No loans on file.</p></div>`;
      return;
    }
    list.innerHTML = loans.map((l) => `
      <div class="card" style="margin-bottom:12px;">
        <div style="display:flex; justify-content:space-between; align-items:flex-start;">
          <div>
            <h3>${formatCurrency(l.principalAmount)}</h3>
            <p class="muted">${l.interestRate}% · ${l.tenureMonths} months</p>
          </div>
          <span class="badge ${statusBadgeClass(l.status)}">${l.status}</span>
        </div>
        <p class="faint" style="font-size:0.85rem;">EMI: ${formatCurrency(l.emiAmount)}/month</p>
      </div>`).join('');
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', async () => {
  customerId = requireSelectedCustomer();
  if (!customerId) return;
  myAccounts = await api.get(`/customers/${customerId}/accounts`).catch(() => []);
  qs('#calcBtn').addEventListener('click', calculate);
  qs('#applyBtn').addEventListener('click', applyForLoan);
  calculate();
  loadLoans();
});
