// utils.js — pure formatting/helper functions shared across pages.

function formatCurrency(amount) {
  const n = Number(amount) || 0;
  return '₹' + n.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatDate(isoString) {
  if (!isoString) return '—';
  const d = new Date(isoString);
  return d.toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' }) +
    ' · ' + d.toLocaleTimeString('en-IN', { hour: '2-digit', minute: '2-digit' });
}

function statusBadgeClass(status) {
  switch (status) {
    case 'ACTIVE': case 'SUCCESS': case 'APPROVED': case 'CLOSED':
      return 'badge-success';
    case 'FROZEN': case 'FAILED': case 'REJECTED': case 'HIGH':
      return 'badge-danger';
    case 'PENDING': case 'MEDIUM':
      return 'badge-warning';
    default:
      return 'badge-neutral';
  }
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str == null ? '' : String(str);
  return div.innerHTML;
}

function qs(selector, root) { return (root || document).querySelector(selector); }
function qsa(selector, root) { return Array.from((root || document).querySelectorAll(selector)); }

function debounce(fn, delay) {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}

function initials(fullName) {
  if (!fullName) return '?';
  const parts = fullName.trim().split(/\s+/);
  return (parts[0][0] + (parts[1] ? parts[1][0] : '')).toUpperCase();
}
