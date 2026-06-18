CREATE TYPE troop_spawn AS
(
    troop_id            VARCHAR(255),
    troop_level         INT,
    spawned_by_attacker BOOLEAN,
    spawned_at_x        INT,
    spawned_at_y        INT,
    spawn_time          DOUBLE PRECISION
);

CREATE TYPE initial_battle_building AS
(
    building_id UUID,
    grid_x      INT,
    grid_y      INT,
    level       INT,
    is_broken   BOOLEAN
);

CREATE TABLE battle_record
(
    battle_id         UUID PRIMARY KEY REFERENCES battle_history (battle_id),
    troop_spawns      troop_spawn[],
    initial_buildings initial_battle_building[]
);