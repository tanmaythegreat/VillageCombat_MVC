import {_placed_building} from "../views/ui_hud.js";
import {highlightGridSquares, position_scaling} from "./scene.js";
import {AllBuildingData, Grid, LoadMap, PlacedBuildings} from "../models/map.js";
import {SendToServer} from "../controllers/network.js";

export let moving = false

export function moveUpdate(grid_x,grid_y){
    if (!moving) return
    const bData = AllBuildingData[_placed_building.building_id];
    const posX  = (grid_x + (bData.grid_size_y===1?0:(bData.grid_size_x / 2))) * position_scaling;
    const posZ  = (grid_y + (bData.grid_size_y===1?0:(bData.grid_size_y / 4))) * position_scaling;

    _placed_building.Model.position.x = posX
    _placed_building.Model.position.z = posZ
    let badpoints = false;
    for (let x = grid_x ; x < grid_x + AllBuildingData[_placed_building.building_id].grid_size_x; x++) {
        for (let y = grid_y ; y < grid_y + AllBuildingData[_placed_building.building_id].grid_size_y; y++) {
            if (Grid[[x,y]]){
                badpoints = true
                break;
            }
        }
    }
    let bp = {}
    for (let x = grid_x ; x < grid_x + AllBuildingData[_placed_building.building_id].grid_size_x; x++) {
        for (let y = grid_y ; y < grid_y + AllBuildingData[_placed_building.building_id].grid_size_y; y++) {
            bp[[x,y]] = 0
        }
    }

    _placed_building.grid_x = grid_x
    _placed_building.grid_y = grid_y
    if (badpoints)
        highlightGridSquares(Grid,Object.keys(bp))
    else
        highlightGridSquares({...Grid,...bp},[])
}
export function selectToMove(){
    moving = true;
    for (let x = _placed_building.grid_x ; x < _placed_building.grid_x + AllBuildingData[_placed_building.building_id].grid_size_x; x++) {
        for (let y = _placed_building.grid_y ; y < _placed_building.grid_y + AllBuildingData[_placed_building.building_id].grid_size_y; y++) {
            delete Grid[[x,y]]
        }
    }
}
export function putSelectedBuilding(grid_x,grid_y){
    moving = false
    SendToServer({action:"MOVE",message:JSON.stringify({placed_building_id:_placed_building.id,grid_x,grid_y})})
}