// notifications.js — small toast system so we never use window.alert().

function ensureToastStack() {
  let stack = document.querySelector('.toast-stack');
  if (!stack) {
    stack = document.createElement('div');
    stack.className = 'toast-stack';
    document.body.appendChild(stack);
  }
  return stack;
}

function showToast(message, type) {
  const stack = ensureToastStack();
  const toast = document.createElement('div');
  toast.className = 'toast' + (type ? ' ' + type : '');
  const icon = type === 'success' ? '✓' : type === 'error' ? '✕' : type === 'warning' ? '⚠' : 'ℹ';
  toast.innerHTML = `<span>${icon}</span><span>${escapeHtml(message)}</span>`;
  stack.appendChild(toast);
  setTimeout(() => {
    toast.style.transition = 'opacity 0.2s ease';
    toast.style.opacity = '0';
    setTimeout(() => toast.remove(), 200);
  }, 3200);
}

const notify = {
  success: (msg) => showToast(msg, 'success'),
  error: (msg) => showToast(msg, 'error'),
  warning: (msg) => showToast(msg, 'warning'),
  info: (msg) => showToast(msg, 'info'),
};

// Minimal confirm-style modal to replace window.confirm() for destructive
// actions like removing a beneficiary or freezing an account.
function confirmModal({ title, message, confirmLabel, danger }) {
  return new Promise((resolve) => {
    const backdrop = document.createElement('div');
    backdrop.className = 'modal-backdrop';
    backdrop.innerHTML = `
      <div class="modal">
        <h3>${escapeHtml(title)}</h3>
        <p>${escapeHtml(message)}</p>
        <div class="modal-actions">
          <button class="btn btn-secondary btn-block" data-action="cancel">Cancel</button>
          <button class="btn ${danger ? 'btn-danger' : 'btn-primary'} btn-block" data-action="confirm">${escapeHtml(confirmLabel || 'Confirm')}</button>
        </div>
      </div>`;
    document.body.appendChild(backdrop);

    backdrop.addEventListener('click', (e) => {
      if (e.target === backdrop) { backdrop.remove(); resolve(false); }
    });
    backdrop.querySelector('[data-action="cancel"]').onclick = () => { backdrop.remove(); resolve(false); };
    backdrop.querySelector('[data-action="confirm"]').onclick = () => { backdrop.remove(); resolve(true); };
  });
}
