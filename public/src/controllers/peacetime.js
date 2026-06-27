import { ConstructionType } from '../core/enums.js';
import {
    AllBuildingData, AllTroopsData, PlacedBuildings, TrainedTroopsData, ConstructionTasks,
    setUserData, setPlacedBuildings, setTrainedTroopsData, setConstructionTasks,
    LoadMap, AddLevelDetails, SummontaskCountDown, UserData,
} from '../models/map.js';
import { UpdateResourceUI } from '../views/ui_hud.js';
import {
    FoundMatch, CancelMatchMaking, IncomingAttack,
    setInBattle, setIsAttacker, setReplay, setState,
} from './battle.js';
import { scene } from '../core/scene.js';
import { renderUserInfo, LoadedMoreBattles } from '../views/profile.js';
import {SendToServer} from "./network.js";


export function handlePeacetimeMessage(data) {
    switch (data.msg_type) {
        case 'building_troop_of_user': handleInitialUserLoad(data); break;
        case 'building_troop':         handleStaticGameData(data); break;
        case 'construction_started':   handleConstructionStarted(data); break;
        case 'construction_completed': handleConstructionCompleted(data); break;
        case 'resource_collected':     handleResourceCollected(data); break;
        case 'troop_training_started': handleTroopTrainingStarted(data); break;
        case 'un_attack':              CancelMatchMaking(); break;
        case 'incoming_attack':        handleIncomingAttack(data); break;
        case 'battle_start':           handleBattleStart(data); break;
        case 'replay':                 handleReplay(data); break;
        case 'battle_history':         LoadedMoreBattles(data.history); break;
        case 'moved':                  handleBuildingMoved(data); break;
    }
}

function handleInitialUserLoad(data) {
    setUserData(data.user_data, false);
    UserData.username = data.user.username;
    UserData.email = data.user.email;
    renderUserInfo();
    setPlacedBuildings(data.building);
    localStorage.setItem('Placed_building', JSON.stringify(data.building));
    setConstructionTasks(data.construction_tasks);

    const trained = {};
    for (const troop of data.troops) trained[[troop.troop_id, troop.level]] = troop.count;
    setTrainedTroopsData(trained);
    localStorage.setItem('Trained_troops_data', JSON.stringify(trained));

    SendToServer({ action: 'ALL_BUILDING_TROOP_DATA', message: '' });
    SendToServer({ action: 'CHECK_CONSTRUCTION_WORK', message: '' });
    UpdateResourceUI();
}

function handleStaticGameData(data) {
    const { building: buildings, defence, army, resource, troops } = data;

    for (const troop of troops) AllTroopsData[troop.id] = troop;

    for (const building of buildings) {
        AllBuildingData[building.building_id] = {
            name: building.name,
            category: building.category,
            grid_size_x: building.grid_size_x,
            grid_size_y: building.grid_size_y,
            levels: {},
        };
    }
    for (const res of resource) {
        AllBuildingData[res.building_id].resource_type = res.resource_type;
    }
    for (const def of defence) {
        AllBuildingData[def.building_id].attack_speed_seconds = def.attack_speed_seconds;
        AllBuildingData[def.building_id].attack_range = def.attack_range;
        AllBuildingData[def.building_id].damage_type = def.damage_type;
        AllBuildingData[def.building_id].unit_target = def.unit_target;
    }

    AddLevelDetails(data.particular_level_data);
    localStorage.setItem('All_building_data', JSON.stringify(AllBuildingData));
    localStorage.setItem('All_troops_data', JSON.stringify(AllTroopsData));

    if (PlacedBuildings.length) LoadMap(PlacedBuildings);
    else SendToServer({ action: 'INITIAL_LOAD', message: '' });
}

function handleConstructionStarted(data) {
    ConstructionTasks.push(data.task);
    if (data.placed_building != null) {
        PlacedBuildings.push(data.placed_building);
    }
    LoadMap(PlacedBuildings);
    setUserData(data.user_data);
    UpdateResourceUI();
}

function handleConstructionCompleted(data) {
    AddLevelDetails(data.particular_level_detail);
    localStorage.setItem('All_building_data', JSON.stringify(AllBuildingData));

    const doneIds = new Set(data.construction_done.map(t => t.id));
    const kept = [], removed = [];
    for (const task of ConstructionTasks) (doneIds.has(task.id) ? removed : kept).push(task);
    setConstructionTasks(kept);

    for (const task of removed) {
        applyCompletedTask(task);
        removeCountdownSprite(task);
    }
    localStorage.setItem('Trained_troops_data', JSON.stringify(TrainedTroopsData));

    setUserData(data.user_data);
    UpdateResourceUI();
    LoadMap(PlacedBuildings);
}

function applyCompletedTask(task) {
    if (task.task_type === ConstructionType.TroopTraining) {
        const key = [task.troop_id, task.troop_level_to];
        TrainedTroopsData[key] = (TrainedTroopsData[key] ?? 0) + task.troop_count;
    } else if (task.task_type === ConstructionType.BuildingRepair) {
        PlacedBuildings.find(b => b.id === task.placed_building_id).is_broken = false;
    } else {
        PlacedBuildings.find(b => b.id === task.placed_building_id).level += 1;
    }
}

function removeCountdownSprite(task) {
    const sprite = task.building?.countdownSprite;
    if (!sprite) return;
    scene.remove(sprite);
    sprite.material.dispose();
    sprite.geometry.dispose();
    delete task.building.countdownSprite;
}

function handleResourceCollected(data) {
    setUserData(data.user_data);
    if (data.placed_buildings) {
        setPlacedBuildings(data.placed_buildings);
        LoadMap(PlacedBuildings);
    }
    UpdateResourceUI();
}

function handleTroopTrainingStarted(data) {
    ConstructionTasks.push(data.task);
    SummontaskCountDown(data.task);
}

function handleIncomingAttack(data) {
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
}

function handleBattleStart(data) {
    setState({
        Buildings: data.defender_building, TroopSpawns: [],
        AliveTroopAttacker: [], AliveTroopDefender: [], AliveBuildings: data.alive_buildings,
    });
    LoadMap(data.defender_building);
    setInBattle(true);
    setIsAttacker(true);
    FoundMatch();
}

function handleReplay(data) {
    setReplay(true);
    setInBattle(true);
    setIsAttacker(true);
    setState({
        Buildings: data.defender_building, TroopSpawns: [],
        AliveTroopAttacker: [], AliveTroopDefender: [], AliveBuildings: data.alive_buildings,
    });
    LoadMap(data.defender_building);
    document.getElementById('attack-btn').classList.add('hidden');
}

function handleBuildingMoved(data) {
    const building = PlacedBuildings.find(b => b.id === data.placed_building_id);
    building.grid_x = data.grid_x;
    building.grid_y = data.grid_y;
    LoadMap(PlacedBuildings);
}