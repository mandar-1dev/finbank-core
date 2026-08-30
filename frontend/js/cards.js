// cards.js — renders each card as a small visual, with actions to freeze,
// unfreeze, and simulate a merchant payment.

let customerId;

async function toggleFreeze(cardId, currentStatus) {
  const action = currentStatus === 'ACTIVE' ? 'freeze' : 'unfreeze';
  try {
    await api.post(`/cards/${cardId}/${action}`);
    notify.success(action === 'freeze' ? 'Card frozen' : 'Card unfrozen');
    loadCards();
  } catch (err) {
    notify.error(err.message);
  }
}

function openPaymentModal(cardId) {
  const backdrop = document.createElement('div');
  backdrop.className = 'modal-backdrop';
  backdrop.innerHTML = `
    <div class="modal">
      <h3>Simulate Card Payment</h3>
      <form id="payForm">
        <div class="field"><label>Merchant Name</label><input type="text" id="merchant" required placeholder="e.g. Amazon Simulation"></div>
        <div class="field"><label>Amount (₹)</label><input type="number" id="payAmount" min="1" step="0.01" required></div>
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary btn-block" data-action="cancel">Cancel</button>
          <button type="submit" class="btn btn-primary btn-block">Pay</button>
        </div>
      </form>
    </div>`;
  document.body.appendChild(backdrop);
  backdrop.querySelector('[data-action="cancel"]').onclick = () => backdrop.remove();
  backdrop.querySelector('#payForm').onsubmit = async (e) => {
    e.preventDefault();
    try {
      await api.post(`/cards/${cardId}/pay`, {
        merchant: qs('#merchant', backdrop).value,
        amount: Number(qs('#payAmount', backdrop).value),
      });
      notify.success('Payment simulation successful');
      backdrop.remove();
      loadCards();
    } catch (err) {
      notify.error(err.message);
    }
  };
}

async function loadCards() {
  const grid = qs('#cardsGrid');
  try {
    const cards = await api.get(`/customers/${customerId}/cards`);
    if (!cards || cards.length === 0) {
      grid.innerHTML = `<div class="empty-state"><div class="icon">🪪</div><p>No cards issued yet.</p></div>`;
      return;
    }
    grid.innerHTML = cards.map((c) => `
      <div>
        <div class="card-visual ${c.status === 'FROZEN' ? 'frozen' : ''}">
          <div class="brand-mini">COREBANK</div>
          <div class="number">${c.maskedNumber}</div>
          <div class="meta"><span>${escapeHtml(c.cardHolder)}</span><span>EXP ${String(c.expiryMonth).padStart(2, '0')}/${String(c.expiryYear).slice(-2)}</span></div>
        </div>
        <div style="display:flex; justify-content:space-between; align-items:center; margin-top:10px;">
          <span class="badge ${statusBadgeClass(c.status)}">${c.status}</span>
          <span class="faint" style="font-size:0.8rem;">Limit ${formatCurrency(c.spendingLimit)}</span>
        </div>
        <div style="display:flex; gap:8px; margin-top:10px;">
          <button class="btn btn-secondary btn-sm" data-freeze="${c.id}" data-status="${c.status}">${c.status === 'ACTIVE' ? 'Freeze' : 'Unfreeze'}</button>
          <button class="btn btn-primary btn-sm" data-pay="${c.id}" ${c.status !== 'ACTIVE' ? 'disabled' : ''}>Simulate Payment</button>
        </div>
      </div>`).join('');

    qsa('[data-freeze]', grid).forEach((btn) => {
      btn.addEventListener('click', () => toggleFreeze(btn.dataset.freeze, btn.dataset.status));
    });
    qsa('[data-pay]', grid).forEach((btn) => {
      btn.addEventListener('click', () => openPaymentModal(btn.dataset.pay));
    });
  } catch (err) {
    notify.error(err.message);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  customerId = requireSelectedCustomer();
  if (!customerId) return;
  loadCards();
});
