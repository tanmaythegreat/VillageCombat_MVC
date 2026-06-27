import { access_token } from './auth.js';
import { inBattle } from './battle.js';
import {handlePeacetimeMessage} from "./peacetime.js";
import {handleBattleMessage} from "./battle_network.js";


let socket;

export function SendToServer(dataObject) {
    dataObject.access_token = access_token;
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(dataObject));
    } else {
        console.error('Cannot send action. WebSocket connection is not open.');
    }
}

export function connectToGameServer() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const socketUrl = `${protocol}//${window.location.host}/ws?token=${access_token}`;
    socket = new WebSocket(socketUrl);

    socket.addEventListener('open', () => SendToServer({ action: 'INITIAL_LOAD', message: '' }));
    socket.addEventListener('error', (err) => console.error('WebSocket Error:', err));
    socket.addEventListener('close', () => {
        console.warn('Disconnected. Reconnecting…');
        connectToGameServer();
    });
    socket.addEventListener('message', (event) => {
        const data = JSON.parse(event.data);
        handlePeacetimeMessage(data);
        if (inBattle) handleBattleMessage(data);
        if (data.status === 'error') {
            showToast(data.message);
            console.log(data);
        }
    });
}

export function CreateBuilding(building_id, x, y, use_gems = false) {
    SendToServer({ action: 'CREATE_BUILDING', message: JSON.stringify({ building_id, x, y, use_gems }) });
}

export function UpgradeBuilding(placed_building_id, use_gems = false) {
    SendToServer({ action: 'UPGRADE_BUILDING', message: JSON.stringify({ placed_building_id, use_gems }) });
}

export function RepairBuilding(placed_building_id, use_gems = false) {
    SendToServer({ action: 'REPAIR_BUILDING', message: JSON.stringify({ placed_building_id, use_gems }) });
}

export function TrainTroop(troop_id, count, barrack_placed_building_id, level_to, use_gems) {
    SendToServer({
        action: 'TRAIN_TROOP',
        message: JSON.stringify({ barrack_placed_building_id, troop_id, count, use_gems, level_from: level_to - 1 }),
    });
}

export function Revenge(opponentName) {
    SendToServer({ action: 'REVENGE', message: opponentName });
}

export function getBattleHistory(fought_at, to_load) {
    SendToServer({ action: 'BATTLE_HISTORY', message: JSON.stringify({ fought_at, to_load }) });
}

export function Logout() {
    localStorage.clear();
    SendToServer({ action: 'LOGOUT', message: '' });
    window.location.href = './Login.html';
}

export function CollectAll() {
    SendToServer({ action: 'COLLECT_ALL', message: '' });
}

export function showToast(message, type = 'error') {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;

    const icon = document.createElement('span');
    icon.className = 'toast-icon';
    icon.textContent = type === 'error' ? '⚠️' : '✅';

    const text = document.createElement('span');
    text.textContent = message;

    toast.appendChild(icon);
    toast.appendChild(text);
    container.appendChild(toast);

    requestAnimationFrame(() => toast.classList.add('is-active'));
    setTimeout(() => {
        toast.classList.remove('is-active');
        setTimeout(() => toast.remove(), 250);
    }, 4000);
}