import { ConstructionType } from '../core/enums.js';
import {
    AllBuildingData, AllTroopsData, PlacedBuildings, TrainedTroopsData, ConstructionTasks,
    setUserData, setPlacedBuildings, setTrainedTroopsData, setConstructionTasks,
    LoadMap, AddLevelDetails, SummontaskCountDown,
} from '../models/map.js';
import { UpdateResourceUI } from '../views/ui_hud.js';
import {
    FoundMatch, CancelMatchMaking, IncomingAttack, SpawnTroop, DespawnTroops, BattleOver,
    _hideDeployBar, setInBattle, setIsAttacker, setReplay, setState, inBattle, DealDamage,
} from './battle.js';
import {scene} from "../core/scene.js";

export let access_token = localStorage.getItem('access_token');

// region Auth

async function refreshAuthToken() {
    const refreshToken = localStorage.getItem('refresh_token_b64');
    try {
        const response = await fetch('/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user_id: AllBuildingData._userData?.user_id, refresh_token: refreshToken }),
        });
        if (!response.ok) throw new Error('Session expired');
        const data = await response.json();
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token_b64', data.refresh_token_b64);
        access_token = data.access_token;
    } catch (error) {
        console.error('Refresh failed, redirecting to login:', error);
        window.location.href = '/Login.html';
    }
}

setInterval(refreshAuthToken, 14 * 60 * 1000);

// endregion

// region WebSocket

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
    const protocol  = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const socketUrl = `${protocol}//${window.location.host}/ws?token=${access_token}`;
    socket = new WebSocket(socketUrl);

    socket.addEventListener('open',    ()      => SendToServer({ action: 'INITIAL_LOAD', message: '' }));
    socket.addEventListener('error',   (err)   => console.error('WebSocket Error:', err));
    socket.addEventListener('close',   ()      => { console.warn('Disconnected. Reconnecting…'); connectToGameServer(); });
    socket.addEventListener('message', (event) => {
        const data = JSON.parse(event.data);
        if (!inBattle) _handlePeacetimeMessage(data);
        else           _handleBattleMessage(data);
        if (data.status === 'error') showToast(data.message);
    });
}

// endregion

// region Peacetime Messages

function _handlePeacetimeMessage(data) {
    switch (data.msg_type) {
        case 'building_troop_of_user': {
            console.log("uaer data",data.user_data)
            setUserData(data.user_data);
            setPlacedBuildings(data.building);
            localStorage.setItem('Placed_building', JSON.stringify(data.building));
            setConstructionTasks(data.construction_tasks);

            const trained = {};
            for (const troop of data.troops) trained[[troop.troop_id, troop.level]] = troop.count;
            setTrainedTroopsData(trained);
            localStorage.setItem('Trained_troops_data', JSON.stringify(trained));

            if (Object.keys(AllBuildingData).length && Object.keys(AllTroopsData).length)
                LoadMap(PlacedBuildings);

            SendToServer({ action: 'ALL_BUILDING_TROOP_DATA', message: '' });
            SendToServer({ action: 'CHECK_CONSTRUCTION_WORK', message: '' });
            UpdateResourceUI();
            break;
        }

        case 'building_troop': {
            const buildings = data.building
            const defence = data.defence
            const army = data.army
            const resource = data.resource

            for (const troop of data.troops) {
                AllTroopsData[troop.id] = troop
            }

            localStorage.setItem('All_troops_data', JSON.stringify(AllTroopsData))
            for (const building of buildings) {
                AllBuildingData[building.building_id] = {
                    name: building.name,
                    category: building.category,
                    grid_size_x: building.grid_size_x,
                    grid_size_y: building.grid_size_y,
                    levels: {}
                }
            }
            for (const res of resource) {
                AllBuildingData[res.building_id].resource_type = res.resource_type
            }
            for (const arm of army) {

            }
            for (const def of defence) {
                AllBuildingData[def.building_id].attack_speed_seconds = def.attack_speed_seconds
                AllBuildingData[def.building_id].attack_range = def.attack_range
                AllBuildingData[def.building_id].damage_type = def.damage_type
                AllBuildingData[def.building_id].unit_target = def.unit_target
            }

            AddLevelDetails(data.particular_level_data);
            localStorage.setItem('All_building_data', JSON.stringify(AllBuildingData));
            localStorage.setItem('All_troops_data',   JSON.stringify(AllTroopsData));

            if (PlacedBuildings.length) LoadMap(PlacedBuildings);
            else SendToServer({ action: 'INITIAL_LOAD', message: '' });
            break;
        }

        case 'construction_started':
            ConstructionTasks.push(data.task);
            if (data.placed_building != null) {
                PlacedBuildings.push(data.placed_building);
                LoadMap(PlacedBuildings);
            } else {
                SummontaskCountDown(data.task);
            }
            setUserData(data.user_data)
            UpdateResourceUI()
            break;

        case 'construction_completed': {
            AddLevelDetails(data.particular_level_detail);
            localStorage.setItem('All_building_data', JSON.stringify(AllBuildingData));
            const doneIds = new Set(data.construction_done.map(t => t.id));
            const kept = [], removed = [];
            for (const task of ConstructionTasks) (doneIds.has(task.id) ? removed : kept).push(task);
            setConstructionTasks(kept);

            for (const el of removed) {
                if (el.task_type === ConstructionType.TroopTraining) {
                    const key = [el.troop_id, el.troop_level_to];
                    TrainedTroopsData[key] = (TrainedTroopsData[key] ?? 0) + el.troop_count;
                } else if (el.task_type === ConstructionType.BuildingRepair) {
                    PlacedBuildings.find(b => b.id === el.placed_building_id).is_broken = false;
                } else {
                    PlacedBuildings.find(b => b.id === el.placed_building_id).level += 1;
                }
                localStorage.setItem('Trained_troops_data', JSON.stringify(TrainedTroopsData));
                const sprite = el.building?.Model?.userData?.countdownSprite;
                if (sprite) {
                    scene.remove(sprite);
                    sprite.material.dispose();
                    sprite.geometry.dispose();
                    delete el.building.Model.userData.countdownSprite;
                }
            }
            setUserData(data.user_data);
            UpdateResourceUI();
            LoadMap(PlacedBuildings);
            break;
        }

        case 'resource_collected':
            setUserData(data.user_data);
            UpdateResourceUI();
            break;

        case 'troop_training_started':
            ConstructionTasks.push(data.task);
            SummontaskCountDown(data.task);
            break;

        case 'un_attack':
            CancelMatchMaking();
            break;

        case 'incoming_attack':
            setInBattle(true);
            setIsAttacker(false);
            IncomingAttack();
            setState({
                Buildings: data.defender_building, TroopSpawns: [],
                AliveTroopAttacker: [], AliveTroopDefender: [], AliveBuildings: data.alive_buildings,
            });
            LoadMap(data.defender_building);
            FoundMatch();
            SendToServer({ action: 'DEFEND', message: '' });
            break;

        case 'battle_start':
            setState({
                Buildings: data.defender_building, TroopSpawns: [],
                AliveTroopAttacker: [], AliveTroopDefender: [], AliveBuildings: data.alive_buildings,
            });
            LoadMap(data.defender_building);
            setInBattle(true);
            setIsAttacker(true);
            FoundMatch();
            break;

        case 'replay':
            setReplay(true);
            setInBattle(true);
            setIsAttacker(true);
            setState({
                Buildings: data.defender_building, TroopSpawns: [],
                AliveTroopAttacker: [], AliveTroopDefender: [], AliveBuildings: data.alive_buildings,
            });
            LoadMap(data.defender_building);
            document.getElementById('attack-btn').classList.add('hidden');
            break;
    }
}

// endregion

// region Battle Messages

function _handleBattleMessage(data) {
    switch (data.msg_type) {
        case 'spawn_troop':
            SpawnTroop(data.troop);
            break;
        case 'battle_update':
            DealDamage(
                data.building_damage, data.attacker_troop_damage, data.defender_troop_damage,
                data.building_died, data.attacker_troop_died, data.defender_troop_died,
            );
            break;
        case 'battle_over':
            console.log("battleOver",data)
            DespawnTroops();
            BattleOver(data.battle_outcome, data.attacker_troop_loss, data.buildings_broken, {}, data.opponent_username, data.my_username,data.battle_id);
            setReplay(false);
            _hideDeployBar();
            setInBattle(false);
            CancelMatchMaking();
            SendToServer({ action: 'INITIAL_LOAD', message: '' });
            break;
    }
}

// endregion

// region Action Senders

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
    SendToServer({ action: 'TRAIN_TROOP', message: JSON.stringify({ barrack_placed_building_id, troop_id, count, use_gems, level_from: level_to - 1 }) });
}
export function Revenge(opponentId) {
    SendToServer({ action: 'REVENGE', message: opponentId });
}

// endregion

function showToast(message, type = 'error') {
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