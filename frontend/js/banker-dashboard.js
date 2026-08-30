// banker-dashboard.js — pulls /api/banker/dashboard once and paints the
// stat cards + three Chart.js visualisations.

async function loadBankerDashboard() {
  try {
    const data = await api.get('/banker/dashboard');
    const s = data.stats;

    qs('#statsGrid').innerHTML = `
      ${statCard('Total Customers', s.totalCustomers)}
      ${statCard('Total Accounts', s.totalAccounts)}
      ${statCard('Active Accounts', s.activeAccounts, 'success')}
      ${statCard('Frozen Accounts', s.frozenAccounts, 'danger')}
      ${statCard('Transactions Today', s.transactionsToday)}
      ${statCard('Volume Today', formatCurrency(s.transactionVolumeToday))}
      ${statCard('Active Loans', s.activeLoans)}
      ${statCard('Pending Loans', s.pendingLoans, 'warning')}
      ${statCard('Suspicious Txns', s.suspiciousTransactions, 'danger')}
    `;

    const volume = data.dailyVolume || [];
    renderLineChart('volumeChart',
      volume.map(v => new Date(v.date).toLocaleDateString('en-IN', { day: '2-digit', month: 'short' })),
      volume.map(v => v.volume), 'Transaction Volume (₹)');

    const types = data.accountTypeDistribution || {};
    renderDoughnutChart('accountTypeChart', Object.keys(types), Object.values(types));

    const loanStatus = data.loanStatusDistribution || {};
    renderBarChart('loanStatusChart', Object.keys(loanStatus), Object.values(loanStatus), 'Loans');
  } catch (err) {
    notify.error(err.message);
  }
}

function statCard(label, value, tone) {
  const toneColor = tone === 'success' ? 'var(--color-success)' : tone === 'danger' ? 'var(--color-danger)' : tone === 'warning' ? 'var(--color-warning)' : 'var(--color-text)';
  return `<div class="stat-card"><div class="stat-label">${escapeHtml(label)}</div><div class="stat-value" style="color:${toneColor}">${value}</div></div>`;
}

document.addEventListener('DOMContentLoaded', loadBankerDashboard);
