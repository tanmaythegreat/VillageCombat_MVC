const API_BASE_URL = 'http://localhost:8080';

const loginForm = document.getElementById('login-form');
const registerForm = document.getElementById('register-form');
const statusMessage = document.getElementById('status-message');


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
        const response = await fetch(`${API_BASE_URL}/login`, {
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

    const payload = {
        username: username,
        email: email,
        password_text: passwordText
    };

    try {
        const response = await fetch(`${API_BASE_URL}/register`, {
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