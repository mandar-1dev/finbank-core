// transfer.js — handles three modes (transfer / deposit / withdraw) on one
// page, each with a confirmation modal before hitting the API, a loading
// state that disables the submit button, and a toast on success/failure.

let currentMode = 'transfer';
let myAccounts = [];
let myBeneficiaries = [];
let allAccountsByNumber = {};

function setMode(mode) {
  currentMode = mode;
  qsa('[data-mode]').forEach((b) => b.classList.toggle('btn-primary', b.dataset.mode === mode));
  qsa('[data-mode]').forEach((b) => b.classList.toggle('btn-secondary', b.dataset.mode !== mode));

  const toField = qs('#toAccountField');
  const title = qs('#formTitle');
  const submitBtn = qs('#submitBtn');

  if (mode === 'transfer') {
    toField.classList.remove('hidden');
    title.textContent = 'Transfer Money';
    submitBtn.textContent = 'Review Transfer';
  } else if (mode === 'deposit') {
    toField.classList.add('hidden');
    title.textContent = 'Deposit';
    submitBtn.textContent = 'Review Deposit';
  } else {
    toField.classList.add('hidden');
    title.textContent = 'Withdraw';
    submitBtn.textContent = 'Review Withdrawal';
  }
}

async function loadFormData(customerId) {
  const [accounts, beneficiaries, allAccounts] = await Promise.all([
    api.get(`/customers/${customerId}/accounts`),
    api.get(`/customers/${customerId}/beneficiaries`),
    api.get(`/accounts`),
  ]);
  myAccounts = accounts || [];
  myBeneficiaries = beneficiaries || [];
  allAccountsByNumber = {};
  (allAccounts || []).forEach((a) => { allAccountsByNumber[a.accountNumber] = a; });

  qs('#fromAccount').innerHTML = myAccounts.map((a) =>
    `<option value="${a.id}">${a.accountType} · ${a.maskedNumber} · ${formatCurrency(a.availableBalance)}</option>`
  ).join('');

  qs('#toBeneficiary').innerHTML = myBeneficiaries.map((b) =>
    `<option value="${b.accountNumber}">${escapeHtml(b.nickname || b.name)} · ****${b.accountNumber.slice(-4)}</option>`
  ).join('') || `<option value="">No beneficiaries added yet</option>`;
}

function buildConfirmationRows() {
  const fromId = qs('#fromAccount').value;
  const fromAccount = myAccounts.find((a) => String(a.id) === fromId);
  const amount = Number(qs('#amount').value);
  const description = qs('#description').value;

  if (currentMode === 'transfer') {
    const toNumber = qs('#toBeneficiary').value;
    const beneficiary = myBeneficiaries.find((b) => b.accountNumber === toNumber);
    return {
      rows: [
        ['From', `${fromAccount ? fromAccount.maskedNumber : ''}`],
        ['To', beneficiary ? `${beneficiary.nickname || beneficiary.name} · ****${toNumber.slice(-4)}` : ''],
        ['Amount', formatCurrency(amount)],
        ['Description', description || '—'],
      ],
      title: 'Confirm Transfer',
    };
  }
  return {
    rows: [
      ['Account', fromAccount ? fromAccount.maskedNumber : ''],
      ['Amount', formatCurrency(amount)],
      ['Description', description || '—'],
    ],
    title: currentMode === 'deposit' ? 'Confirm Deposit' : 'Confirm Withdrawal',
  };
}

function showConfirmation({ title, rows }) {
  return new Promise((resolve) => {
    const backdrop = document.createElement('div');
    backdrop.className = 'modal-backdrop';
    backdrop.innerHTML = `
      <div class="modal">
        <h3>${escapeHtml(title)}</h3>
        ${rows.map(([k, v]) => `<div class="modal-row"><span class="muted">${escapeHtml(k)}</span><strong>${v}</strong></div>`).join('')}
        <div class="modal-actions">
          <button class="btn btn-secondary btn-block" data-action="cancel">Cancel</button>
          <button class="btn btn-primary btn-block" data-action="confirm">Confirm</button>
        </div>
      </div>`;
    document.body.appendChild(backdrop);
    backdrop.querySelector('[data-action="cancel"]').onclick = () => { backdrop.remove(); resolve(false); };
    backdrop.querySelector('[data-action="confirm"]').onclick = () => { backdrop.remove(); resolve(true); };
  });
}

function generateIdempotencyKey() {
  if (window.crypto && crypto.randomUUID) return crypto.randomUUID();
  return 'key-' + Date.now() + '-' + Math.random().toString(36).slice(2);
}

async function submitTransaction() {
  const fromId = Number(qs('#fromAccount').value);
  const amount = Number(qs('#amount').value);
  const description = qs('#description').value;

  if (currentMode === 'transfer') {
    const toNumber = qs('#toBeneficiary').value;
    const toAccount = allAccountsByNumber[toNumber];
    if (!toAccount) throw new Error('Selected beneficiary account could not be resolved.');
    return api.post('/transactions/transfer',
      { fromAccountId: fromId, toAccountId: toAccount.id, amount, description },
      { 'Idempotency-Key': generateIdempotencyKey() });
  }
  if (currentMode === 'deposit') {
    return api.post('/transactions/deposit', { accountId: fromId, amount, description });
  }
  return api.post('/transactions/withdraw', { accountId: fromId, amount, description });
}

async function handleSubmit(e) {
  e.preventDefault();
  if (currentMode === 'transfer' && myBeneficiaries.length === 0) {
    notify.warning('Add a beneficiary first to make a transfer.');
    return;
  }

  const confirmed = await showConfirmation(buildConfirmationRows());
  if (!confirmed) return;

  const submitBtn = qs('#submitBtn');
  const originalLabel = submitBtn.textContent;
  submitBtn.disabled = true;
  submitBtn.innerHTML = `<span class="spinner"></span> Processing…`;

  try {
    await submitTransaction();
    notify.success(currentMode === 'transfer' ? 'Transfer successful' : currentMode === 'deposit' ? 'Deposit successful' : 'Withdrawal successful');
    qs('#transferForm').reset();
    await loadFormData(getSelectedCustomer());
    setTimeout(() => { window.location.href = 'dashboard.html'; }, 900);
  } catch (err) {
    notify.error(err.message);
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = originalLabel;
  }
}

document.addEventListener('DOMContentLoaded', async () => {
  const customerId = requireSelectedCustomer();
  if (!customerId) return;

  qsa('[data-mode]').forEach((btn) => btn.addEventListener('click', () => setMode(btn.dataset.mode)));

  const params = new URLSearchParams(window.location.search);
  const initialMode = params.get('action');
  setMode(initialMode === 'deposit' || initialMode === 'withdraw' ? initialMode : 'transfer');

  try {
    await loadFormData(customerId);
  } catch (err) {
    notify.error(err.message);
  }

  qs('#transferForm').addEventListener('submit', handleSubmit);
});
