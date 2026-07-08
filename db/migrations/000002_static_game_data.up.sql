CREATE DOMAIN non_negative_int AS INT CHECK (VALUE >= 0);
CREATE DOMAIN non_negative_bigint AS BIGINT CHECK (VALUE >= 0);
CREATE DOMAIN town_hall_range AS INT CHECK (VALUE BETWEEN 0 AND 12);
CREATE DOMAIN non_negative_numeric AS NUMERIC CHECK (VALUE >= 0.0);

CREATE TYPE attack_type AS ENUM ('melee', 'ranged');
CREATE TYPE building_category AS ENUM ('townhall', 'defense', 'resource', 'army','wall');
CREATE TYPE damage_type AS ENUM ('single_target', 'splash');
CREATE TYPE unit_target_type AS ENUM ('ground', 'ground_and_air', 'air');
CREATE TYPE resource_type AS ENUM ('gold', 'elixir', 'dark_elixir');

CREATE TABLE troop_configs
(
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name             VARCHAR(255)         NOT NULL UNIQUE,
    preferred_target building_category,
    attack_type      attack_type          NOT NULL,
    movement_speed   non_negative_numeric NOT NULL,
    attack_speed_seconds NUMERIC          NOT NULL CHECK (attack_speed_seconds > 0),
    attack_range     non_negative_numeric NOT NULL,
    housing_space    non_negative_int     NOT NULL DEFAULT 1
);

CREATE TABLE building_configs_base
(
    building_id UUID PRIMARY KEY DEFAULT uuid_generate_v4() UNIQUE,
    name        VARCHAR(255)      NOT NULL UNIQUE,
    category    building_category NOT NULL,
    grid_size_x non_negative_int  NOT NULL DEFAULT 0,
    grid_size_y non_negative_int  NOT NULL DEFAULT 0
);

CREATE TABLE troop_level_stats
(
    troop_id               UUID             NOT NULL REFERENCES troop_configs (id) ON DELETE CASCADE ON UPDATE CASCADE,
    level                  town_hall_range  NOT NULL,
    health                 non_negative_int NOT NULL,
    damage_per_shot        non_negative_int NOT NULL,
    PRIMARY KEY (troop_id, level)
);

CREATE TABLE building_level_stats
(
    building_id UUID             NOT NULL REFERENCES building_configs_base (building_id) ON DELETE CASCADE ON UPDATE CASCADE,
    level       town_hall_range  NOT NULL,
    health      non_negative_int NOT NULL,
    PRIMARY KEY (building_id, level)
);

CREATE TABLE upgrade_costs
(
    id                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    troop_id                 UUID REFERENCES troop_configs (id) ON DELETE CASCADE ON UPDATE CASCADE,
    building_id              UUID REFERENCES building_configs_base (building_id) ON DELETE CASCADE ON UPDATE CASCADE,
    upgrade_to_level         town_hall_range  NOT NULL,
    gold_required            non_negative_int NOT NULL DEFAULT 0,
    elixir_required          non_negative_int NOT NULL DEFAULT 0,
    dark_elixir_required     non_negative_int NOT NULL DEFAULT 0,
    or_gem_required          non_negative_int NOT NULL DEFAULT 0,
    time_required_seconds    non_negative_int NOT NULL,
    town_hall_level_required town_hall_range  NOT NULL,
    CONSTRAINT only_one_parent CHECK (
        (troop_id IS NOT NULL AND building_id IS NULL) OR
        (troop_id IS NULL AND building_id IS NOT NULL)
        )
);

CREATE TABLE defense_building_stats
(
    building_id          UUID PRIMARY KEY,
    attack_speed_seconds NUMERIC              NOT NULL CHECK (attack_speed_seconds > 0),
    attack_range         non_negative_numeric NOT NULL,
    damage_type          damage_type          NOT NULL,
    unit_target          unit_target_type     NOT NULL,
    FOREIGN KEY (building_id) REFERENCES building_configs_base (building_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE defense_building_level_stats
(
    building_id     UUID             NOT NULL REFERENCES building_configs_base (building_id) ON DELETE CASCADE ON UPDATE CASCADE,
    level           town_hall_range  NOT NULL,
    damage_per_shot non_negative_int NOT NULL,
    PRIMARY KEY (building_id, level)
);

CREATE TABLE resource_building_stats
(
    building_id   UUID PRIMARY KEY,
    resource_type resource_type NOT NULL,
    FOREIGN KEY (building_id) REFERENCES building_configs_base (building_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE resource_building_level_stats
(
    building_id              UUID                 NOT NULL REFERENCES building_configs_base (building_id) ON DELETE CASCADE ON UPDATE CASCADE,
    level                    town_hall_range       NOT NULL,
    generation_rate_per_hour non_negative_numeric NOT NULL,
    storage_capacity         non_negative_int     NOT NULL,
    PRIMARY KEY (building_id, level)
);

CREATE TABLE army_building_stats
(
    building_id   UUID PRIMARY KEY,
    FOREIGN KEY (building_id) REFERENCES building_configs_base (building_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE army_building_level_stats
(
    building_id    UUID             NOT NULL REFERENCES building_configs_base (building_id) ON DELETE CASCADE ON UPDATE CASCADE,
    level          town_hall_range  NOT NULL,
    troop_capacity non_negative_int NOT NULL,
    PRIMARY KEY (building_id, level)
);

CREATE INDEX idx_building_configs_category ON building_configs_base (category);
CREATE INDEX idx_resource_stats_type ON resource_building_stats (resource_type);
CREATE INDEX idx_troop_level_stats_troop ON troop_level_stats (troop_id);
CREATE INDEX idx_building_level_stats_building ON building_level_stats (building_id);
CREATE INDEX idx_upgrade_costs_troop ON upgrade_costs (troop_id);
CREATE INDEX idx_upgrade_costs_building ON upgrade_costs (building_id);