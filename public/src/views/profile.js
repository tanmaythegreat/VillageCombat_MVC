import {UserData} from "../models/map.js";
import {formatNum, formatTime} from "../models/utils.js";
import {getBattleHistory, Revenge} from "../controllers/network.js";

let battleHistoryLoaded = 0;
let isLoadingMore = false;
let battleHistory = [];
let currentTheme = localStorage.getItem('gameTheme') || 'light';
const els = {};

function cacheElements() {
    els.profileBtn       = document.getElementById('profile-btn');
    els.overlay          = document.getElementById('profile-overlay');
    els.themeBtn         = document.getElementById('profile-theme-btn');
    els.closeBtn         = document.getElementById('profile-close-btn');
    els.copyBtn          = document.getElementById('profile-copy-btn');
    els.avatarInitials   = document.getElementById('profile-avatar-initials');
    els.usernameEl       = document.getElementById('profile-username');
    els.emailEl          = document.getElementById('profile-email');
    els.useridEl         = document.getElementById('profile-userid');
    els.attackInput      = document.getElementById('profile-attack-input');
    els.attackBtn         = document.getElementById('profile-attack-btn');
    els.battleList       = document.getElementById('profile-battle-list');
}

function applyTheme(theme) {
    const root = document.documentElement;
    if (theme === 'dark') {
        root.style.setProperty('--theme-bg-primary', '#1e1d1b');
        root.style.setProperty('--theme-bg-secondary', '#2C2C2A');
        root.style.setProperty('--theme-text-primary', '#fff');
        root.style.setProperty('--theme-text-secondary', '#B4B2A9');
        root.style.setProperty('--theme-border', '#444441');
    } else {
        root.style.setProperty('--theme-bg-primary', '#fff');
        root.style.setProperty('--theme-bg-secondary', '#f5f4f0');
        root.style.setProperty('--theme-text-primary', '#2C2C2A');
        root.style.setProperty('--theme-text-secondary', '#888');
        root.style.setProperty('--theme-border', '#ddd');
    }
    currentTheme = theme;
    localStorage.setItem('gameTheme', theme);
}

function renderUserInfo() {

    const initials = UserData.username
        ? UserData.username.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
        : '--';

    if (els.avatarInitials) els.avatarInitials.textContent = initials;
    if (els.usernameEl) els.usernameEl.textContent = UserData.username || '--';
    if (els.emailEl) els.emailEl.textContent = UserData.email || '--';
    if (els.useridEl) els.useridEl.textContent = UserData.user_id || '--';
}

function renderBattleHistory() {
    if (!els.battleList) return;
    if (battleHistory.length === 0) {
        els.battleList.innerHTML = '<p class="profile-no-battles">No battles yet</p>';
        return;
    }
    els.battleList.innerHTML = this.battleHistory.map(battle => {
        const isAttacker = battle.attacker_name === this.userData.username;
        const rawOpponent = isAttacker ? battle.defender_name : battle.attacker_name;
        const opponent = rawOpponent.replace(/</g, "&lt;").replace(/>/g, "&gt;");
        const isWin = isAttacker ? battle.winner_attacker : !battle.winner_attacker;

        return `
                <div class="profile-battle-card">
                    <div class="battle-card-header">
                        <div class="battle-info">
                            <p class="battle-opponent">vs ${opponent}</p>
                            <p class="battle-meta">${isAttacker ? 'Attacked' : 'Defended'} • ${this.formatTime(battle.fought_at)}</p>
                        </div>
                        <span class="battle-result ${isWin ? 'win' : 'loss'}">${isWin ? 'WIN' : 'LOSS'}</span>
                    </div>
                    <div class="battle-loot">
                        <span>🪙 ${this.formatNum(battle.gold_looted)}</span>
                        <span>🧪 ${this.formatNum(battle.elixir_looted)}</span>
                        <span>⚗️ ${this.formatNum(battle.dark_elixir_looted)}</span>
                    </div>
                    <button class="battle-action-btn" data-opponent="${opponent}">
                        ${isAttacker ? '⚔️ Attack again' : '⚔️ Revenge'}
                    </button>
                </div>
            `;
    }).join('');


    els.battleList.querySelectorAll('.battle-action-btn').forEach(btn => {
        btn.addEventListener('click', handleRevengeClick);
    });
}

function loadMoreBattles() {
    isLoadingMore = true;
    getBattleHistory(battleHistoryLoaded===0?new Date(Date.now() + 10 * 24 * 60 * 60 * 1000).toISOString():battleHistory[battleHistoryLoaded-1].fought_at, 10)
}

export function LoadedMoreBattles(battles){
    if (battles?.length > 0) {
        battleHistory.push(...battles);
        battleHistoryLoaded += battles.length;
        renderBattleHistory();
        isLoadingMore = false;
    }
}

function handleRevengeClick(e) {
    if (e) e.stopPropagation()
    closeProfile()
    Revenge(e.currentTarget.dataset.opponent);
}

function handleCustomAttack() {
    if (!els.attackInput) return;
    const username = els.attackInput.value.trim();
    if (username) {
        Revenge(username);
        closeProfile()
        els.attackInput.value = '';
    }
}

function handleInputKey(e) {
    if (e.key === 'Enter') handleCustomAttack();
}

function handleScroll() {
    if (!isLoadingMore && els.battleList.scrollTop + els.battleList.clientHeight >= els.battleList.scrollHeight - 100) {
        loadMoreBattles();
    }
}

function handleOverlayClick(e) {
    if (e.target.id === 'profile-overlay') closeProfile();
}

export function openProfile() {
    els.overlay?.classList.add('is-active');
    renderUserInfo();
    battleHistoryLoaded = 0;
    battleHistory = [];
    if (els.battleList) els.battleList.innerHTML = '';
    loadMoreBattles();
}

export function closeProfile() {
    els.overlay?.classList.remove('is-active');
}

export function toggleTheme() {
    const newTheme = currentTheme === 'light' ? 'dark' : 'light';
    applyTheme(newTheme);
    const icon = els.themeBtn?.querySelector('i');
    if (icon) icon.className = newTheme === 'dark' ? 'ti ti-sun' : 'ti ti-moon';
}
async function copyUserId() {
    await navigator.clipboard.writeText(UserData.user_id)
}
export function initProfile() {
    cacheElements();
    // applyTheme(currentTheme);
    els.copyBtn?.addEventListener('click' , (e)=>{e.stopPropagation();copyUserId()})
    els.profileBtn?.addEventListener('click', (e)=>{e.stopPropagation();openProfile(e)});
    els.closeBtn?.addEventListener('click', (e)=>{e.stopPropagation();closeProfile(e)});
    els.overlay?.addEventListener('click', (e)=>{e.stopPropagation();handleOverlayClick(e)});
    els.themeBtn?.addEventListener('click', (e)=>{e.stopPropagation();toggleTheme(e)});
    els.attackBtn?.addEventListener('click', (e)=>{e.stopPropagation();handleCustomAttack(e)});
    els.attackInput?.addEventListener('keydown', (e)=>{handleInputKey(e)});
    els.battleList?.addEventListener('scroll', (e)=>{handleScroll(e)});
}
