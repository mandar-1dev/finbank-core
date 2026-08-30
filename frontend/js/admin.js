/**
 * CoreBank — admin.js
 */

async function loadAdminPage() {
    auth.requireAdmin();
    utils.renderShell('admin.html');
    document.getElementById('page-title').textContent = 'Admin Dashboard';

    const content = document.getElementById('page-content');
    content.innerHTML = `
        <div class="page-header"><div><h1>Admin Dashboard</h1><p>Platform-wide overview and controls.</p></div></div>

        <div class="grid-4" id="admin-stats" style="margin-bottom:32px;">
            ${Array.from({ length: 8 }).map(() => '<div class="skeleton" style="height:90px;border-radius:10px;"></div>').join('')}
        </div>

        <div class="section-title"><h2>Customer Search</h2></div>
        <div class="filter-bar">
            <div class="form-group" style="flex:2;">
                <label class="form-label">Search by name, email or customer number</label>
                <input type="text" class="form-control" id="customer-search" placeholder="e.g. Rahul Sharma">
            </div>
            <button class="btn btn-primary" id="search-btn">Search</button>
        </div>
        <div class="table-wrap" id="customer-results" style="margin-bottom:32px;"></div>

        <div class="section-title"><h2>Recent Audit Log</h2></div>
        <div class="table-wrap" id="audit-log">
            <div style="padding:16px;"><div class="skeleton" style="height:20px;border-radius:6px;"></div></div>
        </div>
    `;

    loadDashboardStats();
    loadAuditLogs();

    document.getElementById('search-btn').addEventListener('click', () => searchCustomers());
    document.getElementById('customer-search').addEventListener('keydown', (e) => {
        if (e.key === 'Enter') searchCustomers();
    });
    searchCustomers(); // initial unfiltered load
}

async function loadDashboardStats() {
    try {
        const d = await api.get('/admin/dashboard');
        const stats = [
            ['Total Customers', d.totalCustomers],
            ['Total Accounts', d.totalAccounts],
            ['Active Accounts', d.activeAccounts],
            ['Frozen Accounts', d.frozenAccounts],
            ['Transactions Today', d.transactionsToday],
            ['Volume Today', utils.formatCurrency(d.transactionVolumeToday)],
            ['Pending Loans', d.pendingLoans],
            ['Suspicious Txns (24h)', d.suspiciousTransactions],
        ];
        document.getElementById('admin-stats').innerHTML = stats.map(([label, value]) => `
            <div class="stat-card">
                <div class="stat-card__label">${label}</div>
                <div class="stat-card__value">${value}</div>
            </div>`).join('');
    } catch (err) {
        notify.error(err.message || 'Could not load dashboard stats');
    }
}

async function searchCustomers() {
    const query = document.getElementById('customer-search').value.trim();
    const mount = document.getElementById('customer-results');
    mount.innerHTML = '<div style="padding:16px;"><div class="skeleton" style="height:20px;border-radius:6px;"></div></div>';

    try {
        const qs = query ? `?query=${encodeURIComponent(query)}` : '';
        const result = await api.get(`/admin/customers${qs}`);
        if (result.content.length === 0) {
            mount.innerHTML = `<div class="empty-state"><h3>No customers found</h3><p>Try a different search term.</p></div>`;
            return;
        }
        mount.innerHTML = `
            <table class="data-table">
                <thead><tr><th>Customer #</th><th>Name</th><th>Email</th><th>Role</th><th>Status</th><th></th></tr></thead>
                <tbody>
                    ${result.content.map((c) => `
                        <tr>
                            <td class="mono">${c.customerNumber}</td>
                            <td>${c.fullName}</td>
                            <td>${c.email}</td>
                            <td><span class="badge badge-neutral">${c.role}</span></td>
                            <td><span class="badge ${utils.statusBadgeClass(c.status)}">${c.status}</span></td>
                            <td><button class="btn btn-sm btn-secondary" onclick="viewCustomerAccounts(${c.id}, '${c.fullName.replace(/'/g, "\\'")}')">View Accounts</button></td>
                        </tr>`).join('')}
                </tbody>
            </table>`;
    } catch (err) {
        mount.innerHTML = `<div class="error-state"><h3>Search failed</h3><p>${err.message}</p></div>`;
    }
}

async function viewCustomerAccounts(customerId, name) {
    const overlay = utils.el('div', { class: 'modal-overlay' });
    overlay.innerHTML = `<div class="modal" style="max-width:560px;"><h3>${name}'s Accounts</h3><div id="modal-account-list"><div class="skeleton" style="height:60px;border-radius:8px;"></div></div><div class="modal-actions"><button class="btn btn-secondary btn-block" id="close-modal">Close</button></div></div>`;
    document.body.appendChild(overlay);
    overlay.querySelector('#close-modal').addEventListener('click', () => overlay.remove());

    try {
        const accounts = await api.get(`/admin/customers/${customerId}/accounts`);
        overlay.querySelector('#modal-account-list').innerHTML = accounts.map((a) => `
            <div class="flex-between" style="padding:12px 0;border-bottom:1px solid var(--border);">
                <div>
                    <div class="mono" style="font-weight:600;">${a.maskedAccountNumber}</div>
                    <div class="text-muted" style="font-size:12px;">${a.accountType} · ${utils.formatCurrency(a.balance)}</div>
                </div>
                <div style="display:flex;gap:8px;align-items:center;">
                    <span class="badge ${utils.statusBadgeClass(a.status)}">${a.status}</span>
                    ${a.status === 'FROZEN'
                        ? `<button class="btn btn-sm btn-secondary" onclick="adminToggleFreeze(${a.id}, false, ${customerId}, '${name.replace(/'/g, "\\'")}')">Unfreeze</button>`
                        : `<button class="btn btn-sm btn-danger" onclick="adminToggleFreeze(${a.id}, true, ${customerId}, '${name.replace(/'/g, "\\'")}')">Freeze</button>`}
                </div>
            </div>`).join('') || '<p class="text-muted">No accounts found.</p>';
    } catch (err) {
        overlay.querySelector('#modal-account-list').innerHTML = `<p class="text-danger">${err.message}</p>`;
    }
}

async function adminToggleFreeze(accountId, freeze, customerId, name) {
    try {
        await api.post(`/admin/accounts/${accountId}/${freeze ? 'freeze' : 'unfreeze'}`, {});
        notify.success(freeze ? 'Account frozen' : 'Account unfrozen');
        document.querySelectorAll('.modal-overlay').forEach((el) => el.remove());
        viewCustomerAccounts(customerId, name);
        loadDashboardStats();
        loadAuditLogs();
    } catch (err) {
        notify.error(err.message || 'Could not update account');
    }
}

async function loadAuditLogs() {
    const mount = document.getElementById('audit-log');
    try {
        const result = await api.get('/admin/audit-logs?size=15');
        if (result.content.length === 0) {
            mount.innerHTML = `<div class="empty-state"><h3>No audit events yet</h3></div>`;
            return;
        }
        mount.innerHTML = `
            <table class="data-table">
                <thead><tr><th>Timestamp</th><th>User ID</th><th>Action</th><th>Entity</th><th>Description</th></tr></thead>
                <tbody>
                    ${result.content.map((log) => `
                        <tr>
                            <td>${utils.formatDateTime(log.timestamp)}</td>
                            <td>${log.userId ?? '—'}</td>
                            <td><span class="badge badge-neutral">${log.action}</span></td>
                            <td>${log.entityType || '—'} ${log.entityId ? '#' + log.entityId : ''}</td>
                            <td>${log.description || '—'}</td>
                        </tr>`).join('')}
                </tbody>
            </table>`;
    } catch (err) {
        mount.innerHTML = `<div class="error-state"><h3>Could not load audit log</h3><p>${err.message}</p></div>`;
    }
}

document.addEventListener('DOMContentLoaded', loadAdminPage);
