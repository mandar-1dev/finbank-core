// beneficiaries.js — view, add, and remove beneficiaries for the selected
// customer. Uses a lightweight in-page modal for the add form instead of
// navigating to a separate page.

let customerId;

function openAddModal() {
  const backdrop = document.createElement('div');
  backdrop.className = 'modal-backdrop';
  backdrop.innerHTML = `
    <div class="modal">
      <h3>Add Beneficiary</h3>
      <form id="addBeneficiaryForm">
        <div class="field"><label>Full Name</label><input type="text" id="benName" required></div>
        <div class="field"><label>Account Number</label><input type="text" id="benAccount" required maxlength="16" pattern="[0-9]{6,16}"></div>
        <div class="field"><label>Nickname (optional)</label><input type="text" id="benNickname"></div>
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary btn-block" data-action="cancel">Cancel</button>
          <button type="submit" class="btn btn-primary btn-block">Add</button>
        </div>
      </form>
    </div>`;
  document.body.appendChild(backdrop);
  backdrop.querySelector('[data-action="cancel"]').onclick = () => backdrop.remove();
  backdrop.querySelector('#addBeneficiaryForm').onsubmit = async (e) => {
    e.preventDefault();
    try {
      await api.post('/beneficiaries', {
        customerId,
        name: qs('#benName', backdrop).value,
        accountNumber: qs('#benAccount', backdrop).value,
        ifscCode: 'CORE0000001',
        nickname: qs('#benNickname', backdrop).value,
      });
      notify.success('Beneficiary added');
      backdrop.remove();
      loadBeneficiaries();
    } catch (err) {
      notify.error(err.message);
    }
  };
}

async function removeBeneficiary(id) {
  const confirmed = await confirmModal({
    title: 'Remove Beneficiary',
    message: 'Are you sure you want to remove this beneficiary?',
    confirmLabel: 'Remove',
    danger: true,
  });
  if (!confirmed) return;
  try {
    await api.delete(`/beneficiaries/${id}?customerId=${customerId}`);
    notify.success('Beneficiary removed');
    loadBeneficiaries();
  } catch (err) {
    notify.error(err.message);
  }
}

async function loadBeneficiaries() {
  const grid = qs('#beneficiaryGrid');
  try {
    const list = await api.get(`/customers/${customerId}/beneficiaries`);
    if (!list || list.length === 0) {
      grid.innerHTML = `<div class="empty-state"><div class="icon">👥</div><p>No beneficiaries yet. Add one to start transferring.</p></div>`;
      return;
    }
    grid.innerHTML = list.map((b) => `
      <div class="card">
        <h3>${escapeHtml(b.nickname || b.name)}</h3>
        <p class="muted">${escapeHtml(b.name)}</p>
        <p class="mono faint">****${b.accountNumber.slice(-4)}</p>
        <div style="display:flex; gap:8px; margin-top:14px;">
          <a class="btn btn-secondary btn-sm" href="transfer.html">Transfer</a>
          <button class="btn btn-danger btn-sm" data-remove="${b.id}">Remove</button>
        </div>
      </div>`).join('');
    qsa('[data-remove]', grid).forEach((btn) => {
      btn.addEventListener('click', () => removeBeneficiary(btn.dataset.remove));
    });
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  customerId = requireSelectedCustomer();
  if (!customerId) return;
  qs('#addBeneficiaryBtn').addEventListener('click', openAddModal);
  loadBeneficiaries();
});
