// navigation.js — shared chrome behaviour for every customer-facing page:
// highlights the active sidebar link and wires up "Switch Customer" /
// "Back to Home", both of which clear the selected demo customer.

function highlightActiveNav() {
  const current = window.location.pathname.split('/').pop() || 'dashboard.html';
  qsa('.sidebar a[data-page]').forEach((a) => {
    a.classList.toggle('active', a.getAttribute('data-page') === current);
  });
}

function wireSwitchCustomer() {
  const el = qs('[data-action="switch-customer"]');
  if (el) {
    el.addEventListener('click', (e) => {
      e.preventDefault();
      clearSelectedCustomer();
      window.location.href = 'customer-selection.html';
    });
  }
}

function wireBackToHome() {
  qsa('[data-action="back-home"]').forEach((el) => {
    el.addEventListener('click', (e) => {
      e.preventDefault();
      clearSelectedCustomer();
      window.location.href = 'index.html';
    });
  });
}

// Populates the little "Hello, X" avatar chip in the topbar using the
// currently selected customer. Safe to call on every customer page.
async function paintTopbarCustomer() {
  const chip = qs('[data-topbar-customer]');
  if (!chip) return;
  const id = getSelectedCustomer();
  if (!id) return;
  try {
    const customer = await api.get(`/customers/${id}`);
    chip.innerHTML = `
      <span class="avatar">${initials(customer.fullName)}</span>
      <span class="name">${escapeHtml(customer.fullName.split(' ')[0])}</span>`;
  } catch (_) {
    // silently ignore — page-level loaders will surface any real error
  }
}

function initCustomerChrome() {
  highlightActiveNav();
  wireSwitchCustomer();
  wireBackToHome();
  paintTopbarCustomer();
}

document.addEventListener('DOMContentLoaded', () => {
  if (document.body.dataset.chrome === 'customer') {
    initCustomerChrome();
  } else if (document.body.dataset.chrome === 'banker') {
    highlightActiveNav();
    wireBackToHome();
  }
});
