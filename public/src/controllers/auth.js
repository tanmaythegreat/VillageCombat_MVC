
startAuthRefreshTimer();

import { UserData } from '../models/map.js';

export let access_token = localStorage.getItem('access_token');

export async function refreshAuthToken() {
    const refreshToken = localStorage.getItem('refresh_token_b64');
    try {
        const response = await fetch('/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user_id: UserData.user_id, refresh_token: refreshToken }),
        });
        if (!response.ok) throw new Error('Session expired');
        const data = await response.json();
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token_b64', data.refresh_token_b64);
        console.log('Token Refreshed');
        access_token = data.access_token;
    } catch (error) {
        console.error('Refresh failed, redirecting to login:', error);
        window.location.href = '/Login.html';
    }
}

export function startAuthRefreshTimer(intervalMs = 14 * 60 * 1000) {
    setInterval(refreshAuthToken, intervalMs);
}