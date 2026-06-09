import * as THREE from './THREE/three.module.js';
import { GLTFLoader } from './THREE/GLTFLoader.js';

const gltfLoader = new GLTFLoader();
const scene = new THREE.Scene();
const clock = new THREE.Clock()

// region Three.js setup
const aspect = window.innerWidth / window.innerHeight;
scene.background = new THREE.Color(0x87CEEB);// blue color

const camera = new THREE.PerspectiveCamera(60,aspect,0.1,1000)
camera.position.set(20, 60, 60);
camera.lookAt(0, 0, 0);

const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
document.getElementById('game-container').appendChild(renderer.domElement);

const ambientLight = new THREE.AmbientLight(0xffffff, 0.6);
scene.add(ambientLight);

const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
dirLight.position.set(20, 40, 20);
scene.add(dirLight);

const textureLoader = new THREE.TextureLoader();
const gridTexture = textureLoader.load('./grid.jpg');
gridTexture.wrapS = THREE.RepeatWrapping;
gridTexture.wrapT = THREE.RepeatWrapping;
const groundGeometry = new THREE.PlaneGeometry(74, 74);
const groundMaterial = new THREE.MeshStandardMaterial({
    map: gridTexture,
    transparent: true
});
const ground = new THREE.Mesh(groundGeometry, groundMaterial);
ground.rotation.x = -Math.PI / 2;
scene.add(ground);

const keys = { w: false, a: false, s: false, d: false };
const moveSpeed = 40;

window.addEventListener('keydown', (e) => {
    if (keys.hasOwnProperty(e.key.toLowerCase())) keys[e.key.toLowerCase()] = true;
});
window.addEventListener('keyup', (e) => {
    if (keys.hasOwnProperty(e.key.toLowerCase())) keys[e.key.toLowerCase()] = false;
});
function handleMovement(dt) {
    if (keys.w) camera.position.z -= moveSpeed*dt;
    if (keys.s) camera.position.z += moveSpeed*dt;
    if (keys.a) camera.position.x -= moveSpeed*dt;
    if (keys.d) camera.position.x += moveSpeed*dt;
}
function animate() {
    requestAnimationFrame(animate);

    handleMovement(clock.getDelta());

    renderer.render(scene, camera);
}
animate();

window.addEventListener('resize', () => {
    const newAspect = window.innerWidth / window.innerHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
});
// endregion

let AllBuildingData = JSON.parse(localStorage.getItem('All_building_data') || '{}');
let PlacedBuildings = JSON.parse(localStorage.getItem('Placed_building') || '[]');

/** @type {Record<string, THREE.Object3D[]>} */
const LoadedObjects = {}
/** @type {Record<string, THREE.Object3D[]>} */
const Pool = {}
const position_scaling = 1
const size_scaling = 10
function LoadMap(){
    for (const [key,val] of Object.entries(LoadedObjects)){
        for (const v of val){
            v.visible = false
            if (key in Pool)
                Pool[key].push(v);
            else Pool[key] = [v]
        }
        val.length = 0
    }

    for (const building of PlacedBuildings){
        console.log(building.building_id)
        console.log(AllBuildingData)
        const name = AllBuildingData[building.building_id].name
        if (name in Pool && Pool[name].length>0){
            const Model = Pool[name].pop()
            Model.scale.set(size_scaling, size_scaling, size_scaling)
            Model.position.set(building.grid_x * position_scaling, 0, building.grid_y * position_scaling);
            Model.visible = true
            if (name in LoadedObjects)
                LoadedObjects[name].push(Model)
            else LoadedObjects[name] = [Model]
        }
        else {
            gltfLoader.load(
                `./Models/GLB format/${name}.glb`,
                function (glb) {
                    const Model = glb.scene;
                    Model.scale.set(size_scaling, size_scaling, size_scaling);
                    Model.position.set(building.grid_x * position_scaling, 0, building.grid_y * position_scaling);
                    scene.add(Model);
                    if (name in LoadedObjects)
                        LoadedObjects[name].push(Model)
                    else
                        LoadedObjects[name] = [Model]
                },
                function (xhr) {
                    console.log((xhr.loaded / xhr.total * 100) + '% loaded');
                },
                function (error) {
                    console.error('Error loading the model:', error);
                }
            );
        }
    }
}

// region Network
var access_token = localStorage.getItem('access_token');
var refresh_token_b64 = localStorage.getItem('refresh_token_b64');
console.log(access_token)
console.log(refresh_token_b64)
let socket
function connectToGameServer() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const socketUrl = `${protocol}//${host}/ws?token=${access_token}`;
    socket = new WebSocket(socketUrl);
    socket.addEventListener("open", (event) => {
        console.log("Connected to the Village Combat backend via WebSocket!");
        console.log(access_token)
        SendToServer({ action: "INITIAL_LOAD", message: "",access_token });
    });

    socket.addEventListener("message", (event) => {
        console.log("Message from server: ", event.data);
        const data = JSON.parse(event.data)
        switch (data['msg_type']){
            case 'building_troop_of_user':
                PlacedBuildings = data.building
                localStorage.setItem('Placed_building',JSON.stringify(PlacedBuildings))
                // const troops = JSON.parse(data['troops'])
                if (Object.keys(AllBuildingData).length!==0){
                    LoadMap()
                }else {
                    SendToServer({action:'ALL_BUILDING_TROOP_DATA',access_token,message:""})
                }
                // TODO : troop related thing also
                break
            case 'building_troop':
                const buildings = data.building
                // const troops = data.troops
                const defence = data.defence
                const army = data.army
                const resource = data.resource
                const ParticularLevelData = data.particular_level_data
                for (const building of buildings){
                    AllBuildingData[building.building_id] = {name:building.name,category:building.category,grid_size_x:building.grid_size_x,grid_size_y:building.grid_size_y,levels:{}}
                }
                for (const res of resource){
                    AllBuildingData[res.building_id].resource_type = res.resource_type
                }
                for (const arm of army){

                }
                for (const def of defence){
                    AllBuildingData[def.building_id].attack_speed_seconds = def.attack_speed_seconds
                    AllBuildingData[def.building_id].attack_range = def.attack_range
                    AllBuildingData[def.building_id].damage_type = def.damage_type
                    AllBuildingData[def.building_id].unit_target = def.unit_target
                }
                for (const [key,val] of Object.entries(ParticularLevelData)){
                    const val = JSON.parse(val)
                    AllBuildingData[val.building_id].levels[val.level].health=val.base_stats.health
                    AllBuildingData[val.building_id].levels[val.level].update_gold_required=val.updrade_cost.gold_required
                    AllBuildingData[val.building_id].levels[val.level].update_elxir_required=val.updrade_cost.elixir_required
                    AllBuildingData[val.building_id].levels[val.level].update_dark_elixir_required=val.updrade_cost.dark_elixir_required
                    AllBuildingData[val.building_id].levels[val.level].update_or_gem_required=val.updrade_cost.or_gem_required
                    AllBuildingData[val.building_id].levels[val.level].update_time_required_required=val.updrade_cost.time_required_seconds
                    AllBuildingData[val.building_id].levels[val.level].update_townhall_level_required=val.updrade_cost.town_hall_level_required
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
                localStorage.setItem('All_building_data',JSON.stringify(AllBuildingData))
                if (PlacedBuildings.length!==0) LoadMap()
                else SendToServer({ action: "INITIAL_LOAD", message: "",access_token })
                // TODO : troop related thing also
                break
            case "building_level_detail":
                break
            case 'construction_complete':
                PlacedBuildings.push(data.placed_building)
                localStorage.setItem('Placed_building',JSON.stringify(PlacedBuildings))
                LoadMap()
                break
        }

    });

    socket.addEventListener("error", (error) => {
        console.error("WebSocket Error detected:", error);
    });

    socket.addEventListener("close", (event) => {
        console.warn("Disconnected from server. Attempting reconnection can go here.");
        window.location.href = './Login.html'
    });
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
//
// class PlacedBuilding{
//     constructor(building_id, grid_x, grid_y, current_level, dynamic_state,sceneObject) {
//         this.building_id = building_id
//         this.grid_x = grid_x
//         this.grid_y = grid_y
//         this.current_level = current_level
//         this.dynamic_state = dynamic_state
//         this.object = sceneObject
//         LoadBuildingFromCache(this)
//     }
// }
// function LoadBuildingFromCache(building){
//     if (AllBuildingData!=null){
//         AllBuildingData['']
//     }else {
//         console.error("buildings not found asking server!")
//         SendToServer({action:"ALL_BUILDING_TROOP_DATA"})
//     }
// }

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

// region Debug
function CreateBuilding(name,x,y){
    SendToServer({action:"CREATE_BUILDING",message:JSON.stringify({
            building_id:(Object.entries(AllBuildingData).find(([key,val])=>val.name===name))[0],x,y
        }),access_token})
}
// endregion
console.log(CreateBuilding)
window.CreateBuilding = CreateBuilding
