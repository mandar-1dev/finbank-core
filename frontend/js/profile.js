/**
 * CoreBank — profile.js
 * Read-only profile summary sourced from local session data plus a
 * lightweight account overview. There is no dedicated GET /api/customers/me
 * endpoint in this iteration, so the page composes what's available from
 * /api/accounts and the logged-in user's session info.
 */

async function loadProfilePage() {
    auth.requireAuth();
    utils.renderShell('profile.html');
    document.getElementById('page-title').textContent = 'Profile';

    const user = auth.getUser();
    const content = document.getElementById('page-content');
    content.innerHTML = `
        <div class="page-header"><div><h1>Profile</h1><p>Your CoreBank account summary.</p></div></div>

        <div class="grid-2" style="align-items:start;">
            <div class="card">
                <div style="display:flex;align-items:center;gap:16px;margin-bottom:20px;">
                    <div class="avatar" style="width:56px;height:56px;font-size:18px;">${utils.initials(user.fullName)}</div>
                    <div>
                        <h3 style="margin-bottom:2px;">${user.fullName}</h3>
                        <span class="badge badge-neutral">${user.role}</span>
                    </div>
                </div>
                <div class="txn-detail-row"><span>Customer ID</span><span>${user.id}</span></div>
                <div class="txn-detail-row"><span>Role</span><span>${user.role}</span></div>
            </div>

            <div class="card">
                <h3>Linked Accounts</h3>
                <div id="profile-accounts"><div class="skeleton" style="height:60px;border-radius:8px;"></div></div>
            </div>
        </div>
    `;

    try {
        const accounts = await api.get('/accounts');
        const mount = document.getElementById('profile-accounts');
        mount.innerHTML = accounts.length
            ? accounts.map((a) => `
                <div class="flex-between" style="padding:10px 0;border-bottom:1px solid var(--border);">
                    <span class="mono">${a.maskedAccountNumber}</span>
                    <span class="badge ${utils.statusBadgeClass(a.status)}">${a.status}</span>
                </div>`).join('')
            : '<p class="text-muted">No accounts linked.</p>';
    } catch (err) {
        notify.error(err.message || 'Could not load account summary');
    }
}

document.addEventListener('DOMContentLoaded', loadProfilePage);
