import {
    SpawnTroop, DespawnTroops, BattleOver, _hideDeployBar,
    setInBattle, setReplay, DealDamage, CancelMatchMaking,
} from './battle.js';
import { SendToServer } from './network.js';

export function handleBattleMessage(data) {
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
            handleBattleOver(data);
            break;
    }
}

function handleBattleOver(data) {
    DespawnTroops();
    BattleOver(
        data.battle_outcome, data.attacker_troop_loss, data.buildings_broken,
        data.defender_troop_loss, data.opponent_username, data.my_username, data.battle_id,
    );
    setReplay(false);
    _hideDeployBar();
    setInBattle(false);
    CancelMatchMaking();
    SendToServer({ action: 'INITIAL_LOAD', message: '' });
}

