/**
 * CoreBank — auth.js
 * Handles the login form and simple client-side session storage. The JWT
 * itself is never trusted for authorization decisions — the backend
 * re-validates on every request. This only gates which pages render.
 */

const auth = {
    getToken() {
        return localStorage.getItem('corebank_token');
    },

    getUser() {
        const raw = localStorage.getItem('corebank_user');
        return raw ? JSON.parse(raw) : null;
    },

    isLoggedIn() {
        return !!this.getToken();
    },

    isAdmin() {
        const user = this.getUser();
        return user && user.role === 'ADMIN';
    },

    logout() {
        localStorage.removeItem('corebank_token');
        localStorage.removeItem('corebank_user');
        window.location.href = 'login.html';
    },

    /** Call at the top of every protected page. Redirects if not logged in. */
    requireAuth() {
        if (!this.isLoggedIn()) {
            window.location.href = 'login.html';
        }
    },

    requireAdmin() {
        this.requireAuth();
        if (!this.isAdmin()) {
            window.location.href = 'dashboard.html';
        }
    },
};

window.auth = auth;

function initLoginForm() {
    const form = utils.qs('#login-form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const submitBtn = utils.qs('button[type="submit"]', form);
        const email = utils.qs('#email', form).value.trim();
        const password = utils.qs('#password', form).value;

        submitBtn.disabled = true;
        submitBtn.innerHTML = '<span class="btn-spinner"></span> Signing in...';

        try {
            const response = await api.post('/auth/login', { email, password });
            localStorage.setItem('corebank_token', response.token);
            localStorage.setItem('corebank_user', JSON.stringify({
                id: response.customerId, fullName: response.fullName, role: response.role,
            }));
            notify.success(`Welcome back, ${response.fullName.split(' ')[0]}!`);
            window.location.href = response.role === 'ADMIN' ? 'admin.html' : 'dashboard.html';
        } catch (err) {
            notify.error(err.message || 'Invalid email or password');
            submitBtn.disabled = false;
            submitBtn.textContent = 'Log In';
        }
    });
}

document.addEventListener('DOMContentLoaded', initLoginForm);
