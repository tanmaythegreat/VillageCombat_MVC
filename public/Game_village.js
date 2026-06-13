import * as THREE from './THREE/three.module.js';
import {GLTFLoader} from './THREE/GLTFLoader.js';

const gltfLoader = new GLTFLoader();
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

    camera.position.set(-600, 600, 600);
    camera.lookAt(0, 0, 0);
}

const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
document.getElementById('game-container').appendChild(renderer.domElement);

const ambientLight = new THREE.AmbientLight(0xffffff, 0.6);
scene.add(ambientLight);

const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
dirLight.position.set(20, 40, 20);
scene.add(dirLight);

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
    handleMovement(clock.getDelta());
    tickCountdowns(scene,camera);
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
let UserData = null
let ConstructionTasks = null
/** @type {Record<string, THREE.Object3D[]>} */
const LoadedObjects = {}
/** @type {Record<string, THREE.Object3D[]>} */
const Pool = {}
const position_scaling = 2
const size_scaling = 20
async function LoadMap() {
    const textureLoadPromises = [];

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
    console.log(PlacedBuildings)

    for (const building of PlacedBuildings) {
        const name = AllBuildingData[building.building_id].name

        if (name in Pool && Pool[name].length > 0) {
            const Model = Pool[name].pop()
            Model.position.set(building.grid_x * position_scaling, 73 / 67 * size_scaling / 2, building.grid_y * position_scaling);
            Model.visible = true
            Model.userData = building;

            if (name in LoadedObjects)
                LoadedObjects[name].push(Model)
            else LoadedObjects[name] = [Model]

            building.Model = Model
        } else {
            const loadPromise = new Promise((resolve, reject) => {
                textureLoader.load(
                    `./Models/${AllBuildingData[building.building_id].name}.png`,
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
                            alphaTest: 0.3
                        });

                        const object_building = new THREE.Mesh(mesh, material);
                        object_building.rotation.y = -Math.PI / 4;
                        object_building.position.set(
                            building.grid_x * position_scaling,
                            size_scaling / 2,
                            building.grid_y * position_scaling
                        );
                        object_building.userData = building;
                        scene.add(object_building);
                        building.Model = object_building;

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

    console.log("MAP loaded , summoning jutsu!")
    for (const constructionTask of ConstructionTasks) {
        SummontaskCountDown(constructionTask)
    }
}


function AddLevelDetails(ParticularLevelData){
    console.log(ParticularLevelData)
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

// region CountDown
function SummontaskCountDown(task){

    const placed_building = PlacedBuildings.find((building) => building.id === task.placed_building_id)

    task.building=placed_building

    const durationSeconds = task.duration_seconds;
    const countdown = createCountdownSprite(durationSeconds*1000,new Date(task.started_at).getTime(),()=>{
        SendToServer({action:"CHECK_CONSTRUCTION_WORK",access_token,message:""})
    });
    console.log(placed_building)
    countdown.position.copy(placed_building.Model.position);
    countdown.position.y += size_scaling + 0.6;

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

// region Shop – state
let _shopGridX = 0;
let _shopGridY = 0;
let _shopSelectedId = null;
let _shopActiveFilter = 'All';
let _shopSearchQuery = '';
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
};

document.getElementById('shop-gem-btn').onclick = (e) => {
    e.stopPropagation()
    if (_shopSelectedId === null) return;
    closeShop();
    CreateBuilding(_shopSelectedId, _shopGridX, _shopGridY, true);
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

function canAfford(level, checkGems = false,placed_building_id = null) {
    const alreading = !(placed_building_id!=null && ConstructionTasks.some(task=>task.placed_building_id===placed_building_id))
    const townhall_Level = UserData.town_hall_level>=level.update_townhall_level_required
    console.log("consoling")
    console.log(alreading)
    console.log(townhall_Level)
    console.log(UserData)
    console.log(level)
    if (checkGems) {
        return alreading&&townhall_Level&&(UserData.current_gems ?? 0) >= (level.update_or_gem_required ?? 0);
    }
    const goldOk      = (UserData.current_gold       ?? 0) >= (level.update_gold_required        ?? 0);
    const elixirOk    = (UserData.current_elixir     ?? 0) >= (level.update_elxir_required       ?? 0);
    const darkElixirOk= (UserData.current_dark_elixir?? 0) >= (level.update_dark_elixir_required ?? 0);
    return alreading&&townhall_Level&&goldOk && elixirOk && darkElixirOk;
}


function setAffordability(btn, affordable) {
    if (!btn) return;
    btn.disabled = !affordable;
    btn.classList.toggle('btn-unaffordable', !affordable);
}

// endregion

function onMouseClick(event) {
    console.log(event.eventPhase)
    const bmOverlay   = document.getElementById('bm-overlay');
    const shopOverlay = document.getElementById('shop-overlay');
    if (bmOverlay.classList.contains('is-active') || shopOverlay.classList.contains('is-active')) return;
    mouse.x = (event.clientX / window.innerWidth)  *  2 - 1;
    mouse.y = -(event.clientY / window.innerHeight) *  2 + 1;
    raycaster.setFromCamera(mouse, camera);
    const intersects = raycaster.intersectObjects(scene.children, false);

    if (intersects.length > 0) {
        const hit = intersects[0];

        if (hit.object.userData && hit.object.userData.building_id) {
            triggerBuildingMenu(hit.object.userData);
        } else {
            const worldPos = hit.point;
            const gridX    = Math.round(worldPos.x / position_scaling);
            const gridY    = Math.round(worldPos.z / position_scaling);
            openBuildingShop(gridX, gridY);
        }
    }
}

window.addEventListener('click', onMouseClick);

// region AI generated part
function escapeHTML(str) {
    if (str === null || str === undefined) return '';
    return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

function triggerBuildingMenu(data) {
    console.log(`Opening menu for building ID: ${data.id} at grid position: ${data.gridX}, ${data.gridY}`);

    const building = AllBuildingData[data.building_id];
    const levelDetails = building.levels[data.level];

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

    if (isMaxLevel) {
        upgradeSection.style.display = 'none';
        maxNotice.style.display = 'block';
        upgradeBtn.disabled = true;
        if (gemUpgradeBtn) gemUpgradeBtn.disabled = true;
    } else {
        upgradeSection.style.display = 'block';
        maxNotice.style.display = 'none';

        document.getElementById('bm-upgrade-title-text').textContent = `Upgrade to level ${nextLevel}`;

        const thTag = document.getElementById('bm-th-req-tag');
        if (levelDetails.update_townhall_level_required) {
            thTag.textContent = `TH ${levelDetails.update_townhall_level_required} required`;
            thTag.style.display = 'inline-block';
        } else {
            thTag.style.display = 'none';
        }

        document.getElementById('bm-costs-container').innerHTML = buildCostRows(levelDetails);

        const affordable    = canAfford(levelDetails, false,data.id);
        const gemAffordable = canAfford(levelDetails, true,data.id);

        setAffordability(upgradeBtn,    affordable);
        setAffordability(gemUpgradeBtn, gemAffordable);

        if (gemUpgradeBtn) {
            const gemCost = levelDetails.update_or_gem_required ?? 0;
            gemUpgradeBtn.textContent = `💎 Instant (${formatNum(gemCost)})`;
        }
    }

    const overlay = document.getElementById('bm-overlay');
    overlay.classList.add('is-active');
    document.getElementById('bm-close-btn').onclick = (e) =>{e.stopPropagation(); overlay.classList.remove('is-active');}
    overlay.onclick = (e) => {e.stopPropagation(); if (e.target === overlay) overlay.classList.remove('is-active');};

    document.getElementById('bm-move-btn').onclick = (e) => {
        e.stopPropagation()
        overlay.classList.remove('is-active');
        // enterMoveMode(data);
    };

    upgradeBtn.onclick = (e) => {
        e.stopPropagation()
        if (!isMaxLevel) {
            overlay.classList.remove('is-active');
            UpgrageBuilding(data.id,false)
        }
    };

    if (gemUpgradeBtn) {
        gemUpgradeBtn.onclick = (e) => {
            e.stopPropagation()
            if (!isMaxLevel) {
                overlay.classList.remove('is-active');
                console.log(`Instant gem upgrade triggered for building: ${data.id}`);
                UpgrageBuilding(data.id,true)
            }
        };
    }
}

function getBuildingStats(building, level) {
    console.log(building)
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

function buildCostRows(level) {
    // Note: Gems removeda from here since they are now handled exclusively by the instant upgrade action button
    const costs = [
        { label: 'Gold',        icon: '🪙', val: level.update_gold_required },
        { label: 'Elixir',      icon: '🧪', val: level.update_elxir_required },
        { label: 'Dark elixir', icon: '⚗️',  val: level.update_dark_elixir_required },
    ].filter(c => c.val > 0);

    const rows = costs.map(c =>
        `<div class="bm-cost-row">
            <span class="bm-cost-label">${c.icon} ${c.label}</span>
            <span class="bm-cost-val">${formatNum(c.val)}</span>
         </div>`
    ).join('');

    return rows + `
        <div class="bm-divider"></div>
        <div class="bm-time-row">⏱ ${formatTime(level.update_time_required_required)} build time</div>`;
}

function formatNum(n) {
    if (!n) return '0';
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

// region Network
var access_token = localStorage.getItem('access_token');
var refresh_token_b64 = localStorage.getItem('refresh_token_b64');
// console.log(access_token)
// console.log(refresh_token_b64)
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
        // console.log("Message from server: ", event.data);
        const data = JSON.parse(event.data)
        switch (data['msg_type']) {
            case 'building_troop_of_user':
                UserData = data.user_data
                console.log(UserData)
                PlacedBuildings = data.building
                console.log(PlacedBuildings)
                localStorage.setItem('Placed_building', JSON.stringify(PlacedBuildings))
                ConstructionTasks = data.construction_tasks
                // const troops = JSON.parse(data['troops'])
                if (Object.keys(AllBuildingData).length !== 0) {
                    LoadMap()
                } else {
                    SendToServer({action: 'ALL_BUILDING_TROOP_DATA', access_token, message: ""})
                }
                // TODO : troop related thing also
                SendToServer({action:"CHECK_CONSTRUCTION_WORK",access_token,message:""})
                break
            case 'building_troop':
                // const troops = data.troops
                const buildings = data.building
                const defence = data.defence
                const army = data.army
                const resource = data.resource
                const ParticularLevelData = data.particular_level_data
                console.log(ParticularLevelData)
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
                if (PlacedBuildings.length !== 0) LoadMap()
                else SendToServer({action: "INITIAL_LOAD", message: "", access_token})
                // TODO : troop related thing also
                break
            case "building_level_detail":
                break
            case 'construction_started':
                ConstructionTasks.push(data.task)
                if (data.placed_building!=null) {
                    PlacedBuildings.push(data.placed_building)
                    LoadMap()
                }
                else {
                    SummontaskCountDown(data.task);
                }
                break
            case "construction_completed":
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
                    scene.remove(removedElement.building.Model.userData.countdownSprite);
                    removedElement.building.Model.userData.countdownSprite.material.dispose();
                    removedElement.building.Model.userData.countdownSprite.geometry.dispose();
                    delete removedElement.building.Model.userData.countdownSprite;
                }
                break

        }

        if (data['status'] && data['status']==='error'){console.error(data['message'])}
    });

    socket.addEventListener("error", (error) => {
        console.error("WebSocket Error detected:", error);
    });

    socket.addEventListener("close", (event) => {
        console.warn("Disconnected from server. Attempting reconnection can go here.");
        // window.location.href = './Login.html'
        connectToGameServer();
    });

}
function CreateBuilding(building_id, x, y, use_gems=false){
    SendToServer({action:"CREATE_BUILDING",message:JSON.stringify({
            building_id:building_id,x,y, use_gems
        }),access_token})
}
function UpgrageBuilding(placed_building_id,use_gems=false){
    SendToServer({action:"UPGRADE_BUILDING",message:JSON.stringify({
            placed_building_id,use_gems
        }),access_token})
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
const BuildingCategory={
    TownHall : "townhall",
    Defense  :"defense",
    Resource :"resource",
    Army     :"army"
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
    TroopTraining        :"troop_training"
}
// endregion


