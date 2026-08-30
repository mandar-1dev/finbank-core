// customers.js (banker) — searchable customer list linking to details.

let allCustomers = [];

function render() {
  const q = qs('#searchInput').value.trim().toLowerCase();
  const filtered = allCustomers.filter((c) =>
    !q || c.fullName.toLowerCase().includes(q) || c.customerNumber.toLowerCase().includes(q));

  const body = qs('#custBody');
  if (filtered.length === 0) {
    body.innerHTML = `<tr><td colspan="6" class="muted">No customers match your search.</td></tr>`;
    return;
  }
  body.innerHTML = filtered.map((c) => `
    <tr>
      <td class="mono">${escapeHtml(c.customerNumber)}</td>
      <td>${escapeHtml(c.fullName)}</td>
      <td class="muted">${escapeHtml(c.email)}</td>
      <td class="muted">${escapeHtml(c.phone)}</td>
      <td><span class="badge ${statusBadgeClass(c.status)}">${c.status}</span></td>
      <td><a class="btn btn-secondary btn-sm" href="customer-details.html?id=${c.id}">View</a></td>
    </tr>`).join('');
}

async function loadCustomers() {
  try {
    allCustomers = await api.get('/banker/customers') || [];
    render();
  } catch (err) {
    notify.error(err.message);
    qs('#custBody').innerHTML = `<tr><td colspan="6" class="muted">${escapeHtml(err.message)}</td></tr>`;
  }
}

document.addEventListener('DOMContentLoaded', () => {
  loadCustomers();
  qs('#searchInput').addEventListener('input', debounce(render, 150));
});
