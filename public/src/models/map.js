import * as THREE from '../../THREE/three.module.js';
import {
    scene,
    textureLoader,
    position_scaling,
    size_scaling,
    highlightGridSquares, camera
} from '../core/scene.js';
import { BuildingCategory } from '../core/enums.js';
import { formatTime } from './utils.js';
import {inBattle} from "../controllers/battle.js";

// region State

export let AllBuildingData   = JSON.parse(localStorage.getItem('All_building_data')  || '{}');
export let PlacedBuildings   = JSON.parse(localStorage.getItem('Placed_building')    || '[]');
export let AllTroopsData     = JSON.parse(localStorage.getItem('All_troops_data')    || '{}');
export let TrainedTroopsData = JSON.parse(localStorage.getItem('Trained_troops_data')|| '{}');
export let Grid              = {};
export let UserData          = null;
export let ConstructionTasks = null;

export function setUserData(data,keepName_mail = true) {

    if (keepName_mail) {
        data.username = UserData.username
        data.email = UserData.email
    }
    UserData = data;
}
export function setPlacedBuildings(arr)   { PlacedBuildings   = arr;  }
export function setTrainedTroopsData(obj) { TrainedTroopsData = obj;  }
export function setConstructionTasks(arr) { ConstructionTasks = arr;  }
export function setAllBuildingData(obj)   { AllBuildingData   = obj;  }
export function setAllTroopsData(obj)     { AllTroopsData     = obj;  }

// endregion

// region Object Pool

/** @type {Record<string, THREE.Object3D[]>} */
const LoadedObjects = {};
/** @type {Record<string, THREE.Object3D[]>} */
const Pool = {};

// endregion

// region Countdown Sprites

export const activeCountdowns = new Set();

export function SummontaskCountDown(task) {
    const placed_building = PlacedBuildings.find(b => b.id === task.placed_building_id);
    task.building = placed_building;

    const countdown = createCountdownSprite(
        task.duration_seconds * 1000,
        new Date(task.started_at).getTime(),
        () => { import('../controllers/network.js').then(m => m.SendToServer({ action: 'CHECK_CONSTRUCTION_WORK', message: '' })); }
    );
    countdown.position.copy(placed_building.Model.position);
    countdown.position.y +=  size_scaling + 0.6;
    scene.add(countdown);
    placed_building.countdownSprite = countdown;
}

function createCountdownSprite(durationMilliSeconds, started_at, OnDone) {
    const cvs    = document.createElement('canvas');
    cvs.width    = 256;
    cvs.height   = 128;

    const texture  = new THREE.CanvasTexture(cvs);
    const geometry = new THREE.PlaneGeometry(1.5 * size_scaling, size_scaling * 0.75);
    const material = new THREE.MeshBasicMaterial({
        map: texture, transparent: true,
        depthWrite: false, depthTest: false, side: THREE.DoubleSide,
    });

    const sprite = new THREE.Mesh(geometry, material);
    sprite.renderOrder = 200
    const state  = { endTime: started_at + durationMilliSeconds, canvas: cvs, texture, done: false, OnDone, durationMilliSeconds };
    drawCountdown(state);
    sprite.userData.countdown = state;
    sprite.rotation.y = Math.atan2(
        camera.position.x - sprite.position.x,
        camera.position.z - sprite.position.z
    );
    activeCountdowns.add(sprite);
    return sprite;
}

function drawCountdown(state) {
    const ctx = state.canvas.getContext('2d');
    const cvs = state.canvas;
    ctx.clearRect(0, 0, cvs.width, cvs.height);

    const remaining = Math.max(0, Math.ceil((state.endTime - Date.now()) / 1000));
    const pad = 16;

    ctx.fillStyle = 'rgba(20, 20, 30, 0.72)';
    roundRect(ctx, pad, pad, cvs.width - pad * 2, cvs.height - pad * 2, 20);
    ctx.fill();

    const barX = 28, barY = cvs.height - 36, barW = cvs.width - 56, barH = 8;
    ctx.fillStyle = 'rgba(255,255,255,0.15)';
    roundRect(ctx, barX, barY, barW, barH, 4);
    ctx.fill();

    ctx.fillStyle = remaining > 60 ? '#4ade80' : remaining > 20 ? '#fbbf24' : '#f87171';
    roundRect(ctx, barX, barY, Math.max(8, barW * (remaining / state.durationMilliSeconds)), barH, 4);
    ctx.fill();

    ctx.fillStyle    = '#ffffff';
    ctx.font         = 'bold 44px sans-serif';
    ctx.textAlign    = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(formatTime(remaining), cvs.width / 2, cvs.height / 2 - 8);

    state.texture.needsUpdate = true;
    state.done = remaining <= 0;
}

function roundRect(ctx, x, y, w, h, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.lineTo(x + w - r, y);
    ctx.quadraticCurveTo(x + w, y, x + w, y + r);
    ctx.lineTo(x + w, y + h - r);
    ctx.quadraticCurveTo(x + w, y + h, x + w - r, y + h);
    ctx.lineTo(x + r, y + h);
    ctx.quadraticCurveTo(x, y + h, x, y + h - r);
    ctx.lineTo(x, y + r);
    ctx.quadraticCurveTo(x, y, x + r, y);
    ctx.closePath();
}

export function tickCountdowns(camera) {
    for (const sprite of activeCountdowns) {
        const cd = sprite.userData.countdown;
        sprite.rotation.y = Math.atan2(
            camera.position.x - sprite.position.x,
            camera.position.z - sprite.position.z
        );
        if (cd.done) {
            cd.OnDone();
            activeCountdowns.delete(sprite);
            continue;
        }
        const secondsLeft = Math.ceil((cd.endTime - Date.now()) / 1000);
        if (secondsLeft !== cd._lastDrawnSecond) {
            cd._lastDrawnSecond = secondsLeft;
            drawCountdown(cd);
        }
    }
}

// endregion

// region Building Data

export function AddLevelDetails(ParticularLevelData) {
    for (const [, val] of Object.entries(ParticularLevelData)) {
        const b  = AllBuildingData[val.building_id];
        const lv = val.level;
        b.levels[lv] = {
            health:                         val.base_stats.health,
            update_gold_required:           val.upgrade_cost.gold_required,
            update_elxir_required:          val.upgrade_cost.elixir_required,
            update_dark_elixir_required:    val.upgrade_cost.dark_elixir_required,
            update_or_gem_required:         val.upgrade_cost.or_gem_required,
            update_time_required_required:  val.upgrade_cost.time_required_seconds,
            update_townhall_level_required: val.upgrade_cost.town_hall_level_required,
        };
        switch (b.category) {
            case BuildingCategory.Army:
                b.levels[lv].troop_capacity = val.details.troop_capacity; break;
            case BuildingCategory.Resource:
                b.levels[lv].generation_rate  = val.details.generation_rate_per_hour;
                b.levels[lv].storage_capacity = val.details.storage_capacity; break;
            case BuildingCategory.Defense:
                b.levels[lv].damage_per_shot = val.details.damage_per_shot; break;
        }
    }
}

// endregion

// region LoadMap

export async function LoadMap(buildings) {
    const textureLoadPromises = [];
    Grid = {};

    for (const [key, val] of Object.entries(LoadedObjects)) {
        for (const Model of val) {
            Model.visible = false;
            if (key in Pool) Pool[key].push(Model);
            else Pool[key] = [Model];
            const sprite = Model.userData.countdownSprite
            if (sprite) {
                activeCountdowns.delete(sprite)
                scene.remove(sprite);
                sprite.material.dispose();
                sprite.geometry.dispose();
                delete Model.userData.countdownSprite;
            }
            const collectGizmo = Model.userData.collectGizmo;
            if (collectGizmo){
                scene.remove(collectGizmo)
            }
        }
        val.length = 0;
    }

    for (const building of buildings) {
        const name = building.is_broken ? 'Broken' : AllBuildingData[building.building_id].name;
        const bData = AllBuildingData[building.building_id];
        const posX  = (building.grid_x + (bData.grid_size_y===1?0:(bData.grid_size_x / 2))) * position_scaling;
        const posZ  = (building.grid_y + (bData.grid_size_y===1?0:(bData.grid_size_y / 4))) * position_scaling;
        const name_key = name+bData.grid_size_x+bData.grid_size_y
        if (name_key in Pool && Pool[name_key].length > 0) {
            const Model = Pool[name_key].pop();
            Model.position.set(posX, 0, posZ);
            Model.visible  = true;
            Model.userData = building;
            if (name_key in LoadedObjects) LoadedObjects[name_key].push(Model);
            else LoadedObjects[name_key] = [Model];
            _fillGrid(building, Model);
            building.Model = Model;
        } else {
            const loadPromise = new Promise((resolve) => {
                textureLoader.load(
                    `./Models/${name}.png`,
                    (texture) => {
                        const mesh   = new THREE.PlaneGeometry(bData.grid_size_x * size_scaling, bData.grid_size_y * size_scaling);
                        const aspect = texture.image.width / texture.image.height;

                        if (aspect > 1) { texture.repeat.set(1 / aspect, 1); texture.offset.set((1 - 1 / aspect) / 2, 0); }
                        else            { texture.repeat.set(1, aspect);     texture.offset.set(0, (1 - aspect) / 2); }

                        const material = new THREE.MeshStandardMaterial({
                            map: texture, transparent: true, alphaTest: 0.3, depthTest: false, depthWrite: false,
                        });
                        const obj = new THREE.Mesh(mesh, material);
                        obj.rotation.y = -Math.PI / 4;
                        obj.position.set(posX, 0, posZ);
                        obj.renderOrder = 10;
                        obj.userData    = building;
                        _fillGrid(building, obj);
                        scene.add(obj);
                        building.Model = obj;
                        if (name_key in LoadedObjects) LoadedObjects[name_key].push(obj);
                        else LoadedObjects[name_key] = [obj];
                        resolve();
                    },
                    undefined,
                    (error) => { console.error(`Failed to load texture for ${name}:`, error); resolve(); }
                );
            });
            textureLoadPromises.push(loadPromise);
        }
    }

    await Promise.all(textureLoadPromises);
    if (!inBattle)   {
        for (const task of ConstructionTasks) {
            SummontaskCountDown(task);
        }
    }
    highlightGridSquares(Grid);
}

function _fillGrid(building, model) {
    const { grid_size_x, grid_size_y } = AllBuildingData[building.building_id];
    for (let i = building.grid_x; i < building.grid_x + grid_size_x; i++)
        for (let j = building.grid_y; j < building.grid_y + grid_size_y; j++)
            Grid[[i, j]] = model;
}

// endregion