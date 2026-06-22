const loginForm = document.getElementById('login-form');
const registerForm = document.getElementById('register-form');
const statusMessage = document.getElementById('status-message');

const regPasswordInput = document.getElementById('reg-password');
const pwStrengthWrap = document.getElementById('password-strength-wrap');
const pwStrengthFill = document.getElementById('pw-strength-fill');
const pwStrengthLabel = document.getElementById('pw-strength-label');

const PASSWORD_RULES = {
    length: { test: (pw) => pw.length >= 8, elId: 'req-length' },
    upper: { test: (pw) => /[A-Z]/.test(pw), elId: 'req-upper' },
    lower: { test: (pw) => /[a-z]/.test(pw), elId: 'req-lower' },
    number: { test: (pw) => /[0-9]/.test(pw), elId: 'req-number' },
    special: { test: (pw) => /[^A-Za-z0-9]/.test(pw), elId: 'req-special' }
};

function getPasswordChecks(password) {
    const checks = {};
    for (const key in PASSWORD_RULES) {
        checks[key] = PASSWORD_RULES[key].test(password);
    }
    return checks;
}

function isPasswordStrong(password) {
    const checks = getPasswordChecks(password);
    return Object.values(checks).every(Boolean);
}

function updatePasswordStrengthUI(password) {
    const checks = getPasswordChecks(password);

    Object.keys(PASSWORD_RULES).forEach((key) => {
        const li = document.getElementById(PASSWORD_RULES[key].elId);
        if (li) li.classList.toggle('met', checks[key]);
    });

    if (!password) {
        pwStrengthWrap.classList.add('hidden');
        return checks;
    }
    pwStrengthWrap.classList.remove('hidden');

    const score = Object.values(checks).filter(Boolean).length;
    const percent = (score / 5) * 100;
    pwStrengthFill.style.width = `${percent}%`;

    pwStrengthFill.classList.remove('weak', 'fair', 'good', 'strong');
    pwStrengthLabel.classList.remove('weak', 'fair', 'good', 'strong');

    let strengthClass = 'weak';
    let strengthText = 'Weak';
    if (score === 3) {
        strengthClass = 'fair';
        strengthText = 'Fair';
    } else if (score === 4) {
        strengthClass = 'good';
        strengthText = 'Good';
    } else if (score === 5) {
        strengthClass = 'strong';
        strengthText = 'Strong';
    }

    pwStrengthFill.classList.add(strengthClass);
    pwStrengthLabel.classList.add(strengthClass);
    pwStrengthLabel.textContent = strengthText;

    return checks;
}

regPasswordInput.addEventListener('input', (e) => {
    updatePasswordStrengthUI(e.target.value);
});


function switchTab(tabType) {
    hideMessage();

    if (tabType === 'login') {
        document.getElementById('tab-login').classList.add('active');
        document.getElementById('tab-register').classList.remove('active');
        loginForm.classList.remove('hidden');
        registerForm.classList.add('hidden');
    } else {
        document.getElementById('tab-register').classList.add('active');
        document.getElementById('tab-login').classList.remove('active');
        registerForm.classList.remove('hidden');
        loginForm.classList.add('hidden');
    }
}


function showMessage(text, type = 'error') {
    statusMessage.textContent = text;
    statusMessage.className = `status-msg ${type}`;
}

function hideMessage() {
    statusMessage.textContent = '';
    statusMessage.className = 'status-msg hidden';
}

loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    hideMessage();

    const identifier = document.getElementById('login-identifier').value.trim();
    const passwordText = document.getElementById('login-password').value;

    const payload = {
        password_text: passwordText
    };

    if (identifier.includes('@')) {
        payload.email = identifier;
        payload.username = "";
    } else {
        payload.username = identifier;
        payload.email = "";
    }

    try {
        const response = await fetch(`/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errorText = await response.text();

            if (errorText.includes("<!doctype html>")) {
                throw new Error("Target service path not found on deployment server (404).");
            }
            throw new Error(errorText || "Authorization parameters invalid.");
        }

        const data = await response.json();
        showMessage("Access Granted! Loading barracks data...", "success");

        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token_b64', data.refresh_token_b64);

        window.location.href = '/Game_village.html';

    } catch (err) {
        showMessage(err.message, 'error');
    }
});


registerForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    hideMessage();

    const username = document.getElementById('reg-username').value.trim();
    const email = document.getElementById('reg-email').value.trim();
    const passwordText = document.getElementById('reg-password').value;

    if (!username || !email || !passwordText) {
        showMessage("All fields are mandatory to secure a registration grant.", "error");
        return;
    }

    if (!isPasswordStrong(passwordText)) {
        showMessage("Password is too weak. It needs 8+ characters, an uppercase letter, a lowercase letter, a number, and a special character.", "error");
        return;
    }

    const payload = {
        username: username,
        email: email,
        password_text: passwordText
    };

    try {
        const response = await fetch(`/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errorText = await response.text();
            if (errorText.includes("<!doctype html>")) {
                throw new Error("Registration path missing from service router mapping config (404).");
            }
            throw new Error(errorText || "Failed to register user account profile.");
        }

        const data = await response.json();
        showMessage("User Successfully Registered! Access granted.", "success");

        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token_b64', data.refresh_token_b64);

        window.location.href = '/Game_village.html';
    } catch (err) {
        showMessage(err.message, 'error');
    }
});