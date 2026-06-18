import * as THREE from './THREE/three.module.js';
import {GLTFLoader} from './THREE/GLTFLoader.js';

// const gltfLoader = new GLTFLoader();
const scene = new THREE.Scene();
const clock = new THREE.Clock()

// region Three.js setup
const aspect = window.innerWidth / window.innerHeight;
scene.background = new THREE.Color(0x87CEEB);// blue color
const isOrthographic = true;
let camera;
let frustumSize = 240;
const Right = [1, 1];
const Forward = [1, -1];
if (!isOrthographic){
    camera = new THREE.PerspectiveCamera(60, aspect, 0.1, 10000)
    camera.position.set(-60, 60, 60);
    camera.lookAt(0, 0, 0);
}
else {
    const aspect = window.innerWidth / window.innerHeight;
    camera = new THREE.OrthographicCamera(
        (frustumSize * aspect) / -2,
        (frustumSize * aspect) / 2,
        frustumSize / 2,
        frustumSize / -2,
        0.1,
        10000
    );

    camera.position.set(-1000, 1000, 1000);
    camera.lookAt(0, 0, 0);
}

const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
document.getElementById('game-container').appendChild(renderer.domElement);
const canvas= renderer.domElement
const ambientLight = new THREE.AmbientLight(0xffffff, 0.3);
scene.add(ambientLight);

// const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
// dirLight.position.set(20, 40, 20);
// scene.add(dirLight);

const textureLoader = new THREE.TextureLoader();
const gridTexture = textureLoader.load('./Models/Map.jpeg');
// gridTexture.size/=4
const groundGeometry = new THREE.PlaneGeometry(2000, 2000);

const groundMaterial = new THREE.MeshStandardMaterial({
    map: gridTexture,
    transparent: true
});
const ground = new THREE.Mesh(groundGeometry, groundMaterial);
ground.renderOrder = -1
ground.rotation.z = -Math.PI/4;
ground.rotation.x = -Math.PI / 2;
ground.position.set(0,0,0)
scene.add(ground);

const keys = { w: false, a: false, s: false, d: false ,e:false,q:false};
const moveSpeed = 200;

window.addEventListener('keydown', (e) => {
    if (keys.hasOwnProperty(e.key.toLowerCase())) keys[e.key.toLowerCase()] = true;
});
window.addEventListener('keyup', (e) => {
    if (keys.hasOwnProperty(e.key.toLowerCase())) keys[e.key.toLowerCase()] = false;
});
function handleMovement(dt) {
    if (keys.d) {camera.position.z += moveSpeed*dt*Right[1]; camera.position.x += Right[0]*moveSpeed*dt }
    if (keys.w) {camera.position.z += moveSpeed*dt*Forward[1];camera.position.x+=moveSpeed*dt*Forward[0];}
    if (keys.a) {camera.position.z -= moveSpeed*dt*Right[1]; camera.position.x -= Right[0]*moveSpeed*dt }
    if (keys.s) {camera.position.z -= moveSpeed*dt*Forward[1];camera.position.x-=moveSpeed*dt*Forward[0];}
    if (keys.e) if (isOrthographic) {frustumSize+=moveSpeed*dt; UpdateCamera();} else camera.position.y+=moveSpeed*dt;
    if (keys.q) if (isOrthographic) {frustumSize-=moveSpeed*dt; UpdateCamera();}else camera.position.y-=moveSpeed*dt;
}
const activeCountdowns = new Set();

function animate() {
    requestAnimationFrame(animate);
    var dt = clock.getDelta()
    handleMovement(dt);
    updateCollectButton()
    tickCountdowns(scene,camera);
    if (inBattle){
        try{
            simulate(dt)
        }
        catch (e){
            console.log(e)
            inBattle = false
        }
    }
    ambientLight.intensity = Math.max(Math.min(0.3 + Math.cos(clock.elapsedTime*0.01) * 0.5,0.5),0.1) //day-night cycle faking
    renderer.render(scene, camera);

}
animate();

function UpdateCamera(){
    const newAspect = window.innerWidth / window.innerHeight;
    if (isOrthographic) {
        camera.left = (frustumSize * newAspect) / -2;
        camera.right = (frustumSize * newAspect) / 2;
        camera.top = frustumSize / 2;
        camera.bottom = frustumSize / -2;
    } else {
        camera.aspect = newAspect;
    }
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
}
window.addEventListener('resize', () => {
    UpdateCamera();
});


// endregion

// region Game And Map
let AllBuildingData = JSON.parse(localStorage.getItem('All_building_data') || '{}');
let PlacedBuildings = JSON.parse(localStorage.getItem('Placed_building') || '[]');
let AllTroopsData = JSON.parse(localStorage.getItem('All_troops_data')||'{}')
let TrainedTroopsData = JSON.parse(localStorage.getItem('Trained_troops_data')||'{}')
let Grid = {}
let UserData = null
let ConstructionTasks = null
/** @type {Record<string, THREE.Object3D[]>} */
const LoadedObjects = {}
/** @type {Record<string, THREE.Object3D[]>} */
const Pool = {}
const position_scaling = 20
const size_scaling = 15
async function LoadMap(PlacedBuildings) {
    const textureLoadPromises = [];
    Grid = {}
    for (const [key, val] of Object.entries(LoadedObjects)) {
        for (const Model of val) {
            Model.visible = false
            if (key in Pool)
                Pool[key].push(Model);
            else Pool[key] = [Model]
            if (Model.userData.countdownSprite) {
                scene.remove(Model.userData.countdownSprite);
                Model.userData.countdownSprite.material.dispose();
                Model.userData.countdownSprite.geometry.dispose();
                delete Model.userData.countdownSprite;
            }
        }
        val.length = 0
    }

    for (const building of PlacedBuildings) {
        const name = building.is_broken?"Broken":AllBuildingData[building.building_id].name
        console.log(name)
        if (name in Pool && Pool[name].length > 0) {
            const Model = Pool[name].pop()
            Model.position.set((building.grid_x+AllBuildingData[building.building_id].grid_size_x/2) * position_scaling, 0, (building.grid_y +AllBuildingData[building.building_id].grid_size_y/4)* position_scaling);
            Model.visible = true
            Model.userData = building;

            if (name in LoadedObjects)
                LoadedObjects[name].push(Model)
            else LoadedObjects[name] = [Model]
            for (let i = building.grid_x; i < building.grid_x+AllBuildingData[building.building_id].grid_size_x; i++) {
                for (let j = building.grid_y; j < building.grid_y+AllBuildingData[building.building_id].grid_size_y; j++) {
                    Grid[[i,j]] = Model
                }
            }
            building.Model = Model
        } else {
            const loadPromise = new Promise((resolve, reject) => {
                textureLoader.load(
                    `./Models/${name}.png`,
                    (texture) => {
                        const mesh = new THREE.PlaneGeometry(AllBuildingData[building.building_id].grid_size_x*size_scaling, AllBuildingData[building.building_id].grid_size_y*size_scaling);
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

                        const object_building = new THREE.Mesh(mesh, material);
                        object_building.rotation.y = -Math.PI / 4;
                        object_building.position.set((building.grid_x+AllBuildingData[building.building_id].grid_size_x/2) * position_scaling, 0, (building.grid_y +AllBuildingData[building.building_id].grid_size_y/4)* position_scaling);

                        object_building.userData = building;
                        object_building.renderOrder=10
                        for (let i = building.grid_x; i < building.grid_x+AllBuildingData[building.building_id].grid_size_x; i++) {
                            for (let j = building.grid_y; j < building.grid_y+AllBuildingData[building.building_id].grid_size_y; j++) {
                                Grid[[i,j]] = object_building
                            }
                        }
                        scene.add(object_building);
                        building.Model = object_building;
                        if (name in LoadedObjects)
                            LoadedObjects[name].push(object_building)
                        else LoadedObjects[name] = [object_building]
                        resolve();
                    },
                    undefined,
                    (error) => {
                        console.error(`Failed to load texture for ${name}:`, error);
                        resolve();
                    }
                );
            });

            textureLoadPromises.push(loadPromise);
        }
    }

    await Promise.all(textureLoadPromises);


    for (const constructionTask of ConstructionTasks) {
        SummontaskCountDown(constructionTask)
    }
    refreshGridHighlights()
}

function AddLevelDetails(ParticularLevelData){

    for (const [key,val] of Object.entries(ParticularLevelData)){
        AllBuildingData[val.building_id].levels[val.level] = {}
        AllBuildingData[val.building_id].levels[val.level].health=val.base_stats.health
        AllBuildingData[val.building_id].levels[val.level].update_gold_required=val.upgrade_cost.gold_required
        AllBuildingData[val.building_id].levels[val.level].update_elxir_required=val.upgrade_cost.elixir_required
        AllBuildingData[val.building_id].levels[val.level].update_dark_elixir_required=val.upgrade_cost.dark_elixir_required
        AllBuildingData[val.building_id].levels[val.level].update_or_gem_required=val.upgrade_cost.or_gem_required
        AllBuildingData[val.building_id].levels[val.level].update_time_required_required=val.upgrade_cost.time_required_seconds
        AllBuildingData[val.building_id].levels[val.level].update_townhall_level_required=val.upgrade_cost.town_hall_level_required
        switch (AllBuildingData[val.building_id].category){
            case BuildingCategory.TownHall:
                break
            case BuildingCategory.Army:
                AllBuildingData[val.building_id].levels[val.level].troop_capacity = val.details.troop_capacity
                break
            case BuildingCategory.Resource:
                AllBuildingData[val.building_id].levels[val.level].generation_rate = val.details.generation_rate_per_hour
                AllBuildingData[val.building_id].levels[val.level].storage_capacity = val.details.storage_capacity
                break
            case BuildingCategory.Defense:
                AllBuildingData[val.building_id].levels[val.level].damage_per_shot= val.details.damage_per_shot
                break
        }
    }

}
function UpdateResourceUI() {

    document.getElementById('hud-gold-val').textContent       = formatNum(UserData.current_gold        ?? 0);
    document.getElementById('hud-elixir-val').textContent     = formatNum(UserData.current_elixir      ?? 0);
    document.getElementById('hud-dark-elixir-val').textContent= formatNum(UserData.current_dark_elixir ?? 0);
    document.getElementById('hud-gems-val').textContent       = formatNum(UserData.current_gems        ?? 0);
}
// region CountDown
function SummontaskCountDown(task){

    const placed_building = PlacedBuildings.find((building) => building.id === task.placed_building_id)

    task.building=placed_building

    const durationSeconds = task.duration_seconds;
    const countdown = createCountdownSprite(durationSeconds*1000,new Date(task.started_at).getTime(),()=>{
        SendToServer({action:"CHECK_CONSTRUCTION_WORK",access_token,message:""})
    });
    countdown.position.copy(placed_building.Model.position);
    countdown.position.y += AllBuildingData[placed_building.building_id].grid_size_x * size_scaling + 0.6;

    scene.add(countdown);
    placed_building.Model.userData.countdownSprite = countdown;

}
function createCountdownSprite(durationMilliSeconds,started_at,OnDone) {
    const canvas = document.createElement('canvas');
    canvas.width = 256;
    canvas.height = 128;

    const texture = new THREE.CanvasTexture(canvas);

    const geometry = new THREE.PlaneGeometry(1.5*size_scaling, size_scaling*0.75); // world-space size
    const material = new THREE.MeshBasicMaterial({
        map: texture,
        transparent: true,
        depthWrite: false,
        depthTest:false,
        side: THREE.DoubleSide
    });

    const sprite = new THREE.Mesh(geometry, material);

    const state = {
        endTime: started_at + durationMilliSeconds,
        canvas,
        texture,
        done: false,
        OnDone,
        durationMilliSeconds
    };
    drawCountdown(state);
    sprite.userData.countdown = state;
    activeCountdowns.add(sprite)
    return sprite;
}

function drawCountdown(state) {
    const ctx = state.canvas.getContext('2d');
    const { canvas } = state;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    const remaining = Math.max(0, Math.ceil((state.endTime - Date.now()) / 1000));

    const pad = 16;
    ctx.fillStyle = 'rgba(20, 20, 30, 0.72)';
    roundRect(ctx, pad, pad, canvas.width - pad * 2, canvas.height - pad * 2, 20);
    ctx.fill();

    const barX = 28, barY = canvas.height - 36, barW = canvas.width - 56, barH = 8;
    ctx.fillStyle = 'rgba(255,255,255,0.15)';
    roundRect(ctx, barX, barY, barW, barH, 4);
    ctx.fill();

     ctx.fillStyle = remaining > 60 ? '#4ade80' : remaining > 20 ? '#fbbf24' : '#f87171';
    roundRect(ctx, barX, barY, Math.max(8, barW * (remaining / state.durationMilliSeconds)), barH, 4);
    ctx.fill();

    ctx.fillStyle = '#ffffff';
    ctx.font = 'bold 44px sans-serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(formatTime(remaining), canvas.width / 2, canvas.height / 2 - 8);

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

function tickCountdowns(scene, camera) {
    if (!activeCountdowns){
        return;
    }
    for (const sprite of activeCountdowns) {
        const cd = sprite.userData.countdown;

        sprite.rotation.y = Math.atan2(
            camera.position.x - sprite.position.x,
            camera.position.z - sprite.position.z
        );

        if (cd.done) {
            cd.OnDone()
            activeCountdowns.delete(sprite)
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

// endregion

// region Click
const raycaster = new THREE.Raycaster();
const mouse = new THREE.Vector2();
const geometry = new THREE.BoxGeometry(1, 1, 1);
const material = new THREE.MeshNormalMaterial();
const meshCUbe = new THREE.Mesh(geometry, material);
scene.add(meshCUbe);

window.addEventListener('mousemove',(event)=>{
    mouse.x = (event.clientX / window.innerWidth) * 2 - 1;
    mouse.y = -(event.clientY / window.innerHeight) * 2 + 1;
    raycaster.setFromCamera(mouse, camera);
    const intersects = raycaster.intersectObject(ground);
    const hit = intersects[0];
    const worldPos = hit.point;
    const gridX    = Math.round(worldPos.x / position_scaling);
    const gridY    = Math.round(worldPos.z / position_scaling);
    meshCUbe.position.set(gridX*position_scaling,0,gridY*position_scaling)
})
function onMouseClick(event) {
    if (!inBattle) {
        const bmOverlay = document.getElementById('bm-overlay');
        const shopOverlay = document.getElementById('shop-overlay');
        if (bmOverlay.classList.contains('is-active') || shopOverlay.classList.contains('is-active')) return;

        const rect = canvas.getBoundingClientRect();
        mouse.x =  ((event.clientX - rect.left) / rect.width)  *  2 - 1;
        mouse.y = -((event.clientY - rect.top)  / rect.height) *  2 + 1;

        raycaster.setFromCamera(mouse, camera);
        const intersects = raycaster.intersectObject(ground);
        if (!intersects.length) return;

        const worldPos = intersects[0].point;
        const gridX = Math.round(worldPos.x / position_scaling);
        const gridY = Math.round(worldPos.z / position_scaling);

        if (Grid[[gridX, gridY]]) {
            triggerBuildingMenu(Grid[[gridX, gridY]].userData);
        } else {
            openBuildingShop(gridX, gridY);
        }
    }
}
window.addEventListener('click', onMouseClick);
// region AI generated part
// region Troop Training

let _troopSelectedId    = null;
let _troopSelectedLevel = 1;
let _troopCount         = 1;

const trainBtn = document.getElementById('bm-train-btn');
function updateTrainButton(building) {
    trainBtn.style.display = (building.category === BuildingCategory.Army && !_placed_building.is_broken)? 'block' : 'none';
}
const collect = document.getElementById('bm-collect-btn');
function updateCollectButton() {
    let data
    let building
    let levelDetails
    try {
         data = _placed_building
         building = AllBuildingData[data.building_id];
         levelDetails = building.levels[data.level];
    }
    catch (e){
        return
    }
    if (building.category === BuildingCategory.Resource && building.levels[1].generation_rate > 0 && !data.is_broken) {
        collect.style.display = 'block'
        let toCollect = Math.min(levelDetails.storage_capacity, levelDetails.generation_rate * (Date.now() - new Date(_placed_building.last_updated_at).getTime()) / 1000 / 3600);
        if (building.resource_type === ResourceType.Gold) {
            if (UserData.total_gold_capacity === UserData.current_gold) {
                collect.textContent = "Storage Full"
                setAffordability(collect, false);
            } else {
                collect.textContent = `✨️ Collect ${formatNum(Math.min(toCollect, UserData.total_gold_capacity - UserData.current_gold))}`
                setAffordability(collect, true);
            }
        } else if (building.resource_type === ResourceType.Elixir) {
            if (UserData.total_elixir_capacity === UserData.current_elixir) {
                collect.textContent = "Storage Full"
                setAffordability(collect, false);
            } else {
                collect.textContent = `✨️ Collect ${formatNum(Math.min(toCollect, UserData.total_elixir_capacity - UserData.current_elixir))}`
                setAffordability(collect, true);
            }
        } else if (building.resource_type === ResourceType.DarkElixir) {
            if (UserData.total_dark_elixir_capacity === UserData.current_dark_elixir) {
                collect.textContent = "Storage Full"
                setAffordability(collect, false);
            } else {
                collect.textContent = `✨️ Collect ${formatNum(Math.min(toCollect, UserData.total_dark_elixir_capacity - UserData.current_dark_elixir))}`
                setAffordability(collect, true);
            }
        }
    } else {
        collect.style.display = 'none'
    }


}
document.getElementById('bm-train-btn').onclick = (e) => {
    e.stopPropagation();
    document.getElementById('bm-overlay').classList.remove('is-active');
    openTroopTraining(_placed_building_id);
};
collect.onclick = (e)=>{
    e.stopPropagation()
    SendToServer({action:"COLLECT_RESOURCE",access_token,message:JSON.stringify({
            placed_building_id:_placed_building_id
        })});
    _placed_building.last_updated_at = Date.now()
};
function openTroopTraining(placed_building_id) {
    _troopSelectedId    = null;
    _troopSelectedLevel = 1;
    _troopCount         = 1;
    showTroopGrid(placed_building_id);
    document.getElementById('troop-overlay').classList.add('is-active');
}

function closeTroopOverlay() {
    document.getElementById('troop-overlay').classList.remove('is-active');
}

document.getElementById('troop-close-btn').onclick       = (e) => { e.stopPropagation(); closeTroopOverlay(); };
document.getElementById('troop-detail-close-btn').onclick = (e) => { e.stopPropagation(); closeTroopOverlay(); };
document.getElementById('troop-overlay').onclick = (e) => {
    e.stopPropagation();
    if (e.target === document.getElementById('troop-overlay')) closeTroopOverlay();
};
document.getElementById('troop-detail-back').onclick = (e) => { e.stopPropagation(); showTroopGrid(); };

// ── Grid view ────────────────────────────────────────

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
        const realLevels = troop.level_stats.length - 1; // index 0 is dummy

        // Sum troops across all real levels
        let totalOwned = 0;
        for (let lv = 1; lv <= realLevels; lv++) {
            totalOwned += TrainedTroopsData[[troopId, lv]] ?? 0;
        }

        const node  = tpl.content.cloneNode(true);
        const card  = node.querySelector('.troop-card');
        const img   = node.querySelector('.troop-card-img');
        const badge = node.querySelector('.troop-card-badge');
        const name  = node.querySelector('.troop-card-name');
        const sub   = node.querySelector('.troop-card-sub');

        card.dataset.id = troopId;
        img.src         = `./Models/${escapeHTML(troop.name)}.png`;
        img.alt         = troop.name;
        img.onerror     = function() {
            this.style.display = 'none';
            this.nextElementSibling.style.display = 'flex';
        };
        name.textContent = troop.name;
        sub.textContent  = `Lv 1–${realLevels}`;

        if (totalOwned > 0) {
            badge.textContent    = totalOwned;
            badge.style.display  = 'flex';
        }

        card.addEventListener('click', (e) => {
            e.stopPropagation();
            showTroopDetail(troopId,placed_building_id);
        });

        grid.appendChild(node);
    });
}

// ── Detail view ──────────────────────────────────────

function showTroopDetail(troopId,placed_building_id) {
    _troopSelectedId    = troopId;
    _troopSelectedLevel = 1;
    _troopCount         = 1;

    document.getElementById('troop-grid-view').style.display   = 'none';
    document.getElementById('troop-detail-view').style.display = 'block';

    renderTroopSidebar(troopId,placed_building_id);
    renderTroopDetailPane(troopId,placed_building_id);
}

function renderTroopSidebar(activeTroopId,placed_building_id) {
    const sidebar = document.getElementById('troop-sidebar');
    const tpl     = document.getElementById('troop-sidebar-tpl');
    sidebar.innerHTML = '';

    Object.entries(AllTroopsData).forEach(([troopId, troop]) => {
        const node = tpl.content.cloneNode(true);
        const item = node.querySelector('.troop-sidebar-item');
        const img  = node.querySelector('.troop-sidebar-img');
        const icon = node.querySelector('.troop-sidebar-icon');
        const name = node.querySelector('.troop-sidebar-name');

        item.dataset.id = troopId;
        if (troopId === activeTroopId) item.classList.add('active');

        img.src   = `./Models/${escapeHTML(troop.name)}.png`;
        img.alt   = troop.name;
        img.onerror = function() {
            this.style.display = 'none';
            icon.style.display = 'flex';
        };
        name.textContent = troop.name;

        item.addEventListener('click', (e) => {
            e.stopPropagation();
            _troopSelectedLevel = 1;
            _troopCount         = 1;
            _troopSelectedId    = troopId;
            sidebar.querySelectorAll('.troop-sidebar-item').forEach(i => i.classList.remove('active'));
            item.classList.add('active');
            renderTroopDetailPane(troopId,placed_building_id);
        });

        sidebar.appendChild(node);
    });
}

function renderTroopDetailPane(troopId,placed_building_id) {
    const troop      = AllTroopsData[troopId];
    const realLevels = troop.level_stats.length - 1; // index 0 is dummy

    document.getElementById('troop-detail-name').textContent = troop.name;
    document.getElementById('troop-pane-name').textContent   = troop.name;
    document.getElementById('troop-pane-meta').textContent   =
        `${troop.attack_type} · ${troop.preferred_target} target · ${troop.housing_space} housing`;

    // Level tabs
    const tabsEl = document.getElementById('troop-level-tabs');
    const tabTpl = document.getElementById('troop-level-tab-tpl');
    tabsEl.innerHTML = '';

    for (let lv = 1; lv <= realLevels; lv++) {
        const owned = TrainedTroopsData[[troopId, lv]] ?? 0;
        const node  = tabTpl.content.cloneNode(true);
        const tab   = node.querySelector('.troop-level-tab');
        const lvEl  = node.querySelector('.troop-tab-lv');
        const cntEl = node.querySelector('.troop-tab-count');

        tab.dataset.lv = lv;
        lvEl.textContent = `Lv ${lv}`;

        if (owned > 0) {
            cntEl.textContent    = `×${owned}`;
            cntEl.style.display  = 'block';
        }

        if (lv === _troopSelectedLevel) tab.classList.add('active');

        tab.addEventListener('click', (e) => {
            e.stopPropagation();
            _troopSelectedLevel = lv;
            _troopCount         = 1;
            // update active tab highlight without full re-render
            tabsEl.querySelectorAll('.troop-level-tab').forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            refreshTroopDetailPane(troopId,placed_building_id);
        });

        tabsEl.appendChild(node);
    }

    refreshTroopDetailPane(troopId,placed_building_id);
}

function refreshTroopDetailPane(troopId,placed_building_id) {
    const troop      = AllTroopsData[troopId];
    const lv         = _troopSelectedLevel;
    const stats      = troop.level_stats[lv];         // index 0 is dummy, so lv maps directly
    const upgCost    = troop.upgrade_costs[lv - 1];   // upgrade_costs[0] = train to Lv1
    const isNewTrain = lv === 1;
    document.getElementById('troop-owned-val').textContent = TrainedTroopsData[[troopId, lv]] ?? 0;
    // Stats grid — update existing elements by textContent only
    const statDefs = [
        ['HP',         (stats.health             ?? 0).toLocaleString()],
        ['Dmg / shot', stats.damage_per_shot      ?? '—'],
        ['Move speed', troop.movement_speed        ?? '—'],
        ['Atk speed',  troop.attack_speed_seconds ? `${troop.attack_speed_seconds}s` : '—'],
        ['Range',      troop.attack_range         ? `${troop.attack_range} tiles`    : '—'],
        ['Housing',    troop.housing_space         ?? '—'],
    ];
    const statsGrid = document.getElementById('troop-stats-grid');
    // Build once if empty, then only update values on subsequent calls
    if (statsGrid.children.length !== statDefs.length) {
        statsGrid.innerHTML = statDefs.map(([key, val]) =>
            `<div class="bm-stat">
                <p class="bm-stat-val">${escapeHTML(String(val))}</p>
                <p class="bm-stat-key">${escapeHTML(key)}</p>
             </div>`
        ).join('');
    } else {
        statsGrid.querySelectorAll('.bm-stat').forEach((el, i) => {
            el.querySelector('.bm-stat-val').textContent = String(statDefs[i][1]);
        });
    }

    // TH requirement
    const thTag = document.getElementById('troop-th-req');
    if (upgCost?.town_hall_level_required) {
        thTag.textContent    = `TH ${upgCost.town_hall_level_required} required`;
        thTag.style.display  = 'inline-block';
    } else {
        thTag.style.display  = 'none';
    }

    // Requires note (upgrade only)
    const noteEl = document.getElementById('troop-requires-note');
    if (!isNewTrain) {
        const prevOwned = TrainedTroopsData[[troopId, lv - 1]] ?? 0;
        noteEl.textContent  = `⚠ Requires ${_troopCount}× Lv ${lv - 1} — you have ${prevOwned}`;
        noteEl.style.display = 'block';
    } else {
        noteEl.style.display = 'none';
    }

    // Count controls
    document.getElementById('troop-count-dec').onclick = (e) => {
        e.stopPropagation();
        if (_troopCount > 1) { _troopCount--; updateTroopCosts(troopId,placed_building_id); }
    };
    document.getElementById('troop-count-inc').onclick = (e) => {
        e.stopPropagation();
        _troopCount++;
        updateTroopCosts(troopId,placed_building_id);
    };

    // Train / gem buttons
    const trainBtn = document.getElementById('troop-train-btn');
    const gemBtn   = document.getElementById('troop-gem-btn');

    trainBtn.onclick = (e) => {
        e.stopPropagation();
        if (!canAffordTroop(upgCost, false)) return;
        closeTroopOverlay();
        TrainTroop(troopId, _troopCount, placed_building_id,lv,false);
        UserData.current_gold-=upgCost.gold_required
        UserData.current_elixir-=upgCost.elxir_required
        UserData.current_dark_elixir-=upgCost.dark_elixir_required
        if (lv!==0){
            TrainedTroopsData[[troop,lv]]-=_troopCount;
        }
    };
    gemBtn.onclick = (e) => {
        e.stopPropagation();
        if (!canAffordTroop(upgCost, true)) return;
        closeTroopOverlay();
        TrainTroop(troopId, _troopCount, placed_building_id,lv,true);
        UserData.current_gems-=upgCost.or_gem_required
        if (lv!==0){
            TrainedTroopsData[[troop,lv]]-=_troopCount;
        }
    };

    updateTroopCosts(troopId,placed_building_id);
}

function updateTroopCosts(troopId,placed_building_id) {
    const troop   = AllTroopsData[troopId];
    const lv      = _troopSelectedLevel;
    const upgCost = troop.upgrade_costs[lv - 1];
    const n       = _troopCount;
    const isNewTrain = lv === 1;
    document.getElementById('troop-count-val').textContent = n;
    document.getElementById('troop-owned-val').textContent = TrainedTroopsData[[troopId, _troopSelectedLevel]] ?? 0;

    // Requires note count update
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
        scaled.map(c =>
            `<div class="bm-cost-row">
                <span class="bm-cost-label">${c.icon} ${c.label}</span>
                <span class="bm-cost-val">${formatNum(c.val)}</span>
             </div>`
        ).join('') +
        `<div class="bm-divider"></div>
         <div class="bm-time-row">⏱ ${formatTime(timeSec)} total time</div>`;

    const trainBtn = document.getElementById('troop-train-btn');
    const gemBtn   = document.getElementById('troop-gem-btn');

    trainBtn.textContent = isNewTrain
        ? `Train ${n}× ${troop.name} (Lv 1)`
        : `Upgrade ${n}× to Lv ${lv}`;

    const gemCost = (upgCost?.or_gem_required ?? 0) * n;
    gemBtn.textContent = `💎 Instant (${formatNum(gemCost)})`;

    // Disable if upgrade count exceeds owned Lv(lv-1)
    const prevOwned   = isNewTrain ? Infinity : (TrainedTroopsData[[troopId, lv - 1]] ?? 0);
    const enoughTroops = isNewTrain || n <= prevOwned;

    setAffordability(trainBtn, canAffordTroop(upgCost, false,placed_building_id) && enoughTroops);
    setAffordability(gemBtn,   canAffordTroop(upgCost, true,placed_building_id)  && enoughTroops);
}

// endregion

// region Shop – state
let _shopGridX = 0;
let _shopGridY = 0;
let _shopSelectedId = null;
let _shopActiveFilter = 'All';
let _shopSearchQuery = '';
let _placed_building_id = '';
let _placed_building = null;
// endregion

// region Shop – open / close
function openBuildingShop(gridX, gridY) {
    _shopGridX = gridX;
    _shopGridY = gridY;
    _shopSelectedId = null;

    document.getElementById('shop-grid-sub').textContent =
        `Grid (${gridX}, ${gridY}) · choose a building`;

    // Reset to list view
    showShopList();

    const overlay = document.getElementById('shop-overlay');
    overlay.classList.add('is-active');
}

function closeShop() {
    document.getElementById('shop-overlay').classList.remove('is-active');
}

document.getElementById('shop-close-btn').onclick = closeShop;
document.getElementById('shop-overlay').onclick = (e) => {
    e.stopPropagation()
    if (e.target === document.getElementById('shop-overlay')) closeShop();
};
// endregion

// region Shop – list view
function showShopList() {
    document.getElementById('shop-list').style.display = 'flex';
    document.getElementById('shop-detail').style.display = 'none';
    renderShopCards();
}

function renderShopCards() {
    const list = document.getElementById('shop-list');
    const query = _shopSearchQuery.toLowerCase();

    const entries = Object.entries(AllBuildingData).filter(([id, b]) => {
        const matchCat = _shopActiveFilter === 'All' || b.category === _shopActiveFilter;
        const matchQ   = !query || b.name.toLowerCase().includes(query);
        return matchCat && matchQ;
    });

    if (entries.length === 0) {
        list.innerHTML = `<p style="text-align:center;color:#aaa;padding:24px 0;font-size:13px;">No buildings found</p>`;
        return;
    }

    list.innerHTML = entries.map(([building_id, building]) => {
        const level1 = building.levels[1];
        const costs  = level1 ? getConstructionCostPills(level1) : [];
        const imgSrc = `./Models/${building.name}.png`;
        const catIcon = {
            defense:  '🛡️',
            resource: '💰',
            army:     '⚔️',
            townhall: '🏰',
        }[building.category] ?? '🏠';

        const imgHtml = `
            <img  class="shop-card-img"
                  src="${escapeHTML(imgSrc)}"
                  alt="${escapeHTML(building.name)}"
                  onerror="this.style.display='none';this.nextElementSibling.style.display='flex';" />
            <div  class="shop-card-img-fallback" style="display:none;">${catIcon}</div>`;

        const pillsHtml = costs.map(c =>
            `<span class="shop-card-cost-pill">${c}</span>`
        ).join('');

        return `
        <div class="shop-card" data-id="${escapeHTML(building_id)}">
            ${imgHtml}
            <div class="shop-card-info">
                <p class="shop-card-name">${escapeHTML(building.name)}</p>
                <p class="shop-card-cat">${building.category}</p>
                <div class="shop-card-costs">${pillsHtml}</div>
            </div>
            <span class="shop-card-arrow">›</span>
        </div>`;
    }).join('');

    list.querySelectorAll('.shop-card').forEach(card => {
        card.addEventListener('click', (e) => {
            e.stopPropagation()
            showShopDetail(card.dataset.id);
        });
    });
}

function getConstructionCostPills(level) {
    const pills = [];
    if (level.update_gold_required        > 0) pills.push(`🪙 ${formatNum(level.update_gold_required)}`);
    if (level.update_elxir_required       > 0) pills.push(`🧪 ${formatNum(level.update_elxir_required)}`);
    if (level.update_dark_elixir_required > 0) pills.push(`⚗️ ${formatNum(level.update_dark_elixir_required)}`);
    if (pills.length === 0 && level.update_or_gem_required > 0) {
        pills.push(`💎 ${formatNum(level.update_or_gem_required)}`);
    }
    return pills;
}
// endregion

// region Shop – detail view
function showShopDetail(building_id) {
    _shopSelectedId = building_id;
    const building  = AllBuildingData[building_id];
    const level1    = building.levels[1];

    document.getElementById('shop-detail-name').textContent = building.name;
    document.getElementById('shop-detail-cat').textContent  = building.category;

    const img = document.getElementById('shop-detail-img');
    img.src = `./Models/${building.name}.png`;
    img.alt = building.name;

    // Level-1 stats
    const stats = getBuildingStats(building, level1);
    document.getElementById('shop-detail-stats').innerHTML = stats.map(([key, val]) =>
        `<div class="bm-stat">
            <p class="bm-stat-val">${escapeHTML(val)}</p>
            <p class="bm-stat-key">${escapeHTML(key)}</p>
         </div>`
    ).join('');

    // Build costs (reuse existing helper – it generates cost rows + time)
    document.getElementById('shop-detail-costs').innerHTML = buildCostRows(level1);

    // TH requirement tag
    const thTag = document.getElementById('shop-th-req');
    if (level1.update_townhall_level_required) {
        thTag.textContent = `TH ${level1.update_townhall_level_required} required`;
        thTag.style.display = 'inline-block';
    } else {
        thTag.style.display = 'none';
    }
    const gemCost = level1.update_or_gem_required ?? 0;
    document.getElementById('shop-gem-btn').textContent = `💎 Instant (${formatNum(gemCost)})`;

    setAffordability(document.getElementById('shop-build-btn'), canAfford(level1, false));
    setAffordability(document.getElementById('shop-gem-btn'),   canAfford(level1, true));

    document.getElementById('shop-list').style.display = 'none';
    document.getElementById('shop-detail').style.display = 'flex';
}

document.getElementById('shop-detail-back').onclick = showShopList;

document.getElementById('shop-build-btn').onclick = (e) => {
    e.stopPropagation()
    if (_shopSelectedId === null) return;
    closeShop();
    CreateBuilding(_shopSelectedId, _shopGridX, _shopGridY, false);
    const levelDetails = AllBuildingData[_shopSelectedId].levels[1]
    UserData.current_gold-=levelDetails.update_gold_required
    UserData.current_elixir-=levelDetails.update_elxir_required
    UserData.current_dark_elixir-=levelDetails.update_dark_elixir_required
    UpdateResourceUI()
};

document.getElementById('shop-gem-btn').onclick = (e) => {
    e.stopPropagation()
    if (_shopSelectedId === null) return;
    closeShop();
    CreateBuilding(_shopSelectedId, _shopGridX, _shopGridY, true);
    const levelDetails = AllBuildingData[_shopSelectedId].levels[1]
    UserData.current_gems-=levelDetails.update_or_gem_required
    UpdateResourceUI()
};
// endregion

// region Shop – filter tabs & search
document.getElementById('shop-filter-tabs').addEventListener('click', (e) => {
    const tab = e.target.closest('.shop-tab');
    if (!tab) return;
    document.querySelectorAll('.shop-tab').forEach(t => t.classList.remove('active'));
    tab.classList.add('active');
    _shopActiveFilter = tab.dataset.cat;
    renderShopCards();
});

document.getElementById('shop-search').addEventListener('input', (e) => {
    _shopSearchQuery = e.target.value;
    renderShopCards();
});
// endregion

// region Affordability

function canAffordTroop(upgCost, checkGems = false,placed_building_id) {
    if (!upgCost) return false;
    const alreading = !(placed_building_id!=null && ConstructionTasks.some(task=>task.placed_building_id===placed_building_id))
    const townhall_Level = UserData.town_hall_level>=upgCost.town_hall_level_required
    const n = _troopCount;
    if (checkGems) return (UserData.current_gems ?? 0) >= (upgCost.or_gem_required ?? 0) * n;
    return alreading&&townhall_Level&&(UserData.current_gold        ?? 0) >= (upgCost.gold_required        ?? 0) * n
        && (UserData.current_elixir      ?? 0) >= (upgCost.elixir_required      ?? 0) * n
        && (UserData.current_dark_elixir ?? 0) >= (upgCost.dark_elixir_required ?? 0) * n;
}

function canAfford(level, checkGems = false,placed_building_id = null,is_broken = false) {
    const alreading = !(placed_building_id!=null && ConstructionTasks.some(task=>task.placed_building_id===placed_building_id))
    const townhall_Level = is_broken?true:(UserData.town_hall_level>=level.update_townhall_level_required)

    if (checkGems) {
        return alreading&&townhall_Level&&(UserData.current_gems ?? 0) >= (is_broken?1+level.update_or_gem_required /10 : level.update_or_gem_required);
    }
    const goldOk      = (UserData.current_gold       ?? 0) >= (is_broken?level.update_gold_required/10:level.update_gold_required);
    const elixirOk    = (UserData.current_elixir     ?? 0) >= (is_broken?level.update_elxir_required/10:level.update_elxir_required);
    const darkElixirOk= (UserData.current_dark_elixir?? 0) >= (is_broken?level.update_dark_elixir_required /10 : level.update_dark_elixir_required);
    return alreading&&townhall_Level&&goldOk && elixirOk && darkElixirOk;
}


function setAffordability(btn, affordable) {
    if (!btn) return;
    btn.disabled = !affordable;
    btn.classList.toggle('btn-unaffordable', !affordable);
}

// endregion

// region Building Menu

function escapeHTML(str) {
    if (str === null || str === undefined) return '';
    return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

function triggerBuildingMenu(data) {
    _placed_building_id = data.id
    _placed_building = data
    const building = AllBuildingData[data.building_id];
    const levelDetails = building.levels[data.level];
    if (data.is_broken){

    }
    updateTrainButton(building)
    updateCollectButton(building,data,levelDetails)

    const maxLevel = 6;
    const nextLevel = data.level + 1;
    const isMaxLevel = data.level >= maxLevel;

    const categoryIcon = {
        Defense:  'ti-shield',
        Resource: 'ti-database',
        Army:     'ti-sword',
        TownHall: 'ti-building-castle',
    }[building.category] ?? 'ti-home';

    document.getElementById('bm-icon-class').className = `ti ${categoryIcon}`;
    document.getElementById('bm-name-text').textContent = building.name;
    document.getElementById('bm-sub-text').textContent = `${building.category} · Level ${data.level}`;

    document.getElementById('bm-level-dots').innerHTML = Array.from({length: maxLevel}, (_, i) =>
        `<span class="bm-dot${i < data.level ? ' filled' : ''}"></span>`
    ).join('');
    document.getElementById('bm-level-text').textContent = `${data.level} / ${maxLevel}`;
    document.getElementById('bm-stats-grid').innerHTML = getBuildingStats(building, levelDetails).map(([key, val]) =>
        `<div class="bm-stat"><p class="bm-stat-val">${escapeHTML(val)}</p><p class="bm-stat-key">${escapeHTML(key)}</p></div>`
    ).join('');

    const upgradeSection = document.getElementById('bm-upgrade-section');
    const maxNotice = document.getElementById('bm-max-notice');
    const upgradeBtn = document.getElementById('bm-upgrade-btn');
    const gemUpgradeBtn = document.getElementById('bm-gem-upgrade-btn');
    const gemRepairBtn = document.getElementById('bm-gem-repair-btn')
    const repairBtn = document.getElementById('bm-repair-btn')

    gemRepairBtn.style.display = data.is_broken?'block':'none'
    repairBtn.style.display = data.is_broken?'block':'none'
    gemUpgradeBtn.style.display = data.is_broken ? 'none':'block'
    upgradeBtn.style.display = data.is_broken ? 'none':'block'
    if (isMaxLevel && !data.is_broken) {
        upgradeSection.style.display = 'none';
        maxNotice.style.display = 'block';
        upgradeBtn.disabled = true;
        if (gemUpgradeBtn) gemUpgradeBtn.disabled = true;
    }
    else {
        upgradeSection.style.display = 'block';
        maxNotice.style.display = 'none';

        document.getElementById('bm-upgrade-title-text').textContent = data.is_broken?`Repair to level ${nextLevel-1}`:`Upgrade to level ${nextLevel}`;
        document.getElementById('bm-upgrade-title').textContent = data.is_broken? "Repair cost":"Upgrade cost"
        const thTag = document.getElementById('bm-th-req-tag');
        if (levelDetails.update_townhall_level_required && !data.is_broken) {
            thTag.textContent = `TH ${levelDetails.update_townhall_level_required} required`;
            thTag.style.display = 'inline-block';
        } else {
            thTag.style.display = 'none';
        }

        document.getElementById('bm-costs-container').innerHTML = buildCostRows(levelDetails,data.is_broken);

        const affordable    = canAfford(levelDetails, false,data.id,data.is_broken);
        const gemAffordable = canAfford(levelDetails, true,data.id,data.is_broken);

        setAffordability(data.is_broken?repairBtn:upgradeBtn,    affordable);
        setAffordability(data.is_broken?gemRepairBtn:gemUpgradeBtn, gemAffordable);

        if (gemUpgradeBtn) {
            let gemCost = levelDetails.update_or_gem_required ?? 0;
            if (data.is_broken) gemCost = 1 + gemCost/10
            if (data.is_broken ){gemRepairBtn.textContent = `💎 Instant (${formatNum(gemCost)})`}else {gemUpgradeBtn.textContent = `💎 Instant (${formatNum(gemCost)})`;}
        }
    }

    const overlay = document.getElementById('bm-overlay');
    overlay.classList.add('is-active');
    document.getElementById('bm-close-btn').onclick = (e) =>{e.stopPropagation(); overlay.classList.remove('is-active');}
    overlay.onclick = (e) => {e.stopPropagation(); if (e.target === overlay) overlay.classList.remove('is-active');};

    upgradeBtn.onclick = (e) => {
        e.stopPropagation()
        if (!isMaxLevel) {
            overlay.classList.remove('is-active');
            UpgradeBuilding(data.id,false)
            UserData.current_gold-=levelDetails.update_gold_required
            UserData.current_elixir-=levelDetails.update_elxir_required
            UserData.current_dark_elixir-=levelDetails.update_dark_elixir_required
            UpdateResourceUI()
        }
    };

    if (gemUpgradeBtn) {
        gemUpgradeBtn.onclick = (e) => {
            e.stopPropagation()
            if (!isMaxLevel) {
                overlay.classList.remove('is-active');

                UpgradeBuilding(data.id,true)
                UserData.current_gems-=levelDetails.update_or_gem_required
                UpdateResourceUI()
            }
        };
    }
    repairBtn.onclick = (e) => {
        e.stopPropagation()
        overlay.classList.remove('is-active');
        RepairBuilding(data.id,false)
        UserData.current_gold-=levelDetails.update_gold_required/10
        UserData.current_elixir-=levelDetails.update_elxir_required/10
        UserData.current_dark_elixir-=levelDetails.update_dark_elixir_required/10
        UpdateResourceUI()

    };

    if (gemUpgradeBtn) {
        gemRepairBtn.onclick = (e) => {
            e.stopPropagation()
            if (!isMaxLevel) {
                overlay.classList.remove('is-active');
                RepairBuilding(data.id,true)
                UserData.current_gems-=1+levelDetails.update_or_gem_required/10
                UpdateResourceUI()
            }
        };
    }
}

function getBuildingStats(building, level) {

    const base = [['HP', (level.health ?? 0).toLocaleString()]];

    switch (building.category) {
        case BuildingCategory.Defense:
            return [
                ...base,
                ['Damage / shot',  level.damage_per_shot  ?? '—'],
                ['Range',          building.attack_range  ? `${building.attack_range} tiles` : '—'],
                ['Attack speed',   building.attack_speed_seconds ? `${building.attack_speed_seconds}s` : '—'],
                ['Damage type',    building.damage_type   ?? '—'],
                ['Targets',        building.unit_target   ?? '—'],
            ];
        case BuildingCategory.Resource:
            return [
                ...base,
                ['Gen rate / hr',  level.generation_rate  ? formatNum(level.generation_rate) : '—'],
                ['Capacity',       level.storage_capacity ? formatNum(level.storage_capacity) : '—'],
                ['Resource',       building.resource_type ?? '—'],
            ];
        case BuildingCategory.Army:
            return [
                ...base,
                ['Troop capacity', level.troop_capacity ?? '—'],
            ];
        default:
            return base;
    }
}
function buildCostRows(level,is_broken = false) {
    const costs = [
        { label: 'Gold',        icon: '🪙', val: is_broken?level.update_gold_required/10:level.update_gold_required },
        { label: 'Elixir',      icon: '🧪', val: is_broken?level.update_elxir_required/10:level.update_elxir_required },
        { label: 'Dark elixir', icon: '⚗️',  val: is_broken?level.update_dark_elixir_required/10:level.update_dark_elixir_required },
    ].filter(c => c.val > 0);

    const rows = costs.map(c =>
        `<div class="bm-cost-row">
            <span class="bm-cost-label">${c.icon} ${c.label}</span>
            <span class="bm-cost-val">${formatNum(c.val)}</span>
         </div>`
    ).join('');

    return rows + `
        <div class="bm-divider"></div>
        <div class="bm-time-row">⏱ ${formatTime(level.update_time_required_required/10)} build time</div>`;
}
function formatNum(n) {
    if (!n) return '0';
    n = Math.floor(n)
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000)      return (n / 1_000).toFixed(0) + 'K';
    return n.toString();
}
function formatTime(seconds) {
    if (!seconds) return '—';
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    if (h > 0) return `${h}h ${m}m`;
    if (m > 0) return `${m}m ${s}s`;
    return `${s}s`;
}
// endregion

// endregion

// endregion

// region Network
var access_token = localStorage.getItem('access_token');

async function refreshAuthToken() {
    const userId = UserData.user_id;
    const refreshToken = localStorage.getItem('refresh_token_b64');

    try {
        const response = await fetch('/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user_id: userId, refresh_token: refreshToken })
        });

        if (!response.ok) throw new Error('Session expired');

        const data = await response.json();

        localStorage.setItem('access_token', data.access_token);
        access_token = data.access_token
        localStorage.setItem('refresh_token_b64',data.refresh_token_b64)
    } catch (error) {
        console.error("Refresh failed, redirecting to login:", error);
        window.location.href = '/Login.html';
    }
}

const REFRESH_INTERVAL = 14 * 60 * 1000; // 14 minutes
setInterval(refreshAuthToken, REFRESH_INTERVAL);

let socket
function connectToGameServer() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const socketUrl = `${protocol}//${host}/ws?token=${access_token}`;
    socket = new WebSocket(socketUrl);
    socket.addEventListener("open", (event) => {
        SendToServer({ action: "INITIAL_LOAD", message: "",access_token });
    });

    socket.addEventListener("message", (event) => {
        const data = JSON.parse(event.data)
        if (!inBattle) {
            switch (data['msg_type']) {
                case 'building_troop_of_user':
                    UserData = data.user_data
                    PlacedBuildings = data.building

                    localStorage.setItem('Placed_building', JSON.stringify(PlacedBuildings))
                    ConstructionTasks = data.construction_tasks
                    TrainedTroopsData = {}
                    for (const troop of data.troops) {
                        TrainedTroopsData[[troop.troop_id, troop.level]] = troop.count
                    }

                    localStorage.setItem('Trained_troops_data', JSON.stringify(TrainedTroopsData))
                    if (Object.keys(AllBuildingData).length !== 0 && Object.keys(AllTroopsData).length !== 0) {
                        LoadMap(PlacedBuildings)
                    } else {
                        SendToServer({action: 'ALL_BUILDING_TROOP_DATA', access_token, message: ""})
                    }
                    // TODO : troop related thing also
                    SendToServer({action: "CHECK_CONSTRUCTION_WORK", access_token, message: ""})
                    UpdateResourceUI()
                    break
                case 'building_troop':
                    const buildings = data.building
                    const defence = data.defence
                    const army = data.army
                    const resource = data.resource
                    const ParticularLevelData = data.particular_level_data

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
                    AddLevelDetails(ParticularLevelData)
                    localStorage.setItem('All_building_data', JSON.stringify(AllBuildingData))
                    if (PlacedBuildings.length !== 0) LoadMap(PlacedBuildings)
                    else SendToServer({action: "INITIAL_LOAD", message: "", access_token})
                    // TODO : troop related thing also
                    break
                case "building_level_detail":
                    break
                case 'construction_started':
                    ConstructionTasks.push(data.task)
                    if (data.placed_building != null) {
                        PlacedBuildings.push(data.placed_building)
                        LoadMap(PlacedBuildings)
                    } else {
                        SummontaskCountDown(data.task);
                    }
                    break
                case "construction_completed":
                    console.log(data.particular_level_detail)
                    AddLevelDetails(data.particular_level_detail)
                    const doneIds = new Set(data.construction_done.map(task => task.id));
                    const {kept, removed} = ConstructionTasks.reduce(
                        (acc, task) => {
                            if (doneIds.has(task.id)) {
                                acc.removed.push(task);
                            } else {
                                acc.kept.push(task);
                            }
                            return acc;
                        },
                        {kept: [], removed: []}
                    );
                    ConstructionTasks = kept;
                    for (const removedElement of removed) {
                        if (removedElement.task_type === ConstructionType.TroopTraining) {
                            if (TrainedTroopsData[[removedElement.troop_id, removedElement.troop_level_to]]) TrainedTroopsData[[removedElement.troop_id, removedElement.troop_level_to]] += removedElement.troop_count
                            else TrainedTroopsData[[removedElement.troop_id, removedElement.troop_level_to]]=removedElement.troop_count
                        }else if (removedElement.task_type===ConstructionType.BuildingRepair) {
                            PlacedBuildings.find(building => building.id === removedElement.placed_building_id).is_broken = false
                        } else{
                            PlacedBuildings.find(building => building.id === removedElement.placed_building_id).level += 1
                        }
                        localStorage.setItem('Trained_troops_data', JSON.stringify(TrainedTroopsData))
                        scene.remove(removedElement.building.Model.userData.countdownSprite);
                        removedElement.building.Model.userData.countdownSprite.material.dispose();
                        removedElement.building.Model.userData.countdownSprite.geometry.dispose();
                        delete removedElement.building.Model.userData.countdownSprite;
                    }
                    UserData = data.user_data
                    UpdateResourceUI()
                    LoadMap(PlacedBuildings)
                    break
                case "resource_collected":
                    UserData = data.user_data;
                    UpdateResourceUI()
                    break
                case "troop_training_started":
                    ConstructionTasks.push(data.task)
                    SummontaskCountDown(data.task);
                    break
                case "un_attack":
                    CancelMatchMaking()
                    break
                case "incoming_attack":
                    console.log("defend",data)
                    inBattle = true
                    IsAttacker = false
                    IncomingAttack()
                    state = {
                        Buildings:data.defender_building,
                        TroopSpawns:[],
                        AliveTroopAttacker:[],
                        AliveTroopDefender: [],
                        AliveBuildings:data.alive_buildings
                    }
                    LoadMap(data.defender_building)
                    FoundMatch()
                    SendToServer({action:"DEFEND",message:"",access_token})
                    break
                case "battle_start":
                    console.log(data)
                    state = {
                        Buildings:data.defender_building,
                        TroopSpawns:[],
                        AliveTroopAttacker:[],
                        AliveTroopDefender: [],
                        AliveBuildings:data.alive_buildings
                    }
                    LoadMap(data.defender_building)
                    inBattle = true
                    IsAttacker = true
                    FoundMatch()
                    break
                case "replay":
                    replay = true
                    inBattle = true
                    IsAttacker = true // replay happens as attacker ,when the server send opponent name and my name at the end of battle opponent name will be defender name and my name will be attacker name
                    state = {
                        Buildings:data.defender_building,
                        TroopSpawns:[],
                        AliveTroopAttacker:[],
                        AliveTroopDefender: [],
                        AliveBuildings:data.alive_buildings
                    }
                    LoadMap(data.defender_building)
                    attackBtn.classList.add('hidden');
                    break
            }
        }
        else {
            switch (data.msg_type){
                case "spawn_troop":
                    console.log("spawnTroop From server")
                    SpawnTroop(data.troop)
                    break
                case "battle_update":
                    DealDamage(data.building_damage,data.attacker_troop_damage,data.defender_troop_damage,data.building_died,data.attacker_troop_died,data.defender_troop_died)
                    break
                case "battle_over":
                    DespawnTroops()
                    BattleOver(data.battle_outcome,data.attacker_troop_loss,data.buildings_broken,{},data.opponent_usernamedata.my_username)
                    replay = false
                    _hideDeployBar()
                    inBattle = false
                    CancelMatchMaking()
                    SendToServer({action:"INITIAL_LOAD",access_token,message:""})
                    break
            }
        }
        if (data['status'] && data['status']==='error'){console.error(data['message'])}
    });

    socket.addEventListener("error", (error) => {
        console.error("WebSocket Error detected:", error);
    });

    socket.addEventListener("close", (event) => {
        console.warn("Disconnected from server.Reconecting..");
        // window.location.href = './Login.html'
        connectToGameServer();
    });

}
function CreateBuilding(building_id, x, y, use_gems=false){
    SendToServer({action:"CREATE_BUILDING",message:JSON.stringify({
            building_id:building_id,x,y, use_gems
        }),access_token})
}
function UpgradeBuilding(placed_building_id,use_gems=false){
    SendToServer({action:"UPGRADE_BUILDING",message:JSON.stringify({
            placed_building_id,use_gems
        }),access_token})
}
function RepairBuilding(placed_building_id,use_gems=false){
    SendToServer({action:"REPAIR_BUILDING",message:JSON.stringify({
            placed_building_id,use_gems
        }),access_token})
}
function TrainTroop(troop_id, count, barrack_placed_building_id   , level_to, use_gems) {
    console.log(`TrainTroop: troopId=${troop_id}, count=${count}, levelFrom=${level_to}`);
    SendToServer({action:"TRAIN_TROOP",access_token,message:JSON.stringify({barrack_placed_building_id,troop_id,count,use_gems,
            level_from: level_to-1})})
}
function SendToServer(dataObject) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(dataObject));
    } else {
        console.error("Cannot send action. WebSocket connection is not open.");
    }
}
connectToGameServer()
// endregion

// region Enums
const AttackType = {
    Melee  : "melee",
    Ranged : "ranged"
}
const BuildingCategory= {
    TownHall: "townhall",
    Defense: "defense",
    Resource: "resource",
    Army: "army",
    Wall: "wall"
}
const DamageType = {
    SingleTarget: "single_target",
    Splash: "splash"
}
const UnitTargetType = {
    Ground: "ground",
    GroundAndAir: "ground_and_air",
    Air: "air"
}
const ResourceType  = {
    Gold: "gold",
    Elixir: "elixir",
    DarkElixir: "dark_elixir"
}
const ConstructionType = {
    BuildingConstruction :"building_construction",
    BuildingUpgrade      :"building_upgrade",
    TroopTraining        :"troop_training",
    BuildingRepair : "building_repair"
}
// endregion

//region battle

var inBattle = false
var replay = false
var IsAttacker = false;
const attackBtn         = document.getElementById('attack-btn');
const incomingWarning   = document.getElementById('incoming-warning');
const matchmakingOverlay = document.getElementById('matchmaking-overlay');
const matchmakingTimer  = document.getElementById('matchmaking-timer');
const matchmakingCancel = document.getElementById('matchmaking-cancel-btn');
const deployBar         = document.getElementById('deploy-bar');
const deployTroopList   = document.getElementById('deploy-troop-list');

let _mmInterval        = null;   // matchmaking timer interval
let _mmSeconds         = 0;      // elapsed seconds
let _warningTimeout    = 5;   // auto-hide timeout for warning
let _selectedTroopId   = null;   // currently selected deploy troop id
let _selectedTroopLvl  = null;   // level of selected troop
let _deployMode        = false;  // true while waiting for ground click
let _troopButtons      = {};     // { `${id},${level}` : { btn, countEl, count } }

function _startMMTimer() {
    _mmSeconds = 0;
    matchmakingTimer.textContent = '0:00';
    clearInterval(_mmInterval);
    _mmInterval = setInterval(() => {
        _mmSeconds++;
        matchmakingTimer.textContent = formatTime(_mmSeconds);
    }, 1000);
}

function _stopMMTimer() {
    clearInterval(_mmInterval);
    _mmInterval = null;
}

function _showMatchmaking() {
    matchmakingOverlay.classList.add('is-active');
    _startMMTimer();
}

function _hideMatchmaking() {
    matchmakingOverlay.classList.remove('is-active');
    _stopMMTimer();
}


attackBtn.addEventListener('click', (e) => {
    e.stopPropagation()
    attackBtn.classList.add('hidden');
    _showMatchmaking();
    SendToServer({action:"ATTACK",access_token,message:""});
});

function CancelMatchMaking(e){
    _hideMatchmaking();
    attackBtn.classList.remove('hidden');
}
matchmakingCancel.addEventListener('click', CancelMatchMaking);

function FoundMatch() {
    _hideMatchmaking();
    _buildDeployBar();
    _showDeployBar();
    attackBtn.classList.add('hidden');
}

function IncomingAttack() {
    if (_warningTimeout) {
        clearTimeout(_warningTimeout);
        _warningTimeout = null;
    }

    incomingWarning.classList.add('is-active');

    _warningTimeout = setTimeout(() => {
        incomingWarning.classList.remove('is-active');
        _warningTimeout = null;
    }, 6000);
}


function _buildDeployBar() {
    deployTroopList.innerHTML = '';
    _troopButtons = {};
    _selectedTroopId  = null;
    _selectedTroopLvl = null;
    for (const key of Object.keys(TrainedTroopsData)) {
        const [troopId,levl] = key.split(',')
        const level = Number(levl)
        const count = TrainedTroopsData[[troopId,level]];
        if (!count || count <= 0) continue;

        const troopDef = AllTroopsData[troopId];
        if (!troopDef) continue;

        const name    = troopDef.name;
        const imgSrc  = `./Models/${name}.png`;

        const btn = document.createElement('button');
        btn.className = 'deploy-troop-btn';
        btn.dataset.troopId = troopId;
        btn.dataset.level   = level;

        const img = document.createElement('img');
        img.className = 'deploy-troop-img';
        img.src   = imgSrc;
        img.alt   = name;
        img.onerror = () => { img.style.display = 'none'; };

        const nameEl = document.createElement('p');
        nameEl.className = 'deploy-troop-name';
        nameEl.textContent = name;

        const countEl = document.createElement('span');
        countEl.className = 'deploy-troop-count';
        countEl.textContent = count;

        btn.appendChild(img);
        btn.appendChild(nameEl);
        btn.appendChild(countEl);
        deployTroopList.appendChild(btn);

        _troopButtons[[troopId,level]] = { btn, countEl, count };

        btn.addEventListener('click', (e) => {e.stopPropagation();_onTroopBtnClick(troopId, level)});
    }
}

function _onTroopBtnClick(troopId, level) {
    const entry = _troopButtons[[troopId,level]];
    if (!entry || entry.count <= 0) return;

    if (_selectedTroopId !== null) {
        const prev = _troopButtons[[_selectedTroopId,_selectedTroopLvl]];
        if (prev) prev.btn.classList.remove('selected');
    }

    if (_selectedTroopId === troopId && _selectedTroopLvl === level) {
        _selectedTroopId  = null;
        _selectedTroopLvl = null;
        _deployMode = false;
        return;
    }

    _selectedTroopId  = troopId;
    _selectedTroopLvl = level;
    _deployMode = true;
    entry.btn.classList.add('selected');
}

function _showDeployBar() {
    deployBar.style.display = 'flex';
    void deployBar.offsetHeight;
    deployBar.classList.add('is-active');
}

function _hideDeployBar() {
    deployBar.classList.remove('is-active');

    setTimeout(() => {
        deployBar.style.display = 'none';
    }, 320);
    _deployMode = false;
    _selectedTroopId  = null;
    _selectedTroopLvl = null;
}

window.addEventListener('click', (event) => {
    if (!inBattle) return;
    if (replay) return
    const rect = canvas.getBoundingClientRect();
    mouse.x = ((event.clientX - rect.left) / rect.width)  *  2 - 1;
    mouse.y = -((event.clientY - rect.top)  / rect.height) *  2 + 1;
    raycaster.setFromCamera(mouse, camera);
    const intersects = raycaster.intersectObject(ground);
    if (!intersects.length) return;
    const hit      = intersects[0];
    const worldPos = hit.point;
    const gridX    = Math.round(worldPos.x / position_scaling);
    const gridY    = Math.round(worldPos.z / position_scaling);
    _spawnTroop(_selectedTroopId, _selectedTroopLvl, gridX, gridY);
});

function _spawnTroop(troopId, level, gridX, gridY) {
    const entry  = _troopButtons[[troopId,level]];
    if (!entry || entry.count <= 0) return;
    SendToServer({action:"spawn_troop",troop_id:troopId,troop_level:level,x:gridX,y: gridY});
}

const PooledArmy = {}
let LoadedArmy = {}
function SpawnTroop(datafromServer){
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
            `./Models/${AllTroopsData[troopId].name}.png`,
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
function DespawnTroops(){
    for (const [key,armyy] of Object.entries(LoadedArmy)) {
        for (const army of armyy) {
            army.visible = false
            if (PooledArmy[key]) PooledArmy[key].push(army)
            else PooledArmy[key] = [army]
            console.log("Despawning..", key, army)
        }
    }
    LoadedArmy = {}
}
var state = undefined

function simulate(deltaTime) {
    const aliveTroopsAttacker = state.AliveTroopAttacker;
    const aliveTroopsDefender = state.AliveTroopDefender;
    const aliveBuildings = state.AliveBuildings;
    const buildings = state.Buildings;

    const troopCountAttacker = aliveTroopsAttacker.length;
    const troopCountDefender = aliveTroopsDefender.length;
    const buildingCount = aliveBuildings.length;

    for (let i = 0; i < troopCountAttacker; i++) {
        const troop = aliveTroopsAttacker[i];
        const config = AllTroopsData[troop.userData.troopId];

        const tX = troop.position.x / position_scaling;
        const tY = troop.position.z / position_scaling;

        const prefCat = config.preferred_target;
        const hasPreferred = prefCat !== undefined && prefCat !== null;

        let bestBuildingIdx = -1;
        let minBuildingDstSq = Infinity;

        for (let abIdx = 0; abIdx < buildingCount; abIdx++) {
            const ab = aliveBuildings[abIdx];
            const bIdx = ab.BuildingIndex;
            const placedBuilding = buildings[bIdx];
            if (hasPreferred && AllBuildingData[placedBuilding.building_id].category !== prefCat && AllTroopsData[placedBuilding.building_id]!==BuildingCategory.Wall) {
                continue;
            }
            const dx = placedBuilding.grid_x - tX;
            const dy = placedBuilding.grid_y - tY;
            const dstSq = dx * dx + dy * dy;
            if (dstSq < minBuildingDstSq) {
                minBuildingDstSq = dstSq;
                bestBuildingIdx = bIdx;
            }
        }

        if (hasPreferred && bestBuildingIdx === -1) {
            for (let abIdx = 0; abIdx < buildingCount; abIdx++) {
                const ab = aliveBuildings[abIdx];
                const bIdx = ab.BuildingIndex;
                const placedBuilding = buildings[bIdx];
                const dx = placedBuilding.grid_x - tX;
                const dy = placedBuilding.grid_y - tY;
                const dstSq = dx * dx + dy * dy;
                if (dstSq < minBuildingDstSq) {
                    minBuildingDstSq = dstSq;
                    bestBuildingIdx = bIdx;
                }
            }
        }

        let bestDefenderTroopIdx = -1;
        let minDefenderDstSq = Infinity;

        for (let dtIdx = 0; dtIdx < troopCountDefender; dtIdx++) {
            const target = aliveTroopsDefender[dtIdx];
            const dx = target.position.x / position_scaling - tX;
            const dy = target.position.z / position_scaling - tY;
            const dstSq = dx * dx + dy * dy;
            if (dstSq < minDefenderDstSq) {
                minDefenderDstSq = dstSq;
                bestDefenderTroopIdx = dtIdx;
            }
        }

        const attackRangeSq = config.attack_range * config.attack_range;
        const troopIsNearer = bestDefenderTroopIdx !== -1 && minDefenderDstSq < minBuildingDstSq;

        if (troopIsNearer) {
            if (minDefenderDstSq > attackRangeSq) {
                const target = aliveTroopsDefender[bestDefenderTroopIdx];
                const dist = Math.sqrt(minDefenderDstSq);
                let moveDist = config.movement_speed * deltaTime;
                if (moveDist > dist) moveDist = dist;
                const ratio = moveDist / dist;
                const dx = target.position.x / position_scaling - tX;
                const dy = target.position.z / position_scaling - tY;
                troop.position.x += dx * ratio * position_scaling;
                troop.position.z += dy * ratio * position_scaling;
            }
        } else if (bestBuildingIdx !== -1) {
            if (minBuildingDstSq > attackRangeSq) {
                const targetBuilding = buildings[bestBuildingIdx];
                const dist = Math.sqrt(minBuildingDstSq);
                let moveDist = config.movement_speed * deltaTime;
                if (moveDist > dist) moveDist = dist;
                const ratio = moveDist / dist;
                const dx = targetBuilding.grid_x - tX;
                const dy = targetBuilding.grid_y - tY;
                troop.position.x += dx * ratio * position_scaling;
                troop.position.z += dy * ratio * position_scaling;
            }
        }
    }

    for (let i = 0; i < troopCountDefender; i++) {
        const troop = aliveTroopsDefender[i];
        const config = AllTroopsData[troop.userData.troopId];
        const tX = troop.position.x / position_scaling;
        const tY = troop.position.z / position_scaling;

        let bestTroopIdx = -1;
        let minDstSq = Infinity;

        for (let atIdx = 0; atIdx < troopCountAttacker; atIdx++) {
            const target = aliveTroopsAttacker[atIdx];
            const dx = target.position.x / position_scaling - tX;
            const dy = target.position.z / position_scaling - tY;
            const dstSq = dx * dx + dy * dy;
            if (dstSq < minDstSq) {
                minDstSq = dstSq;
                bestTroopIdx = atIdx;
            }
        }

        if (bestTroopIdx === -1) {
            continue;
        }

        const targetTroop = aliveTroopsAttacker[bestTroopIdx];
        const attackRangeSq = config.attack_range * config.attack_range;

        if (minDstSq > attackRangeSq) {
            const dist = Math.sqrt(minDstSq);
            let moveDist = config.movement_speed * deltaTime;
            if (moveDist > dist) moveDist = dist;
            const ratio = moveDist / dist;
            const dx = targetTroop.position.x / position_scaling - tX;
            const dy = targetTroop.position.z / position_scaling - tY;
            troop.position.x += dx * ratio * position_scaling;
            troop.position.z += dy * ratio * position_scaling;
        }
    }
}

function DealDamage(building_damage, attacker_troop_damage, defender_troop_damage, building_died, attacker_troop_died, defender_troop_died) {
    console.log("Damage", building_damage, attacker_troop_damage)
    const aliveTroopsAttacker = state.AliveTroopAttacker;
    const aliveTroopsDefender = state.AliveTroopDefender;
    const aliveBuildings = state.AliveBuildings;
    const container = canvas.parentElement;
    const W = canvas.clientWidth;
    const H = canvas.clientHeight;

    function worldToScreen(worldPos) {
        const vec = worldPos.clone().project(camera);
        return {
            x: (vec.x * 0.5 + 0.5) * W,
            y: (-vec.y * 0.5 + 0.5) * H,
        };
    }

    function spawnLabel(screenX, screenY, text, extraClass = '') {
        const el = document.createElement('div');
        el.className = 'dmg-label' + (extraClass ? ' ' + extraClass : '');
        el.textContent = text;
        const jitter = (Math.random() - 0.5) * 30;
        el.style.left = (screenX + jitter - 15) + 'px';
        el.style.top = (screenY - 20) + 'px';
        container.style.position = container.style.position || 'relative';
        container.appendChild(el);
        el.addEventListener('animationend', () => el.remove());
    }

    aliveBuildings.forEach((building, i) => {
        const dmg = building_damage[i];
        if (!dmg || dmg <= 0) return;
        const screen = worldToScreen(state.Buildings[building.BuildingIndex].Model.position);
        const died = building_died.includes(i);
        spawnLabel(screen.x, screen.y, `-${dmg}`, died ? 'died' : '');
    });

    aliveTroopsAttacker.forEach((troop, i) => {
        const dmg = attacker_troop_damage[i];
        if (!dmg || dmg <= 0) return;
        const screen = worldToScreen(troop.position);
        const died = attacker_troop_died.includes(i);
        spawnLabel(screen.x, screen.y, `-${dmg}`, died ? 'troop died' : 'troop');
    });

    aliveTroopsDefender.forEach((troop, i) => {
        const dmg = defender_troop_damage[i];
        if (!dmg || dmg <= 0) return;
        const screen = worldToScreen(troop.position);
        const died = defender_troop_died.includes(i);
        spawnLabel(screen.x, screen.y, `-${dmg}`, died ? 'troop died' : 'troop');
    });

    for (const buildingDiedElement of building_died.toReversed()) {
        state.Buildings[buildingDiedElement].Model.userData.is_broken = true;
        aliveBuildings.splice(buildingDiedElement, 1);
    }

    for (const attackerTroopDiedElement of attacker_troop_died.toReversed()) {
        const ded = aliveTroopsAttacker[attackerTroopDiedElement];
        aliveTroopsAttacker.splice(attackerTroopDiedElement, 1);
        ded.visible = false;
        if ((PooledArmy[ded.userData.troopId] ?? []).length > 0) PooledArmy[ded.userData.troopId].push(ded);
        else PooledArmy[ded.userData.troopId] = [ded];
    }

    for (const defenderTroopDiedElement of defender_troop_died.toReversed()) {
        const ded = aliveTroopsDefender[defenderTroopDiedElement];
        aliveTroopsDefender.splice(defenderTroopDiedElement, 1);
        ded.visible = false;
        if ((PooledArmy[ded.userData.troopId] ?? []).length > 0) PooledArmy[ded.userData.troopId].push(ded);
        else PooledArmy[ded.userData.troopId] = [ded];
    }

    if (building_died.length > 0) {
        LoadMap(state.Buildings);
    }
}
function fmt(n) {
    return (n != null && !isNaN(n)) ? Number(n).toLocaleString() : '0';
}

function makeImgWithFallback(src, alt, imgClass, fbClass, fbEmoji) {
    const img = document.createElement('img');
    img.className = imgClass;
    img.src = src; img.alt = alt;
    img.onerror = function () {
        const fb = document.createElement('div');
        fb.className = fbClass; fb.textContent = fbEmoji;
        this.parentNode && this.parentNode.replaceChild(fb, this);
    };
    return img;
}
function renderTroopRows(container, troopLoss) {
    container.innerHTML = '';
    const entries = Object.entries(troopLoss || {}).filter(([, v]) => v > 0);
    if (!entries.length) {
        const p = document.createElement('p');
        p.className = 'bo-no-loss'; p.textContent = 'No losses';
        container.appendChild(p); return;
    }
    entries.forEach(([id, count]) => {
        const data = (typeof AllTroopsData !== 'undefined' && AllTroopsData[id]) || {};
        const name = data.name || id;
        const row  = document.createElement('div');
        row.className = 'bo-troop-row';
        row.appendChild(makeImgWithFallback(
            `./Models/${name}.png`, name, 'bo-troop-img', 'bo-troop-img-fb', '⚔️'
        ));
        const nm = document.createElement('span');
        nm.className = 'bo-troop-name'; nm.textContent = name;
        const ct = document.createElement('span');
        ct.className = 'bo-troop-count'; ct.textContent = `${count} fallen`;
        row.appendChild(nm); row.appendChild(ct);
        container.appendChild(row);
    });
}

function BattleOver(battle_outcome, attacker_troop_loss, buildings_broken, defender_troop_loss,opponent_username,my_name) {
    const isAtt = IsAttacker;

    const dateStr = battle_outcome.fought_at
        ? new Date(battle_outcome.fought_at).toLocaleDateString('en-US',
            { month: 'short', day: 'numeric', year: 'numeric' })
        : '';
    document.getElementById('bo-eyebrow').textContent =
        'battle Report' + (dateStr ? ' · ' + dateStr : '');

    document.getElementById('bo-headline').textContent = replay?`${my_name} raided ${opponent_username}'s village.`:(isAtt
        ? `You struck ${opponent_username}'s village`
        : `${opponent_username} raided your village`);

    document.getElementById('bo-gold').textContent   = fmt(battle_outcome.gold_looted);
    document.getElementById('bo-elixir').textContent = fmt(battle_outcome.elixir_looted);
    document.getElementById('bo-dark').textContent   = fmt(battle_outcome.dark_elixir_looted);

    document.getElementById('bo-your-title').textContent  = replay? 'Attacker losses':(isAtt ? 'Your losses'  : 'Their losses');
    document.getElementById('bo-their-title').textContent = replay? 'defender losses':(isAtt ? 'Their losses' : 'Your losses');
    document.getElementById('bo-section-title-building').textContent = replay?'Buildings destroyed':(isAtt ? 'Buildings destroyed (their village)' : 'Buildings destroyed (your village)')
    renderTroopRows(
        document.getElementById('bo-your-troops'),
        isAtt ? attacker_troop_loss : defender_troop_loss
    );
    renderTroopRows(
        document.getElementById('bo-their-troops'),
        isAtt ? defender_troop_loss : attacker_troop_loss
    );

    const bldEntries = Object.entries(buildings_broken || {}).filter(([, v]) => v > 0);
    const bldSection = document.getElementById('bo-buildings-section');
    const bldGrid    = document.getElementById('bo-buildings-grid');
    bldGrid.innerHTML = '';
    if (bldEntries.length) {
        bldEntries.forEach(([id, count]) => {
            const data = (typeof AllBuildingData !== 'undefined' && AllBuildingData[id]) || {};
            const name = data.name || id;
            const card = document.createElement('div');
            card.className = 'bo-bld-card';
            card.appendChild(makeImgWithFallback(
                `./Models/${name}.png`, name, 'bo-bld-img', 'bo-bld-img-fb', '🏚️'
            ));
            card.insertAdjacentHTML('beforeend', `
                    <div>
                        <p class="bo-bld-name">${name}</p>
                        <p class="bo-bld-count">×${count} destroyed</p>
                    </div>
                `);
            bldGrid.appendChild(card);
        });
        bldSection.style.display = '';
    } else {
        bldSection.style.display = 'none';
    }

    const revengeBtn = document.getElementById('bo-revenge-btn');
    revengeBtn.style.display = !isAtt ? '' : 'none';

    document.getElementById('battle-over-overlay').classList.add('is-active');
}

function closeBattleOver(e) {
    if (e) e.stopPropagation()
    document.getElementById('battle-over-overlay').classList.remove('is-active');
}
document.getElementById('bo-close-btn').addEventListener('click', closeBattleOver);


document.getElementById('bo-replay-btn').addEventListener('click', (e)=> {
    e.stopPropagation()
    console.log('Replay clicked');
})

document.getElementById('bo-revenge-btn').addEventListener('click', (e)=> {
    e.stopPropagation()
    Revenge()
})
function Revenge(){

}
// endregion

// region Grid
const gridWidth = position_scaling * 100;
const gridDivisions = 200;
const centerLineColor = 0xff0055;
const gridLineColor = 0x444444;
const gridHelper = new THREE.GridHelper(gridWidth, gridDivisions, centerLineColor, gridLineColor);
gridHelper.position.set(0, 0, 0);
gridHelper.material.transparent = true;
gridHelper.material.opacity = 0.5;
scene.add(gridHelper);

const cellSize = position_scaling;
const highlightGeo = new THREE.PlaneGeometry(cellSize * 0.95, cellSize * 0.95);
const highlightMat = new THREE.MeshBasicMaterial({
    color: 0xffff00,
    transparent: true,
    opacity: 0.10,
    depthWrite: false,
});

let highlightMesh = null;

function refreshGridHighlights() {
    if (highlightMesh) {
        scene.remove(highlightMesh);
        highlightMesh.dispose?.();
    }

    const filledCells = Object.keys(Grid);
    if (filledCells.length === 0) return;

    highlightMesh = new THREE.InstancedMesh(highlightGeo, highlightMat, filledCells.length);
    highlightMesh.position.y = 0.01;
    highlightMesh.rotation.x = -Math.PI / 2;
    // highlightMesh.rotation.y = Math.PI/2

    const dummy = new THREE.Object3D();
    filledCells.forEach((key, i) => {
        const [x, y] = key.split(',').map(Number);
        dummy.position.set(x * cellSize, -y * cellSize, 0);
        dummy.updateMatrix();
        highlightMesh.setMatrixAt(i, dummy.matrix);
    });

    highlightMesh.instanceMatrix.needsUpdate = true;
    scene.add(highlightMesh);
}

refreshGridHighlights();
// endregion