// transactions.js — fetches the customer's full transaction history once,
// then filters and paginates entirely client-side (fine at demo data
// scale; a real system would push filters down to the API).

const PAGE_SIZE = 10;
let allTxns = [];
let myAccountIds = [];
let currentPage = 1;

function applyFilters() {
  const q = qs('#searchInput').value.trim().toLowerCase();
  const type = qs('#typeFilter').value;
  const date = qs('#dateFilter').value;

  return allTxns.filter((t) => {
    if (type && t.type !== type) return false;
    if (date && !t.createdAt.startsWith(date)) return false;
    if (q) {
      const haystack = (t.description || '').toLowerCase() + ' ' + t.transactionRef.toLowerCase();
      if (!haystack.includes(q)) return false;
    }
    return true;
  });
}

function renderPage() {
  const filtered = applyFilters();
  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  currentPage = Math.min(currentPage, totalPages);
  const pageItems = filtered.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE);

  const body = qs('#txnBody');
  if (pageItems.length === 0) {
    body.innerHTML = `<tr><td colspan="6" class="muted">No transactions match your filters.</td></tr>`;
  } else {
    body.innerHTML = pageItems.map((t) => {
      const isCredit = myAccountIds.includes(t.destinationAccountId);
      const cls = isCredit ? 'amount-in' : 'amount-out';
      const sign = isCredit ? '+' : '−';
      return `<tr>
        <td class="mono faint">${t.transactionRef}</td>
        <td><span class="badge badge-neutral">${t.type}</span></td>
        <td>${escapeHtml(t.description || '')}</td>
        <td class="muted">${formatDate(t.createdAt)}</td>
        <td><span class="badge ${statusBadgeClass(t.status)}">${t.status}</span></td>
        <td style="text-align:right;" class="${cls}">${sign}${formatCurrency(t.amount)}</td>
      </tr>`;
    }).join('');
  }

  const pagination = qs('#pagination');
  let buttons = '';
  for (let i = 1; i <= totalPages; i++) {
    buttons += `<button class="${i === currentPage ? 'active' : ''}" data-page-btn="${i}">${i}</button>`;
  }
  pagination.innerHTML = buttons;
  qsa('[data-page-btn]', pagination).forEach((b) => {
    b.addEventListener('click', () => { currentPage = Number(b.dataset.pageBtn); renderPage(); });
  });
}

async function loadTransactions() {
  const customerId = requireSelectedCustomer();
  if (!customerId) return;
  try {
    const [txns, accounts] = await Promise.all([
      api.get(`/customers/${customerId}/transactions`),
      api.get(`/customers/${customerId}/accounts`),
    ]);
    allTxns = txns || [];
    myAccountIds = (accounts || []).map((a) => a.id);
    renderPage();
  } catch (err) {
    notify.error(err.message);
    qs('#txnBody').innerHTML = `<tr><td colspan="6" class="muted">${escapeHtml(err.message)}</td></tr>`;
  }
}

document.addEventListener('DOMContentLoaded', () => {
  loadTransactions();
  qs('#searchInput').addEventListener('input', debounce(() => { currentPage = 1; renderPage(); }, 200));
  qs('#typeFilter').addEventListener('change', () => { currentPage = 1; renderPage(); });
  qs('#dateFilter').addEventListener('change', () => { currentPage = 1; renderPage(); });
});
