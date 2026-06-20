import * as THREE from '../../THREE/three.module.js';

export const scene = new THREE.Scene();
export const clock = new THREE.Clock();

scene.background = new THREE.Color(0x87CEEB);

export const isOrthographic = true;
export let camera;
export let frustumSize = 240;

export const Right   = [1,  1];
export const Forward = [1, -1];

export const raycaster = new THREE.Raycaster();
export const mouse     = new THREE.Vector2();

// region Camera

const aspect = window.innerWidth / window.innerHeight;

if (!isOrthographic) {
    camera = new THREE.PerspectiveCamera(60, aspect, 0.1, 10000);
    camera.position.set(-60, 60, 60);
    camera.lookAt(0, 0, 0);
} else {
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

export function updateCamera() {
    const newAspect = window.innerWidth / window.innerHeight;
    if (isOrthographic) {
        camera.left   = (frustumSize * newAspect) / -2;
        camera.right  = (frustumSize * newAspect) /  2;
        camera.top    =  frustumSize / 2;
        camera.bottom =  frustumSize / -2;
    } else {
        camera.aspect = newAspect;
    }
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
}

window.addEventListener('resize', updateCamera);

// endregion

// region Renderer

export const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
document.getElementById('game-container').appendChild(renderer.domElement);
export const canvas = renderer.domElement;

// endregion

// region Lighting & Ground

export const ambientLight = new THREE.AmbientLight(0xffffff, 0.3);
scene.add(ambientLight);

export const textureLoader = new THREE.TextureLoader();
const gridTexture    = textureLoader.load('./Models/Map.jpeg');
const groundGeometry = new THREE.PlaneGeometry(2000, 2000);
const groundMaterial = new THREE.MeshStandardMaterial({ map: gridTexture, transparent: true });

export const ground = new THREE.Mesh(groundGeometry, groundMaterial);
ground.renderOrder = -1;
ground.rotation.z  = -Math.PI / 4;
ground.rotation.x  = -Math.PI / 2;
ground.position.set(0, 0, 0);
scene.add(ground);

// endregion

// region Camera Movement

const keys      = { w: false, a: false, s: false, d: false, e: false, q: false };
const moveSpeed = 200;

window.addEventListener('keydown', (e) => {
    if (keys.hasOwnProperty(e.key.toLowerCase())) keys[e.key.toLowerCase()] = true;
});
window.addEventListener('keyup', (e) => {
    if (keys.hasOwnProperty(e.key.toLowerCase())) keys[e.key.toLowerCase()] = false;
});

export function handleMovement(dt) {
    if (keys.d) { camera.position.z += moveSpeed * dt * Right[1];   camera.position.x += Right[0]   * moveSpeed * dt; }
    if (keys.w) { camera.position.z += moveSpeed * dt * Forward[1]; camera.position.x += Forward[0] * moveSpeed * dt; }
    if (keys.a) { camera.position.z -= moveSpeed * dt * Right[1];   camera.position.x -= Right[0]   * moveSpeed * dt; }
    if (keys.s) { camera.position.z -= moveSpeed * dt * Forward[1]; camera.position.x -= Forward[0] * moveSpeed * dt; }
    if (keys.e) { if (isOrthographic) { frustumSize += moveSpeed * dt; updateCamera(); } else camera.position.y += moveSpeed * dt; }
    if (keys.q) { if (isOrthographic) { frustumSize -= moveSpeed * dt; updateCamera(); } else camera.position.y -= moveSpeed * dt; }
}

// endregion

// region Grid

export const position_scaling = 20;
export const size_scaling     = 15;

const gridHelper = new THREE.GridHelper(position_scaling * 100, 200, 0xff0055, 0x444444);
gridHelper.material.transparent = true;
gridHelper.material.opacity     = 0.5;
scene.add(gridHelper);
function coordKey(x, y) {
    return `${x},${y}`;
}

let highlightGroup = null;

export function highlightGridSquares(grid, badpoints) {
    if (!badpoints) badpoints = []
    console.log(badpoints)
    if (highlightGroup) {
        scene.remove(highlightGroup);
        highlightGroup.traverse(obj => {
            if (obj.geometry) obj.geometry.dispose();
            if (obj.material) obj.material.dispose();
        });
    }

    highlightGroup = new THREE.Group();

    const squareGeo = new THREE.PlaneGeometry(position_scaling*0.95, position_scaling*0.95);

    const greenMat = new THREE.MeshBasicMaterial({
        color: 0x00ff55,
        transparent: true,
        opacity: 0.5,
        side: THREE.DoubleSide,
    });

    const redMat = new THREE.MeshBasicMaterial({
        color: 0xff0000,
        transparent: true,
        opacity: 0.5,
        side: THREE.DoubleSide,
    });

    // const badSet = new Set(badpoints.map(([x, y]) => coordKey(x, y)));
    const badSet = new Set(badpoints)
    // console.log(badpoints)
    const drawSquare = (x, y, material) => {
        const mesh = new THREE.Mesh(squareGeo, material);
        mesh.rotation.x = -Math.PI / 2; // lie flat, matching the GridHelper
        mesh.position.set(
            x * position_scaling,
            0.01, // tiny offset to avoid z-fighting with grid lines
            y * position_scaling
        );
        highlightGroup.add(mesh);
    };

    // green: every key in grid (unless it's also a bad point)
    for (const key of Object.keys(grid)) {
        const [x, y] = key.split(',').map(Number);
        if (!badSet.has(coordKey(x, y))) {
            drawSquare(x, y, greenMat);
        }
    }

    for (const k of badpoints) {
        const [x,y] = k.split(',')
        console.log(x,y)
        drawSquare(x, y, redMat);
    }

    scene.add(highlightGroup);
    return highlightGroup;
}// endregion
