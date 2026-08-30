// customer-context.js — tracks which predefined demo customer is
// "selected" for the current browser session. This is purely a UI
// convenience, never authentication: anyone can pick any customer at any
// time, and no credentials are involved anywhere in this flow.

const CUSTOMER_STORAGE_KEY = 'selectedCustomerId';

function setSelectedCustomer(customerId) {
  localStorage.setItem(CUSTOMER_STORAGE_KEY, String(customerId));
}

function getSelectedCustomer() {
  const raw = localStorage.getItem(CUSTOMER_STORAGE_KEY);
  return raw ? Number(raw) : null;
}

function clearSelectedCustomer() {
  localStorage.removeItem(CUSTOMER_STORAGE_KEY);
}

// Call this at the top of any page that requires a selected customer.
// Redirects to customer-selection.html if none is set.
function requireSelectedCustomer() {
  const id = getSelectedCustomer();
  if (!id) {
    window.location.href = 'customer-selection.html';
    return null;
  }
  return id;
}
