CREATE TABLE battle_history
(
    battle_id          UUID PRIMARY KEY          DEFAULT uuid_generate_v4(),
    attacker_name      VARCHAR(50)      NOT NULL,
    defender_name      VARCHAR(50)      NOT NULL,
    elixir_looted      non_negative_int          DEFAULT 0,
    gold_looted        non_negative_int          DEFAULT 0,
    dark_elixir_looted non_negative_int          DEFAULT 0,
    fought_at          TIMESTAMP WITH TIME ZONE  DEFAULT CURRENT_TIMESTAMP,
    battle_duration    non_negative_int NOT NULL DEFAULT 0,
    do_defender_know   BOOLEAN          NOT NULL DEFAULT FALSE,
    winner_attacker    BOOLEAN          NOT NULL DEFAULT TRUE
);

CREATE TABLE battle_troop_losses (
     battle_id UUID NOT NULL,
     troop_id UUID NOT NULL,
     loss_count non_negative_int NOT NULL DEFAULT 0,
     is_attacker BOOLEAN NOT NULL DEFAULT True,

     PRIMARY KEY (battle_id, troop_id,is_attacker),
     CONSTRAINT fk_battle_troop FOREIGN KEY (battle_id) REFERENCES battle_history(battle_id),
     CONSTRAINT fk_troop_id FOREIGN KEY (troop_id) REFERENCES troop_configs(id)
);

CREATE TABLE buildings_broken
(
    battle_id   UUID             NOT NULL,
    building_id UUID             NOT NULL,
    count       non_negative_int NOT NULL DEFAULT 0,
    CONSTRAINT fk_battle_building FOREIGN KEY (battle_id) REFERENCES battle_history (battle_id) ON DELETE CASCADE,
    CONSTRAINT fk_building FOREIGN KEY (building_id) REFERENCES building_configs_base (building_id) ON DELETE CASCADE
);

CREATE TABLE user_status
(
    user_id       UUID PRIMARY KEY REFERENCES users (user_id) ON DELETE CASCADE,
    last_defended TIMESTAMPTZ,
    in_battle     BOOLEAN NOT NULL DEFAULT FALSE,
    attack_power  INTEGER NOT NULL DEFAULT 0,
    defence_power INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_battle_history_attacker_log ON battle_history (attacker_name, fought_at DESC);
CREATE INDEX idx_battle_history_defender_log ON battle_history (defender_name, fought_at DESC);
CREATE INDEX idx_buildings_broken_battle_id ON buildings_broken (battle_id);
CREATE INDEX idx_user_status_last_defended ON user_status (last_defended ASC);
CREATE INDEX idx_user_status_power ON user_status (defence_power DESC);