CREATE DOMAIN non_negative_int AS INT CHECK (VALUE >= 0);
CREATE DOMAIN town_hall_range AS INT CHECK (VALUE BETWEEN 1 AND 4);
CREATE DOMAIN non_negative_numeric AS NUMERIC CHECK (VALUE >= 0.0);

CREATE TYPE level_up_config AS
(
    gold_required            non_negative_int,
    elixir_required          non_negative_int,
    dark_elixir_required     non_negative_int,
    or_gem_required          non_negative_int,
    time_required_seconds    non_negative_int,
    town_hall_level_required town_hall_range
);

CREATE TYPE attack_type AS ENUM ('melee', 'ranged');
CREATE TYPE building_category AS ENUM ('townhall', 'defense', 'resource', 'army');
CREATE TYPE damage_type AS ENUM ('single_target', 'splash');
CREATE TYPE unit_target_type AS ENUM ('ground', 'ground_and_air', 'air');
CREATE TYPE resource_type AS ENUM ('gold', 'elixir', 'dark_elixir');

CREATE TABLE troop_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    preferred_target building_type,
    attack_type attack_type NOT NULL,
    movement_speed non_negative_int NOT NULL,
    attack_speed_seconds NUMERIC NOT NULL CHECK (attack_speed_seconds > 0),
    attack_range non_negative_int NOT NULL,
    housing_space non_negative_int NOT NULL DEFAULT 1,
    health non_negative_int[] NOT NULL,
    damage_per_shot non_negative_int[] NOT NULL,
    upgrade_profile level_up_config[] NOT NULL,

    CONSTRAINT match_troop_stat_levels CHECK (
        array_length(health, 1) = array_length(damage_per_shot, 1)
        ),

    CONSTRAINT match_troop_upgrade_levels CHECK (
        array_length(upgrade_profile, 1) = array_length(health, 1)
        )
);
CREATE TABLE building_configs_base
(
    building_id     UUID PRIMARY KEY            DEFAULT uuid_generate_v4() UNIQUE,
    name            VARCHAR(255)       NOT NULL UNIQUE,
    category        building_category  NOT NULL,
    grid_size_x     non_negative_int   NOT NULL DEFAULT 2,
    grid_size_y     non_negative_int   NOT NULL DEFAULT 2,
    health          non_negative_int[] NOT NULL,
    upgrade_profile level_up_config[]  NOT NULL,

    CONSTRAINT match_building_levels CHECK (
        array_length(upgrade_profile, 1) = array_length(health, 1)
        )
);

CREATE TABLE defense_building_stats
(
    building_id          UUID PRIMARY KEY,
    building_type        VARCHAR(255)         NOT NULL,
    attack_speed_seconds NUMERIC              NOT NULL CHECK (attack_speed_seconds > 0),
    attack_range         non_negative_numeric NOT NULL, -- Changed to numeric/float support
    damage_per_shot      non_negative_int[]   NOT NULL,
    damage_type          damage_type          NOT NULL,
    unit_target          unit_target_type     NOT NULL,
    FOREIGN KEY (building_id) REFERENCES building_configs_base (building_id) ON DELETE CASCADE
);
CREATE TABLE resource_building_stats
(
    building_id              UUID PRIMARY KEY,
    building_type            VARCHAR(255)           NOT NULL,
    resource_type            resource_type          NOT NULL,
    generation_rate_per_hour non_negative_numeric[] NOT NULL,
    storage_capacity         non_negative_int[]     NOT NULL,

    FOREIGN KEY (building_id) REFERENCES building_configs_base (building_id) ON DELETE CASCADE
);
CREATE TABLE army_building_stats
(
    building_id     UUID PRIMARY KEY,
    building_type   VARCHAR(255)       NOT NULL,
    troop_capacity  non_negative_int[] NOT NULL,

    FOREIGN KEY (building_id) REFERENCES building_configs_base (building_id) ON DELETE CASCADE
);

CREATE INDEX idx_troop_configs_unlock_level ON troop_configs (unlock_at_level);
CREATE INDEX idx_building_configs_category ON building_configs_base (category);
CREATE INDEX idx_resource_stats_type ON resource_building_stats (resource_type);