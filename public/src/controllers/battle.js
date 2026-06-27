import * as THREE from '../../THREE/three.module.js';
import {
    scene,
    canvas,
    camera,
    ground,
    raycaster,
    mouse,
    position_scaling,
    size_scaling,
    textureLoader
} from '../core/scene.js';
import { AllTroopsData, AllBuildingData, TrainedTroopsData, LoadMap } from '../models/map.js';
import {
    formatTime, fmt, makeImgWithFallback } from '../models/utils.js';
import {Revenge, SendToServer} from './network.js';

export let inBattle  = false;
export let replay    = false;
export let IsAttacker = false;
export let state     = undefined;

export function setInBattle(v) {
    inBattle = v;
    if (v) _startBattleTimer();
    else _stopBattleTimer();
}
export function setIsAttacker(v) { IsAttacker = v; }
export function setReplay(v)     { replay     = v; }
export function setState(v)      { state      = v; }

const attackBtn          = document.getElementById('attack-btn');
const incomingWarning    = document.getElementById('incoming-warning');
const matchmakingOverlay = document.getElementById('matchmaking-overlay');
const matchmakingTimer   = document.getElementById('matchmaking-timer');
const matchmakingCancel  = document.getElementById('matchmaking-cancel-btn');
const battleTimerHud     = document.getElementById('battle-timer-hud');
const battleTimerVal     = document.getElementById('battle-timer-val');

let _mmInterval = null, _mmSeconds = 0, _warningTimeout = 5;
let _battleInterval = null, _battleSeconds = 0;

function _startBattleTimer() {
    _battleSeconds = 0;
    battleTimerVal.textContent = '0:00';
    clearInterval(_battleInterval);
    _battleInterval = setInterval(() => { _battleSeconds++; battleTimerVal.textContent = formatTime(_battleSeconds); }, 1000);
    battleTimerHud.classList.add('is-active');
}
function _stopBattleTimer() {
    clearInterval(_battleInterval);
    _battleInterval = null;
    battleTimerHud.classList.remove('is-active');
}

function _startMMTimer() {
    _mmSeconds = 0;
    matchmakingTimer.textContent = '0:00';
    clearInterval(_mmInterval);
    _mmInterval = setInterval(() => { _mmSeconds++; matchmakingTimer.textContent = formatTime(_mmSeconds); }, 1000);
}
function _stopMMTimer() { clearInterval(_mmInterval); _mmInterval = null; }
function _showMatchmaking() { matchmakingOverlay.classList.add('is-active'); _startMMTimer(); }
function _hideMatchmaking() { matchmakingOverlay.classList.remove('is-active'); _stopMMTimer(); }

attackBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    attackBtn.classList.add('hidden');
    _showMatchmaking();
    import('./network.js').then(m => m.SendToServer({ action: 'ATTACK', message: '' }));
});

export function CancelMatchMaking(e) {
    if (e)e.stopPropagation()
    _hideMatchmaking();
    attackBtn.classList.remove('hidden');
}
matchmakingCancel.addEventListener('click', CancelMatchMaking);

export function FoundMatch() {
    _hideMatchmaking();
    _buildDeployBar();
    _showDeployBar();
    attackBtn.classList.add('hidden');
}

export function IncomingAttack() {
    if (_warningTimeout) { clearTimeout(_warningTimeout); _warningTimeout = null; }
    incomingWarning.classList.add('is-active');
    _warningTimeout = setTimeout(() => { incomingWarning.classList.remove('is-active'); _warningTimeout = null; }, 6000);
}

const deployBar        = document.getElementById('deploy-bar');
const deployTroopList  = document.getElementById('deploy-troop-list');
const deployRetreatBtn = document.getElementById('deploy-retreat-btn');

let _selectedTroopId  = null;
let _selectedTroopLvl = null;
let _deployMode       = false;
let _troopButtons     = {};

function _showNoTroops() {
    if (document.getElementById('deploy-no-troops')) return;
    const msg = document.createElement('p');
    msg.id = 'deploy-no-troops';
    msg.textContent = 'No troops';
    deployTroopList.appendChild(msg);
}

function _buildDeployBar() {
    deployTroopList.innerHTML = '';
    _troopButtons = {}; _selectedTroopId = null; _selectedTroopLvl = null;

    for (const key of Object.keys(TrainedTroopsData)) {
        const [troopId, levl] = key.split(',');
        const level = Number(levl);
        const count = TrainedTroopsData[[troopId, level]];
        if (!count || count <= 0) continue;
        const troopDef = AllTroopsData[troopId];
        if (!troopDef) continue;

        const btn     = document.createElement('button');
        btn.className = 'deploy-troop-btn';
        btn.dataset.troopId = troopId; btn.dataset.level = level;

        const img = document.createElement('img');
        img.className = 'deploy-troop-img'; img.src = `./Models/${troopDef.name}.png`; img.alt = troopDef.name;
        img.onerror = () => { img.style.display = 'none'; };

        const nameEl  = document.createElement('p');   nameEl.className  = 'deploy-troop-name';  nameEl.textContent  = `${troopDef.name} (Lv.${level})`;
        const countEl = document.createElement('span'); countEl.className = 'deploy-troop-count'; countEl.textContent = count;

        btn.appendChild(img); btn.appendChild(nameEl); btn.appendChild(countEl);
        deployTroopList.appendChild(btn);
        _troopButtons[[troopId, level]] = { btn, countEl, count };
        btn.addEventListener('click', (e) => { e.stopPropagation(); _onTroopBtnClick(troopId, level); });
    }

    if (!Object.keys(_troopButtons).length) _showNoTroops();
}

function _onTroopBtnClick(troopId, level) {
    const entry = _troopButtons[[troopId, level]];
    if (!entry || entry.count <= 0) return;
    if (_selectedTroopId !== null) {
        const prev = _troopButtons[[_selectedTroopId, _selectedTroopLvl]];
        if (prev) prev.btn.classList.remove('selected');
    }
    if (_selectedTroopId === troopId && _selectedTroopLvl === level) {
        _selectedTroopId = null; _selectedTroopLvl = null; _deployMode = false; return;
    }
    _selectedTroopId = troopId; _selectedTroopLvl = level; _deployMode = true;
    entry.btn.classList.add('selected');
}

function _showDeployBar() {
    deployRetreatBtn.style.display = IsAttacker ? '' : 'none';
    deployBar.style.display = 'flex'; void deployBar.offsetHeight; deployBar.classList.add('is-active');
    deployRetreatBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        if (!inBattle || !IsAttacker) return;
        import('./network.js').then(m => m.SendToServer({ action: 'retreat', message: '' }));
    });

}
export function _hideDeployBar() {
    deployBar.classList.remove('is-active');
    setTimeout(() => { deployBar.style.display = 'none'; }, 320);
    _deployMode = false; _selectedTroopId = null; _selectedTroopLvl = null;
}


window.addEventListener('click', (event) => {
    if (!inBattle || replay) return;
    const rect = canvas.getBoundingClientRect();
    mouse.x =  ((event.clientX - rect.left) / rect.width)  *  2 - 1;
    mouse.y = -((event.clientY - rect.top)  / rect.height) *  2 + 1;
    raycaster.setFromCamera(mouse, camera);
    const intersects = raycaster.intersectObject(ground);
    if (!intersects.length) return;
    const hit = intersects[0].point;
    _spawnTroop(_selectedTroopId, _selectedTroopLvl, Math.round(hit.x / position_scaling), Math.round(hit.z / position_scaling));
});

function _spawnTroop(troopId, level, gridX, gridY) {
    const entry = _troopButtons[[troopId, level]];
    if (!entry || entry.count <= 0) return;
    import('./network.js').then(m => m.SendToServer({ action: 'spawn_troop', troop_id: troopId, troop_level: level, x: gridX, y: gridY }));
}

const PooledArmy = {};
let LoadedArmy = {};

export function SpawnTroop(datafromServer){

    const troopId = datafromServer.troop_id
    const level = datafromServer.troop_level
    if (!replay && datafromServer.spawned_by_attacker===IsAttacker) {
        const entry = _troopButtons[[troopId, level]];
        entry.count--;
        TrainedTroopsData[[troopId, level]] = entry.count;
        entry.countEl.textContent = entry.count;
        if (entry.count <= 0) {
            entry.countEl.classList.add('zero');
            entry.btn.disabled = true;
            entry.btn.classList.remove('selected');

            if (_selectedTroopId === troopId && _selectedTroopLvl === level) {
                _selectedTroopId = null;
                _selectedTroopLvl = null;
                _deployMode = false;
            }
        }
        const anyLeft = Object.values(_troopButtons).some(e => e.count > 0);
        if (!anyLeft) {
            _hideDeployBar();
        }
    }

    if ((PooledArmy[troopId]??[]).length>0){
        const Model = PooledArmy[troopId].pop()
        Model.visible = true
        Model.position.x = datafromServer.spawnedAt_X*position_scaling
        Model.position.z = datafromServer.spawnedAt_Y*position_scaling
        Model.userData.troopId = troopId
        if (datafromServer.spawned_by_attacker)
            state.AliveTroopAttacker.push(Model)
        else
            state.AliveTroopDefender.push(Model)
        if (LoadedArmy[troopId])
            LoadedArmy[troopId].push(Model)
        else LoadedArmy[troopId] = [Model]
    }
    else {
        textureLoader.load(
            `../../Models/${AllTroopsData[troopId].name}.png`,
            (texture) => {
                const mesh = new THREE.PlaneGeometry(size_scaling, size_scaling);
                const aspect = texture.image.width / texture.image.height;

                if (aspect > 1) {
                    texture.repeat.set(1 / aspect, 1);
                    texture.offset.set((1 - 1 / aspect) / 2, 0);
                } else {
                    texture.repeat.set(1, aspect);
                    texture.offset.set(0, (1 - aspect) / 2);
                }

                const material = new THREE.MeshStandardMaterial({
                    map: texture,
                    transparent: true,
                    alphaTest: 0.3,
                    depthTest:false,
                    depthWrite:false
                });

                const Model = new THREE.Mesh(mesh, material);
                Model.rotation.y = -Math.PI / 4;
                Model.position.x = datafromServer.spawnedAt_X*position_scaling
                Model.position.z = datafromServer.spawnedAt_Y*position_scaling

                Model.renderOrder=10
                Model.userData.troopId = troopId
                scene.add(Model);
                if (datafromServer.spawned_by_attacker)
                    state.AliveTroopAttacker.push(Model)
                else
                    state.AliveTroopDefender.push(Model)
                if (LoadedArmy[troopId]) LoadedArmy[troopId].push(Model)
                else LoadedArmy[troopId] = [Model]
            }
        )
    }
}

export function DespawnTroops() {
    for (const [key, armyArr] of Object.entries(LoadedArmy)) {
        for (const model of armyArr) {
            model.visible = false;
            if (PooledArmy[key]) PooledArmy[key].push(model); else PooledArmy[key] = [model];
        }
    }
    LoadedArmy = {};
}

export function simulate(deltaTime) {
    const { AliveTroopAttacker: att, AliveTroopDefender: def, AliveBuildings: aliveB, Buildings: buildings } = state;

    for (let i = 0; i < att.length; i++) {
        const troop  = att[i];
        const config = AllTroopsData[troop.userData.troopId];
        const tX = troop.position.x / position_scaling;
        const tY = troop.position.z / position_scaling;
        const prefCat    = config.preferred_target;
        const hasPref    = prefCat != null;

        let bestBIdx = -1, minBDstSq = Infinity;
        let bestTargetX = 0, bestTargetY = 0;

        const checkBuilding = (pb, bIdx) => {
            const bData = AllBuildingData[pb.building_id];

            const width = bData.grid_size_x -1;
            const height = bData.grid_size_y-1;

            const minX = pb.grid_x;
            const maxX = pb.grid_x + width;
            const minY = pb.grid_y;
            const maxY = pb.grid_y + height;

            const closestX = Math.max(minX, Math.min(tX, maxX));
            const closestY = Math.max(minY, Math.min(tY, maxY));

            const dx = closestX - tX;
            const dy = closestY - tY;
            const dq = dx * dx + dy * dy;

            if (dq < minBDstSq) {
                minBDstSq = dq;
                bestBIdx = bIdx;
                bestTargetX = closestX;
                bestTargetY = closestY;
            }
        };

        for (let abIdx = 0; abIdx < aliveB.length; abIdx++) {
            const bIdx = aliveB[abIdx].BuildingIndex;
            const pb = buildings[bIdx];

            if (hasPref && AllBuildingData[pb.building_id].category !== prefCat && AllBuildingData[pb.building_id].category !== 'wall') continue;

            checkBuilding(pb, bIdx);
        }

        if (hasPref && bestBIdx === -1) {
            for (let abIdx = 0; abIdx < aliveB.length; abIdx++) {
                const bIdx = aliveB[abIdx].BuildingIndex;
                const pb = buildings[bIdx];
                checkBuilding(pb, bIdx);
            }
        }

        let bestDIdx = -1, minDDstSq = Infinity;
        for (let dtIdx = 0; dtIdx < def.length; dtIdx++) {
            const t = def[dtIdx], dx = t.position.x / position_scaling - tX, dy = t.position.z / position_scaling - tY, dq = dx * dx + dy * dy;
            if (dq < minDDstSq) { minDDstSq = dq; bestDIdx = dtIdx; }
        }

        const atkRangeSq = config.attack_range * config.attack_range;
        const troopNearer = bestDIdx !== -1 && minDDstSq < minBDstSq;

        if (troopNearer) {
            if (minDDstSq > atkRangeSq) {
                const target = def[bestDIdx];
                const dist = Math.sqrt(minDDstSq), moveDist = Math.min(config.movement_speed * deltaTime, dist), ratio = moveDist / dist;
                troop.position.x += (target.position.x / position_scaling - tX) * ratio * position_scaling;
                troop.position.z += (target.position.z / position_scaling - tY) * ratio * position_scaling;
            }
        } else if (bestBIdx !== -1 && minBDstSq > atkRangeSq) {
            const dist = Math.sqrt(minBDstSq), moveDist = Math.min(config.movement_speed * deltaTime, dist), ratio = moveDist / dist;
            troop.position.x += (bestTargetX - tX) * ratio * position_scaling;
            troop.position.z += (bestTargetY - tY) * ratio * position_scaling;
        }
    }

    for (let i = 0; i < def.length; i++) {
        const troop  = def[i];
        const config = AllTroopsData[troop.userData.troopId];
        const tX = troop.position.x / position_scaling, tY = troop.position.z / position_scaling;
        let bestIdx = -1, minDstSq = Infinity;
        for (let atIdx = 0; atIdx < att.length; atIdx++) {
            const t = att[atIdx], dx = t.position.x / position_scaling - tX, dy = t.position.z / position_scaling - tY, dq = dx * dx + dy * dy;
            if (dq < minDstSq) { minDstSq = dq; bestIdx = atIdx; }
        }
        if (bestIdx === -1) continue;
        const atkRangeSq = config.attack_range * config.attack_range;
        if (minDstSq > atkRangeSq) {
            const target = att[bestIdx];
            const dist = Math.sqrt(minDstSq), moveDist = Math.min(config.movement_speed * deltaTime, dist), ratio = moveDist / dist;
            troop.position.x += (target.position.x / position_scaling - tX) * ratio * position_scaling;
            troop.position.z += (target.position.z / position_scaling - tY) * ratio * position_scaling;
        }
    }
}
export function DealDamage(building_damage, attacker_troop_damage, defender_troop_damage, building_died, attacker_troop_died, defender_troop_died) {
    const { AliveTroopAttacker: att, AliveTroopDefender: def, AliveBuildings: aliveB, Buildings: buildings } = state;
    const container = canvas.parentElement;
    const W = canvas.clientWidth, H = canvas.clientHeight;

    function worldToScreen(worldPos) {
        const vec = worldPos.clone().project(camera);
        return { x: (vec.x * 0.5 + 0.5) * W, y: (-vec.y * 0.5 + 0.5) * H };
    }
    function spawnLabel(screenX, screenY, text, extraClass = '') {
        const el = document.createElement('div');
        el.className   = 'dmg-label' + (extraClass ? ' ' + extraClass : '');
        el.textContent = text;
        el.style.left  = (screenX + (Math.random() - 0.5) * 30 - 15) + 'px';
        el.style.top   = (screenY - 20) + 'px';
        container.style.position = container.style.position || 'relative';
        container.appendChild(el);
        el.addEventListener('animationend', () => el.remove());
    }

    aliveB.forEach((b, i) => {
        const dmg = building_damage[i]; if (!dmg || dmg <= 0) return;
        spawnLabel(...Object.values(worldToScreen(buildings[b.BuildingIndex].Model.position)), `-${dmg}`, building_died.includes(i) ? 'died' : '');
    });
    att.forEach((troop, i) => {
        const dmg = attacker_troop_damage[i]; if (!dmg || dmg <= 0) return;
        spawnLabel(...Object.values(worldToScreen(troop.position)), `-${dmg}`, attacker_troop_died.includes(i) ? 'troop died' : 'troop');
    });
    def.forEach((troop, i) => {
        const dmg = defender_troop_damage[i]; if (!dmg || dmg <= 0) return;
        spawnLabel(...Object.values(worldToScreen(troop.position)), `-${dmg}`, defender_troop_died.includes(i) ? 'troop died' : 'troop');
    });

    for (const idx of building_died.toReversed()) {
        state.Buildings[aliveB[idx].BuildingIndex].Model.userData.is_broken = true;
        state.AliveBuildings.splice(idx, 1);
    }
    const pool = (arr, died) => {
        for (const idx of died.toReversed()) {
            const ded = arr[idx]; arr.splice(idx, 1);
            ded.visible = false;
            if ((PooledArmy[ded.userData.troopId] ?? []).length > 0) PooledArmy[ded.userData.troopId].push(ded);
            else PooledArmy[ded.userData.troopId] = [ded];
        }
    };
    pool(att, attacker_troop_died);
    pool(def, defender_troop_died);

    if (building_died.length > 0) LoadMap(state.Buildings);
}

function renderTroopRows(container, troopLoss) {
    container.innerHTML = '';
    const entries = Object.entries(troopLoss || {}).filter(([, v]) => v > 0);
    if (!entries.length) {
        const p = document.createElement('p'); p.className = 'bo-no-loss'; p.textContent = 'No losses';
        container.appendChild(p); return;
    }
    entries.forEach(([id, count]) => {
        const data = AllTroopsData[id] || {};
        const row  = document.createElement('div'); row.className = 'bo-troop-row';
        row.appendChild(makeImgWithFallback(`./Models/${data.name || id}.png`, data.name || id, 'bo-troop-img', 'bo-troop-img-fb', '⚔️'));
        const nm = document.createElement('span'); nm.className = 'bo-troop-name'; nm.textContent = data.name || id;
        const ct = document.createElement('span'); ct.className = 'bo-troop-count'; ct.textContent = `${count} fallen`;
        row.appendChild(nm); row.appendChild(ct); container.appendChild(row);
    });
}

export function BattleOver(battle_outcome, attacker_troop_loss, buildings_broken, defender_troop_loss, opponent_username, my_name,battle_id) {
    const isAtt = IsAttacker;
    const dateStr = battle_outcome.fought_at
        ? new Date(battle_outcome.fought_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : '';

    document.getElementById('bo-eyebrow').textContent  = 'Battle Report' + (dateStr ? ' · ' + dateStr : '');
    document.getElementById('bo-headline').textContent = isAtt===battle_outcome.winner_attacker?"🏆 Victory":"💀 Defeat";
    document.getElementById('bo-subline').textContent = replay
        ? `${my_name} raided ${opponent_username}'s village.`
        : (isAtt ? `You struck ${opponent_username}'s village` : `${opponent_username} raided your village`);

    document.getElementById('bo-gold').textContent   = fmt(battle_outcome.gold_looted);
    document.getElementById('bo-elixir').textContent = fmt(battle_outcome.elixir_looted);
    document.getElementById('bo-dark').textContent   = fmt(battle_outcome.dark_elixir_looted);

    document.getElementById('bo-your-title').textContent  = replay ? 'Attacker losses' : (isAtt ? 'Your losses'  : 'Their losses');
    document.getElementById('bo-their-title').textContent = replay ? 'Defender losses' : (isAtt ? 'Their losses' : 'Your losses');
    document.getElementById('bo-section-title-building').textContent = replay
        ? 'Buildings destroyed'
        : (isAtt ? 'Buildings destroyed (their village)' : 'Buildings destroyed (your village)');

    renderTroopRows(document.getElementById('bo-your-troops'),  isAtt ? attacker_troop_loss : defender_troop_loss);
    renderTroopRows(document.getElementById('bo-their-troops'), isAtt ? defender_troop_loss : attacker_troop_loss);

    const bldEntries = Object.entries(buildings_broken || {}).filter(([, v]) => v > 0);
    const bldSection = document.getElementById('bo-buildings-section');
    const bldGrid    = document.getElementById('bo-buildings-grid');
    bldGrid.innerHTML = '';
    if (bldEntries.length) {
        bldEntries.forEach(([id, count]) => {
            const data = AllBuildingData[id] || {};
            const card = document.createElement('div'); card.className = 'bo-bld-card';
            card.appendChild(makeImgWithFallback(`./Models/${data.name || id}.png`, data.name || id, 'bo-bld-img', 'bo-bld-img-fb', '🏚️'));
            card.insertAdjacentHTML('beforeend', `<div><p class="bo-bld-name">${data.name || id}</p><p class="bo-bld-count">×${count} destroyed</p></div>`);
            bldGrid.appendChild(card);
        });
        bldSection.style.display = '';
    } else { bldSection.style.display = 'none'; }

    const revengeBtn = document.getElementById('bo-revenge-btn');
    revengeBtn.style.display = !isAtt ? '' : 'none';
    revengeBtn.addEventListener('click', (e) => { closeBattleOver(e); Revenge(battle_outcome.attacker_name); });

    setInBattle(false);
    _hideDeployBar();

    document.getElementById('battle-over-overlay').classList.add('is-active');
    document.getElementById('bo-replay-btn').addEventListener('click', (e) => {
        closeBattleOver(e)
        console.log(battle_id)
        SendToServer({action:"REPLAY",message:battle_id})
    });
}
document.getElementById('battle-over-overlay').addEventListener('click', (e) => {
    if (e.target.id === 'battle-over-overlay') closeBattleOver(e);
});
function closeBattleOver(e) { if (e) e.stopPropagation(); document.getElementById('battle-over-overlay').classList.remove('is-active'); }
document.getElementById('bo-close-btn').addEventListener('click', closeBattleOver);
