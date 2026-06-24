import {scene, clock, camera, renderer, ambientLight, canvas, ground, hideCircle, showCircle} from './core/scene.js';
import { handleMovement, raycaster,mouse } from './core/scene.js';
import {tickCountdowns, Grid, UserData, PlacedBuildings, AllBuildingData} from './models/map.js';
import {updateCollectButton, updateResourceDetailBox} from './views/ui_hud.js';
import { inBattle, simulate } from './controllers/battle.js';
import { openBuildingShop, triggerBuildingMenu as triggerMenu } from './views/ui_hud.js';
import { connectToGameServer } from './controllers/network.js';
import * as THREE from '../THREE/three.module.js';
import {initProfile} from "./views/profile.js";
import {moveUpdate, moving, putSelectedBuilding} from "./core/move.js";
import {BuildingCategory} from "./core/enums.js";

const position_scaling = 20;

const meshCube = new THREE.Mesh(new THREE.BoxGeometry(1, 1, 1), new THREE.MeshNormalMaterial());
scene.add(meshCube);

window.addEventListener('mousemove', DragHandle);
export function DragHandle(event)  {
    mouse.x = (event.clientX / window.innerWidth) * 2 - 1;
    mouse.y = -(event.clientY / window.innerHeight) * 2 + 1;
    raycaster.setFromCamera(mouse, camera);
    const intersects = raycaster.intersectObject(ground);
    if (!intersects.length) return;
    const { x, z } = intersects[0].point;
    const X = Math.round(x / position_scaling)
    const Z = Math.round(z/ position_scaling)
    moveUpdate(X,Z)
    meshCube.position.set( X* position_scaling, 0, Z * position_scaling);
    if (Grid[[X,Z]] && AllBuildingData[Grid[[X,Z]].userData.building_id].category===BuildingCategory.Defense){
        showCircle(Grid[[X,Z]].userData.grid_x+Math.floor(AllBuildingData[Grid[[X,Z]].userData.building_id].grid_size_x/2),Grid[[X,Z]].userData.grid_y+Math.floor(AllBuildingData[Grid[[X,Z]].userData.building_id].grid_size_y/2),AllBuildingData[Grid[[X,Z]].userData.building_id].attack_range)
    }
    else {
        hideCircle()
    }
}

window.addEventListener('click', onMouseClick);
export function onMouseClick(event) {

    if (inBattle) return;
    const bmOverlay   = document.getElementById('bm-overlay');
    const shopOverlay = document.getElementById('shop-overlay');
    if (bmOverlay.classList.contains('is-active') || shopOverlay.classList.contains('is-active')) return;

    const rect = canvas.getBoundingClientRect();
    mouse.x =  ((event.clientX - rect.left) / rect.width)  *  2 - 1;
    mouse.y = -((event.clientY - rect.top)  / rect.height) *  2 + 1;

    raycaster.setFromCamera(mouse, camera);
    const intersects = raycaster.intersectObject(ground);
    if (!intersects.length) return;

    const { x, z }  = intersects[0].point;
    const gridX = Math.round(x / position_scaling);
    const gridY = Math.round(z / position_scaling);
    if (moving){
        putSelectedBuilding(gridX,gridY)
        return
    }
    if (Grid[[gridX, gridY]]) {
        triggerMenu(Grid[[gridX, gridY]].userData);
    } else {
        openBuildingShop(gridX, gridY);
    }
}

function animate() {
    requestAnimationFrame(animate);
    const dt = clock.getDelta();
    handleMovement(dt);

    if (inBattle) {
        try { simulate(dt); }
        catch (e) { console.log(e); }
    }

    // day-night cycle faking
    // ambientLight.intensity = Math.max(Math.min(0.5 + Math.cos(clock.elapsedTime * 0.01) * 0.5, 0.7), 0.2);

    renderer.render(scene, camera);
}

animate();
connectToGameServer();
initProfile()
function secondTick() {
    if (!UserData)return;
    updateCollectButton();
    tickCountdowns(camera);
    updateResourceDetailBox();
}
setInterval(secondTick, 1000);
document.addEventListener("DOMContentLoaded", () => {
    const controlsOverlay = document.getElementById("controls-overlay");

    if (controlsOverlay) {
        controlsOverlay.style.pointerEvents = "auto";
        setTimeout(() => {
            controlsOverlay.style.opacity = "0";
            setTimeout(() => {
                controlsOverlay.remove();
            }, 500);
        }, 1000);
    }
});
