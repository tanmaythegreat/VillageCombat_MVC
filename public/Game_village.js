import * as THREE from './THREE/three.module.js';
import { GLTFLoader } from './THREE/GLTFLoader.js';
const gltfLoader = new GLTFLoader();
const scene = new THREE.Scene();
scene.background = new THREE.Color(0x87CEEB);// blue color
const aspect = window.innerWidth / window.innerHeight;

const viewSize = 40; // camera view size

const camera = new THREE.OrthographicCamera(
    (viewSize * aspect) / -2,
    (viewSize * aspect) / 2,
    viewSize / 2,
    viewSize / -2,
    0.1,
    1000
);
camera.position.set(0, 40, 40);
camera.lookAt(0, 0, 0);

const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
document.getElementById('game-container').appendChild(renderer.domElement);

const ambientLight = new THREE.AmbientLight(0xffffff, 0.6);
scene.add(ambientLight);

const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
dirLight.position.set(20, 40, 20);
scene.add(dirLight);

// The Grass/Ground
const groundGeometry = new THREE.PlaneGeometry(1000, 1000);
const groundMaterial = new THREE.MeshStandardMaterial({ color: 0x4CAF50 });
const ground = new THREE.Mesh(groundGeometry, groundMaterial);
ground.rotation.x = -Math.PI / 2;
scene.add(ground);

gltfLoader.load(
    './Models/GLB format/weapon-cannon.glb',
    function (glb) {
        const cannonModel = glb.scene;
        cannonModel.scale.set(10, 10, 10);
        cannonModel.position.set(0, 0, 0);
        scene.add(cannonModel);
    },
    function (xhr) {
        console.log((xhr.loaded / xhr.total * 100) + '% loaded');
    },
    function (error) {
        console.error('Error loading the model:', error);
    }
);

const keys = { w: false, a: false, s: false, d: false };
const moveSpeed = .5;

window.addEventListener('keydown', (e) => {
    if (keys.hasOwnProperty(e.key.toLowerCase())) keys[e.key.toLowerCase()] = true;
});

window.addEventListener('keyup', (e) => {
    if (keys.hasOwnProperty(e.key.toLowerCase())) keys[e.key.toLowerCase()] = false;
});

function handleMovement() {
    if (keys.w) camera.position.z -= moveSpeed;
    if (keys.s) camera.position.z += moveSpeed;
    if (keys.a) camera.position.x -= moveSpeed;
    if (keys.d) camera.position.x += moveSpeed;
}

function animate() {
    requestAnimationFrame(animate);

    handleMovement();

    renderer.render(scene, camera);
}

animate();

window.addEventListener('resize', () => {
    const newAspect = window.innerWidth / window.innerHeight;

    camera.left = (viewSize * newAspect) / -2;
    camera.right = (viewSize * newAspect) / 2;
    camera.top = viewSize / 2;
    camera.bottom = viewSize / -2;

    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
});