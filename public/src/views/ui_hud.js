import { BuildingCategory } from '../core/enums.js';
import { formatNum, formatTime, escapeHTML, setAffordability } from '../models/utils.js';
import { AllBuildingData, AllTroopsData, TrainedTroopsData, UserData, ConstructionTasks } from '../models/map.js';
import { CreateBuilding, UpgradeBuilding, RepairBuilding, TrainTroop } from '../controllers/network.js';
import {selectToMove} from "../core/move.js";

export function UpdateResourceUI() {
    document.getElementById('hud-gold-val').textContent        = formatNum(UserData.current_gold        ?? 0);
    document.getElementById('hud-elixir-val').textContent      = formatNum(UserData.current_elixir      ?? 0);
    document.getElementById('hud-dark-elixir-val').textContent = formatNum(UserData.current_dark_elixir ?? 0);
    document.getElementById('hud-gems-val').textContent        = formatNum(UserData.current_gems        ?? 0);
}

// region Affordablity
let _troopCount = 1;

export function canAfford(level, checkGems = false, placed_building_id = null, is_broken = false) {
    const notBusy     = !(placed_building_id != null && ConstructionTasks.some(t => t.placed_building_id === placed_building_id));
    const thOk        = is_broken ? true : (UserData.town_hall_level >= level.update_townhall_level_required);
    if (checkGems) {
        return notBusy && thOk && (UserData.current_gems ?? 0) >= (is_broken ? 1 + level.update_or_gem_required / 10 : level.update_or_gem_required);
    }
    const div = is_broken ? 10 : 1;
    return notBusy && thOk
        && (UserData.current_gold        ?? 0) >= (level.update_gold_required         ?? 0) / div
        && (UserData.current_elixir      ?? 0) >= (level.update_elxir_required        ?? 0) / div
        && (UserData.current_dark_elixir ?? 0) >= (level.update_dark_elixir_required  ?? 0) / div;
}

export function canAffordTroop(upgCost, checkGems = false, placed_building_id) {
    if (!upgCost) return false;
    const notBusy = !(placed_building_id != null && ConstructionTasks.some(t => t.placed_building_id === placed_building_id));
    const thOk    = UserData.town_hall_level >= upgCost.town_hall_level_required;
    const n       = _troopCount;
    if (checkGems) return (UserData.current_gems ?? 0) >= (upgCost.or_gem_required ?? 0) * n;
    return notBusy && thOk
        && (UserData.current_gold        ?? 0) >= (upgCost.gold_required        ?? 0) * n
        && (UserData.current_elixir      ?? 0) >= (upgCost.elixir_required      ?? 0) * n
        && (UserData.current_dark_elixir ?? 0) >= (upgCost.dark_elixir_required ?? 0) * n;
}
// endregion

// region Collect button
let _placed_building_id = '';
export let _placed_building = null;

const collect = document.getElementById('bm-collect-btn');

export function updateCollectButton() {
    let data, building, levelDetails;
    try {
        data         = _placed_building;
        building     = AllBuildingData[data.building_id];
        levelDetails = building.levels[data.level];
    } catch { return; }

    if (building.category === BuildingCategory.Resource && building.levels[1].generation_rate > 0 && !data.is_broken) {
        collect.style.display = 'block';
        const elapsed  = (Date.now() - new Date(_placed_building.last_updated_at).getTime()) / 3600000;
        const toCollect = Math.min(levelDetails.storage_capacity, levelDetails.generation_rate * elapsed);

        const capacityKey   = { gold: 'total_gold_capacity', elixir: 'total_elixir_capacity', dark_elixir: 'total_dark_elixir_capacity' }[building.resource_type];
        const currentKey    = { gold: 'current_gold',        elixir: 'current_elixir',        dark_elixir: 'current_dark_elixir'        }[building.resource_type];
        const cap  = UserData[capacityKey];
        const curr = UserData[currentKey];

        if (cap === curr) {
            collect.textContent = 'Storage Full';
            setAffordability(collect, false);
        } else {
            collect.textContent = `✨️ Collect ${formatNum(Math.min(toCollect, cap - curr))}`;
            setAffordability(collect, true);
        }
    } else {
        collect.style.display = 'none';
    }
}

collect.onclick = (e) => {
    e.stopPropagation();
    import('../controllers/network.js').then(m =>
        m.SendToServer({ action: 'COLLECT_RESOURCE', message: JSON.stringify({ placed_building_id: _placed_building_id }) })
    );
    _placed_building.last_updated_at = Date.now();
};
// endregion

// region Building Menu
const trainBtn = document.getElementById('bm-train-btn');

function updateTrainButton(building) {
    trainBtn.style.display = (building.category === BuildingCategory.Army && !_placed_building.is_broken) ? 'block' : 'none';
}

document.getElementById('bm-train-btn').onclick = (e) => {
    e.stopPropagation();
    document.getElementById('bm-overlay').classList.remove('is-active');
    openTroopTraining(_placed_building_id);
};

export function triggerBuildingMenu(data) {
    _placed_building_id = data.id;
    _placed_building    = data;

    const building     = AllBuildingData[data.building_id];
    const levelDetails = building.levels[data.level];

    updateTrainButton(building);
    updateCollectButton(building, data, levelDetails);

    const maxLevel  = 6;
    const nextLevel = data.level + 1;
    const isMaxLevel = data.level >= maxLevel;

    const categoryIcon = {
        Defense:  'ti-shield', Resource: 'ti-database',
        Army:     'ti-sword',  TownHall: 'ti-building-castle',
    }[building.category] ?? 'ti-home';

    document.getElementById('bm-icon-class').className   = `ti ${categoryIcon}`;
    document.getElementById('bm-name-text').textContent  = building.name;
    document.getElementById('bm-sub-text').textContent   = `${building.category} · Level ${data.level}`;
    document.getElementById('bm-level-dots').innerHTML   = Array.from({ length: maxLevel }, (_, i) =>
        `<span class="bm-dot${i < data.level ? ' filled' : ''}"></span>`
    ).join('');
    document.getElementById('bm-level-text').textContent = `${data.level} / ${maxLevel}`;
    document.getElementById('bm-stats-grid').innerHTML   = getBuildingStats(building, levelDetails).map(([key, val]) =>
        `<div class="bm-stat"><p class="bm-stat-val">${escapeHTML(val)}</p><p class="bm-stat-key">${escapeHTML(key)}</p></div>`
    ).join('');

    const upgradeSection  = document.getElementById('bm-upgrade-section');
    const maxNotice       = document.getElementById('bm-max-notice');
    const upgradeBtn      = document.getElementById('bm-upgrade-btn');
    const gemUpgradeBtn   = document.getElementById('bm-gem-upgrade-btn');
    const gemRepairBtn    = document.getElementById('bm-gem-repair-btn');
    const repairBtn       = document.getElementById('bm-repair-btn');
    const movebtn = document.getElementById('bm-move-btn')
    movebtn.style.display = 'block'
    movebtn.addEventListener('click',(e)=>{
        e.stopPropagation()
        overlay.classList.remove('is-active');
        selectToMove()
    })
    gemRepairBtn.style.display  = data.is_broken ? 'block' : 'none';
    repairBtn.style.display     = data.is_broken ? 'block' : 'none';
    gemUpgradeBtn.style.display = data.is_broken ? 'none'  : 'block';
    upgradeBtn.style.display    = data.is_broken ? 'none'  : 'block';

    if (isMaxLevel && !data.is_broken) {
        upgradeSection.style.display = 'none';
        maxNotice.style.display      = 'block';
        upgradeBtn.disabled           = true;
        if (gemUpgradeBtn) gemUpgradeBtn.disabled = true;
    } else {
        upgradeSection.style.display = 'block';
        maxNotice.style.display      = 'none';

        document.getElementById('bm-upgrade-title-text').textContent =
            data.is_broken ? `Repair to level ${nextLevel - 1}` : `Upgrade to level ${nextLevel}`;
        document.getElementById('bm-upgrade-title').textContent =
            data.is_broken ? 'Repair cost' : 'Upgrade cost';

        const thTag = document.getElementById('bm-th-req-tag');
        if (levelDetails.update_townhall_level_required && !data.is_broken) {
            thTag.textContent   = `TH ${levelDetails.update_townhall_level_required} required`;
            thTag.style.display = 'inline-block';
        } else { thTag.style.display = 'none'; }

        document.getElementById('bm-costs-container').innerHTML = buildCostRows(levelDetails, data.is_broken);

        setAffordability(data.is_broken ? repairBtn    : upgradeBtn,    canAfford(levelDetails, false, data.id, data.is_broken));
        setAffordability(data.is_broken ? gemRepairBtn : gemUpgradeBtn, canAfford(levelDetails, true,  data.id, data.is_broken));

        if (gemUpgradeBtn) {
            let gemCost = levelDetails.update_or_gem_required ?? 0;
            if (data.is_broken) gemCost = 1 + gemCost / 10;
            (data.is_broken ? gemRepairBtn : gemUpgradeBtn).textContent = `💎 Instant (${formatNum(gemCost)})`;
        }
    }

    const overlay = document.getElementById('bm-overlay');
    overlay.classList.add('is-active');
    document.getElementById('bm-close-btn').onclick = (e) => { e.stopPropagation(); overlay.classList.remove('is-active'); };
    overlay.onclick = (e) => { e.stopPropagation(); if (e.target === overlay) overlay.classList.remove('is-active'); };

    upgradeBtn.onclick = (e) => {
        e.stopPropagation();
        if (isMaxLevel) return;
        overlay.classList.remove('is-active');
        UpgradeBuilding(data.id, false);
        UserData.current_gold        -= levelDetails.update_gold_required;
        UserData.current_elixir      -= levelDetails.update_elxir_required;
        UserData.current_dark_elixir -= levelDetails.update_dark_elixir_required;
        UpdateResourceUI();
    };
    if (gemUpgradeBtn) {
        gemUpgradeBtn.onclick = (e) => {
            e.stopPropagation();
            if (isMaxLevel) return;
            overlay.classList.remove('is-active');
            UpgradeBuilding(data.id, true);
            UserData.current_gems -= levelDetails.update_or_gem_required;
            UpdateResourceUI();
        };
    }
    repairBtn.onclick = (e) => {
        e.stopPropagation();
        overlay.classList.remove('is-active');
        RepairBuilding(data.id, false);
        UserData.current_gold        -= levelDetails.update_gold_required        / 10;
        UserData.current_elixir      -= levelDetails.update_elxir_required       / 10;
        UserData.current_dark_elixir -= levelDetails.update_dark_elixir_required / 10;
        UpdateResourceUI();
    };
    if (gemUpgradeBtn) {
        gemRepairBtn.onclick = (e) => {
            e.stopPropagation();
            if (isMaxLevel) return;
            overlay.classList.remove('is-active');
            RepairBuilding(data.id, true);
            UserData.current_gems -= 1 + levelDetails.update_or_gem_required / 10;
            UpdateResourceUI();
        };
    }
}

function getBuildingStats(building, level) {
    const base = [['HP', (level.health ?? 0).toLocaleString()]];
    switch (building.category) {
        case BuildingCategory.Defense:
            return [...base,
                ['Damage / shot',  level.damage_per_shot  ?? '—'],
                ['Range',          building.attack_range  ? `${building.attack_range} tiles` : '—'],
                ['Attack speed',   building.attack_speed_seconds ? `${building.attack_speed_seconds}s` : '—'],
                ['Damage type',    building.damage_type   ?? '—'],
                ['Targets',        building.unit_target   ?? '—'],
            ];
        case BuildingCategory.Resource:
            return [...base,
                ['Gen rate / hr',  level.generation_rate  ? formatNum(level.generation_rate)  : '—'],
                ['Capacity',       level.storage_capacity ? formatNum(level.storage_capacity) : '—'],
                ['Resource',       building.resource_type ?? '—'],
            ];
        case BuildingCategory.Army:
            return [...base, ['Troop capacity', level.troop_capacity ?? '—']];
        default:
            return base;
    }
}

function buildCostRows(level, is_broken = false) {
    const d = is_broken ? 10 : 1;
    const costs = [
        { label: 'Gold',        icon: '🪙', val: (level.update_gold_required        ?? 0) / d },
        { label: 'Elixir',      icon: '🧪', val: (level.update_elxir_required       ?? 0) / d },
        { label: 'Dark elixir', icon: '⚗️',  val: (level.update_dark_elixir_required ?? 0) / d },
    ].filter(c => c.val > 0);

    return costs.map(c =>
            `<div class="bm-cost-row"><span class="bm-cost-label">${c.icon} ${c.label}</span><span class="bm-cost-val">${formatNum(c.val)}</span></div>`
        ).join('') +
        `<div class="bm-divider"></div><div class="bm-time-row">⏱ ${formatTime(level.update_time_required_required / 10)} build time</div>`;
}

// ── Building Shop ──────────────────────────────────────────────────────────
let _shopGridX = 0, _shopGridY = 0, _shopSelectedId = null;
let _shopActiveFilter = 'All', _shopSearchQuery = '';

export function openBuildingShop(gridX, gridY) {
    _shopGridX = gridX; _shopGridY = gridY; _shopSelectedId = null;
    document.getElementById('shop-grid-sub').textContent = `Grid (${gridX}, ${gridY}) · choose a building`;
    showShopList();
    document.getElementById('shop-overlay').classList.add('is-active');
}

function closeShop() { document.getElementById('shop-overlay').classList.remove('is-active'); }

document.getElementById('shop-close-btn').onclick = closeShop;
document.getElementById('shop-overlay').onclick = (e) => {
    e.stopPropagation();
    if (e.target === document.getElementById('shop-overlay')) closeShop();
};

function showShopList() {
    document.getElementById('shop-list').style.display   = 'flex';
    document.getElementById('shop-detail').style.display = 'none';
    renderShopCards();
}

function renderShopCards() {
    const list  = document.getElementById('shop-list');
    const query = _shopSearchQuery.toLowerCase();
    const entries = Object.entries(AllBuildingData).filter(([, b]) =>
        (_shopActiveFilter === 'All' || b.category === _shopActiveFilter) &&
        (!query || b.name.toLowerCase().includes(query))
    );

    if (!entries.length) {
        list.innerHTML = `<p style="text-align:center;color:#aaa;padding:24px 0;font-size:13px;">No buildings found</p>`;
        return;
    }

    list.innerHTML = entries.map(([building_id, building]) => {
        const level1  = building.levels[1];
        const costs   = level1 ? getConstructionCostPills(level1) : [];
        const imgSrc  = `./Models/${building.name}.png`;
        const catIcon = { defense: '🛡️', resource: '💰', army: '⚔️', townhall: '🏰' }[building.category] ?? '🏠';

        return `<div class="shop-card" data-id="${escapeHTML(building_id)}">
            <img class="shop-card-img" src="${escapeHTML(imgSrc)}" alt="${escapeHTML(building.name)}"
                 onerror="this.style.display='none';this.nextElementSibling.style.display='flex';" />
            <div class="shop-card-img-fallback" style="display:none;">${catIcon}</div>
            <div class="shop-card-info">
                <p class="shop-card-name">${escapeHTML(building.name)}</p>
                <p class="shop-card-cat">${building.category}</p>
                <div class="shop-card-costs">${costs.map(c => `<span class="shop-card-cost-pill">${c}</span>`).join('')}</div>
            </div>
            <span class="shop-card-arrow">›</span>
        </div>`;
    }).join('');

    list.querySelectorAll('.shop-card').forEach(card =>
        card.addEventListener('click', (e) => { e.stopPropagation(); showShopDetail(card.dataset.id); })
    );
}

function getConstructionCostPills(level) {
    const pills = [];
    if (level.update_gold_required        > 0) pills.push(`🪙 ${formatNum(level.update_gold_required)}`);
    if (level.update_elxir_required       > 0) pills.push(`🧪 ${formatNum(level.update_elxir_required)}`);
    if (level.update_dark_elixir_required > 0) pills.push(`⚗️ ${formatNum(level.update_dark_elixir_required)}`);
    if (!pills.length && level.update_or_gem_required > 0) pills.push(`💎 ${formatNum(level.update_or_gem_required)}`);
    return pills;
}

function showShopDetail(building_id) {
    _shopSelectedId = building_id;
    const building  = AllBuildingData[building_id];
    const level1    = building.levels[1];

    document.getElementById('shop-detail-name').textContent = building.name;
    document.getElementById('shop-detail-cat').textContent  = building.category;
    document.getElementById('shop-detail-img').src = `./Models/${building.name}.png`;
    document.getElementById('shop-detail-img').alt = building.name;
    document.getElementById('shop-detail-stats').innerHTML = getBuildingStats(building, level1).map(([key, val]) =>
        `<div class="bm-stat"><p class="bm-stat-val">${escapeHTML(val)}</p><p class="bm-stat-key">${escapeHTML(key)}</p></div>`
    ).join('');
    document.getElementById('shop-detail-costs').innerHTML = buildCostRows(level1);

    const thTag = document.getElementById('shop-th-req');
    if (level1.update_townhall_level_required) {
        thTag.textContent = `TH ${level1.update_townhall_level_required} required`; thTag.style.display = 'inline-block';
    } else { thTag.style.display = 'none'; }

    document.getElementById('shop-gem-btn').textContent = `💎 Instant (${formatNum(level1.update_or_gem_required ?? 0)})`;
    setAffordability(document.getElementById('shop-build-btn'), canAfford(level1, false));
    setAffordability(document.getElementById('shop-gem-btn'),   canAfford(level1, true));

    document.getElementById('shop-list').style.display   = 'none';
    document.getElementById('shop-detail').style.display = 'flex';
}

document.getElementById('shop-detail-back').onclick = showShopList;
document.getElementById('shop-filter-tabs').addEventListener('click', (e) => {
    const tab = e.target.closest('.shop-tab');
    if (!tab) return;
    document.querySelectorAll('.shop-tab').forEach(t => t.classList.remove('active'));
    tab.classList.add('active');
    _shopActiveFilter = tab.dataset.cat;
    renderShopCards();
});
document.getElementById('shop-search').addEventListener('input', (e) => {
    _shopSearchQuery = e.target.value; renderShopCards();
});

document.getElementById('shop-build-btn').onclick = (e) => {
    e.stopPropagation();
    if (!_shopSelectedId) return;
    closeShop();
    CreateBuilding(_shopSelectedId, _shopGridX, _shopGridY, false);
    const ld = AllBuildingData[_shopSelectedId].levels[1];
    UserData.current_gold        -= ld.update_gold_required;
    UserData.current_elixir      -= ld.update_elxir_required;
    UserData.current_dark_elixir -= ld.update_dark_elixir_required;
    UpdateResourceUI();
};
document.getElementById('shop-gem-btn').onclick = (e) => {
    e.stopPropagation();
    if (!_shopSelectedId) return;
    closeShop();
    CreateBuilding(_shopSelectedId, _shopGridX, _shopGridY, true);
    UserData.current_gems -= AllBuildingData[_shopSelectedId].levels[1].update_or_gem_required;
    UpdateResourceUI();
};
// endregion

// region Troop Training
let _troopSelectedId    = null;
let _troopSelectedLevel = 1;

export function openTroopTraining(placed_building_id) {
    _troopSelectedId    = null;
    _troopSelectedLevel = 1;
    _troopCount         = 1;
    showTroopGrid(placed_building_id);
    document.getElementById('troop-overlay').classList.add('is-active');
}

function closeTroopOverlay() { document.getElementById('troop-overlay').classList.remove('is-active'); }

document.getElementById('troop-close-btn').onclick        = (e) => { e.stopPropagation(); closeTroopOverlay(); };
document.getElementById('troop-detail-close-btn').onclick = (e) => { e.stopPropagation(); closeTroopOverlay(); };
document.getElementById('troop-overlay').onclick = (e) => {
    e.stopPropagation();
    if (e.target === document.getElementById('troop-overlay')) closeTroopOverlay();
};
document.getElementById('troop-detail-back').onclick = (e) => { e.stopPropagation(); showTroopGrid(); };

function showTroopGrid(placed_building_id) {
    document.getElementById('troop-grid-view').style.display   = 'block';
    document.getElementById('troop-detail-view').style.display = 'none';
    renderTroopGrid(placed_building_id);
}

function renderTroopGrid(placed_building_id) {
    const grid = document.getElementById('troop-grid');
    const tpl  = document.getElementById('troop-card-tpl');
    grid.innerHTML = '';

    Object.entries(AllTroopsData).forEach(([troopId, troop]) => {
        const realLevels = troop.level_stats.length - 1;
        let totalOwned = 0;
        for (let lv = 1; lv <= realLevels; lv++) totalOwned += TrainedTroopsData[[troopId, lv]] ?? 0;

        const node  = tpl.content.cloneNode(true);
        const card  = node.querySelector('.troop-card');
        const img   = node.querySelector('.troop-card-img');
        const badge = node.querySelector('.troop-card-badge');

        card.dataset.id      = troopId;
        img.src              = `./Models/${escapeHTML(troop.name)}.png`;
        img.alt              = troop.name;
        img.onerror          = function () { this.style.display = 'none'; this.nextElementSibling.style.display = 'flex'; };
        node.querySelector('.troop-card-name').textContent = troop.name;
        node.querySelector('.troop-card-sub').textContent  = `Lv 1–${realLevels}`;

        if (totalOwned > 0) { badge.textContent = totalOwned; badge.style.display = 'flex'; }
        card.addEventListener('click', (e) => { e.stopPropagation(); showTroopDetail(troopId, placed_building_id); });
        grid.appendChild(node);
    });
}

function showTroopDetail(troopId, placed_building_id) {
    _troopSelectedId    = troopId;
    _troopSelectedLevel = 1;
    _troopCount         = 1;
    document.getElementById('troop-grid-view').style.display   = 'none';
    document.getElementById('troop-detail-view').style.display = 'block';
    renderTroopSidebar(troopId, placed_building_id);
    renderTroopDetailPane(troopId, placed_building_id);
}

function renderTroopSidebar(activeTroopId, placed_building_id) {
    const sidebar = document.getElementById('troop-sidebar');
    const tpl     = document.getElementById('troop-sidebar-tpl');
    sidebar.innerHTML = '';

    Object.entries(AllTroopsData).forEach(([troopId, troop]) => {
        const node = tpl.content.cloneNode(true);
        const item = node.querySelector('.troop-sidebar-item');
        const img  = node.querySelector('.troop-sidebar-img');
        const icon = node.querySelector('.troop-sidebar-icon');

        item.dataset.id = troopId;
        if (troopId === activeTroopId) item.classList.add('active');
        img.src   = `./Models/${escapeHTML(troop.name)}.png`;
        img.alt   = troop.name;
        img.onerror = function () { this.style.display = 'none'; icon.style.display = 'flex'; };
        node.querySelector('.troop-sidebar-name').textContent = troop.name;

        item.addEventListener('click', (e) => {
            e.stopPropagation();
            _troopSelectedLevel = 1; _troopCount = 1; _troopSelectedId = troopId;
            sidebar.querySelectorAll('.troop-sidebar-item').forEach(i => i.classList.remove('active'));
            item.classList.add('active');
            renderTroopDetailPane(troopId, placed_building_id);
        });
        sidebar.appendChild(node);
    });
}

function renderTroopDetailPane(troopId, placed_building_id) {
    const troop      = AllTroopsData[troopId];
    const realLevels = troop.level_stats.length - 1;

    document.getElementById('troop-detail-name').textContent = troop.name;
    document.getElementById('troop-pane-name').textContent   = troop.name;
    document.getElementById('troop-pane-meta').textContent   =
        `${troop.attack_type} · ${troop.preferred_target} target · ${troop.housing_space} housing`;

    const tabsEl = document.getElementById('troop-level-tabs');
    const tabTpl = document.getElementById('troop-level-tab-tpl');
    tabsEl.innerHTML = '';

    for (let lv = 1; lv <= realLevels; lv++) {
        const owned = TrainedTroopsData[[troopId, lv]] ?? 0;
        const node  = tabTpl.content.cloneNode(true);
        const tab   = node.querySelector('.troop-level-tab');
        const cntEl = node.querySelector('.troop-tab-count');

        tab.dataset.lv = lv;
        node.querySelector('.troop-tab-lv').textContent = `Lv ${lv}`;
        if (owned > 0) { cntEl.textContent = `×${owned}`; cntEl.style.display = 'block'; }
        if (lv === _troopSelectedLevel) tab.classList.add('active');

        tab.addEventListener('click', (e) => {
            e.stopPropagation();
            _troopSelectedLevel = lv; _troopCount = 1;
            tabsEl.querySelectorAll('.troop-level-tab').forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            refreshTroopDetailPane(troopId, placed_building_id);
        });
        tabsEl.appendChild(node);
    }
    refreshTroopDetailPane(troopId, placed_building_id);
}

function refreshTroopDetailPane(troopId, placed_building_id) {
    const troop   = AllTroopsData[troopId];
    const lv      = _troopSelectedLevel;
    const stats   = troop.level_stats[lv];
    const upgCost = troop.upgrade_costs[lv - 1];
    const isNewTrain = lv === 1;

    document.getElementById('troop-owned-val').textContent = TrainedTroopsData[[troopId, lv]] ?? 0;

    const statDefs = [
        ['HP',         (stats.health ?? 0).toLocaleString()],
        ['Dmg / shot', stats.damage_per_shot ?? '—'],
        ['Move speed', troop.movement_speed   ?? '—'],
        ['Atk speed',  troop.attack_speed_seconds ? `${troop.attack_speed_seconds}s` : '—'],
        ['Range',      troop.attack_range         ? `${troop.attack_range} tiles`    : '—'],
        ['Housing',    troop.housing_space         ?? '—'],
    ];
    const statsGrid = document.getElementById('troop-stats-grid');
    if (statsGrid.children.length !== statDefs.length) {
        statsGrid.innerHTML = statDefs.map(([key, val]) =>
            `<div class="bm-stat"><p class="bm-stat-val">${escapeHTML(String(val))}</p><p class="bm-stat-key">${escapeHTML(key)}</p></div>`
        ).join('');
    } else {
        statsGrid.querySelectorAll('.bm-stat').forEach((el, i) =>
            el.querySelector('.bm-stat-val').textContent = String(statDefs[i][1])
        );
    }

    const thTag  = document.getElementById('troop-th-req');
    if (upgCost?.town_hall_level_required) {
        thTag.textContent = `TH ${upgCost.town_hall_level_required} required`; thTag.style.display = 'inline-block';
    } else { thTag.style.display = 'none'; }

    const noteEl = document.getElementById('troop-requires-note');
    if (!isNewTrain) {
        const prevOwned = TrainedTroopsData[[troopId, lv - 1]] ?? 0;
        noteEl.textContent   = `⚠ Requires ${_troopCount}× Lv ${lv - 1} — you have ${prevOwned}`;
        noteEl.style.display = 'block';
    } else { noteEl.style.display = 'none'; }

    const trainBtn = document.getElementById('troop-train-btn');
    const gemBtn   = document.getElementById('troop-gem-btn');

    document.getElementById('troop-count-dec').onclick = (e) => {
        e.stopPropagation();
        if (_troopCount > 1) { _troopCount--; updateTroopCosts(troopId, placed_building_id); }
    };
    document.getElementById('troop-count-inc').onclick = (e) => {
        e.stopPropagation(); _troopCount++; updateTroopCosts(troopId, placed_building_id);
    };

    trainBtn.onclick = (e) => {
        e.stopPropagation();
        if (!canAffordTroop(upgCost, false)) return;
        closeTroopOverlay();
        TrainTroop(troopId, _troopCount, placed_building_id, lv, false);
        UserData.current_gold        -= upgCost.gold_required;
        UserData.current_elixir      -= upgCost.elxir_required;
        UserData.current_dark_elixir -= upgCost.dark_elixir_required;
        if (lv !== 1) TrainedTroopsData[[troop, lv-1]] -= _troopCount;
    };
    gemBtn.onclick = (e) => {
        console.log(_troopCount,lv)
        e.stopPropagation();
        if (!canAffordTroop(upgCost, true)) return;
        closeTroopOverlay();
        TrainTroop(troopId, _troopCount, placed_building_id, lv, true);
        UserData.current_gems -= upgCost.or_gem_required;
        if (lv !== 1) TrainedTroopsData[[troop, lv-1]] -= _troopCount;
    };

    updateTroopCosts(troopId, placed_building_id);
}

function updateTroopCosts(troopId, placed_building_id) {
    const troop      = AllTroopsData[troopId];
    const lv         = _troopSelectedLevel;
    const upgCost    = troop.upgrade_costs[lv - 1];
    const n          = _troopCount;
    const isNewTrain = lv === 1;

    document.getElementById('troop-count-val').textContent  = n;
    document.getElementById('troop-owned-val').textContent  = TrainedTroopsData[[troopId, _troopSelectedLevel]] ?? 0;

    const noteEl = document.getElementById('troop-requires-note');
    if (!isNewTrain) {
        const prevOwned = TrainedTroopsData[[troopId, lv - 1]] ?? 0;
        noteEl.textContent = `⚠ Requires ${n}× Lv ${lv - 1} — you have ${prevOwned}`;
    }

    const scaled = [
        { label: 'Gold',        icon: '🪙', val: (upgCost?.gold_required         ?? 0) * n },
        { label: 'Elixir',      icon: '🧪', val: (upgCost?.elixir_required       ?? 0) * n },
        { label: 'Dark elixir', icon: '⚗️',  val: (upgCost?.dark_elixir_required  ?? 0) * n },
    ].filter(c => c.val > 0);

    const timeSec = (upgCost?.time_required_seconds ?? 0) * n;
    document.getElementById('troop-costs').innerHTML =
        scaled.map(c => `<div class="bm-cost-row"><span class="bm-cost-label">${c.icon} ${c.label}</span><span class="bm-cost-val">${formatNum(c.val)}</span></div>`).join('') +
        `<div class="bm-divider"></div><div class="bm-time-row">⏱ ${formatTime(timeSec)} total time</div>`;

    const trainBtn = document.getElementById('troop-train-btn');
    const gemBtn   = document.getElementById('troop-gem-btn');
    trainBtn.textContent = isNewTrain ? `Train ${n}× ${troop.name} (Lv 1)` : `Upgrade ${n}× to Lv ${lv}`;
    gemBtn.textContent   = `💎 Instant (${formatNum((upgCost?.or_gem_required ?? 0) * n)})`;

    const prevOwned    = isNewTrain ? Infinity : (TrainedTroopsData[[troopId, lv - 1]] ?? 0);
    const enoughTroops = isNewTrain || n <= prevOwned;
    setAffordability(trainBtn, canAffordTroop(upgCost, false, placed_building_id) && enoughTroops);
    setAffordability(gemBtn,   canAffordTroop(upgCost, true,  placed_building_id) && enoughTroops);
}
// endregion
