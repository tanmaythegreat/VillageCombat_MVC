CREATE DOMAIN non_negative_int AS INT CHECK (VALUE >= 0);
CREATE DOMAIN town_hall_range AS INT CHECK (VALUE BETWEEN 1 AND 4);

CREATE TYPE level_up_config AS (
   gold_required            non_negative_int,
   elixir_required          non_negative_int,
   dark_elixir_required     non_negative_int,
   or_gem_required          non_negative_int,
   time_required_seconds    non_negative_int,
   town_hall_level_required town_hall_range
);

CREATE TYPE attack_type AS ENUM ('melee', 'ranged');
CREATE TYPE building_category AS ENUM ('townhall', 'defense', 'resource', 'army');
CREATE TYPE building_type AS ENUM ('cannon', 'archer_tower', 'air_defense', 'gold_storage', 'gold_mine', 'elixir_collector', 'elixir_storage', 'dark_elixir_storage', 'dark_elixir_drill', 'barracks', 'base_townhall');
CREATE TYPE damage_type AS ENUM ('single_target', 'splash');
CREATE TYPE unit_target_type AS ENUM ('ground', 'ground_and_air', 'air');
CREATE TYPE resource_type AS ENUM ('gold', 'elixir', 'dark_elixir');

CREATE TABLE troop_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    preferred_target building_type, -- NULL means "Any"
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
CREATE TABLE building_configs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR(50) NOT NULL UNIQUE,
  category building_category NOT NULL,
  specific_type building_type NOT NULL,

  health non_negative_int[] NOT NULL,
-- First element tell the building cost
  upgrade_profile level_up_config[] NOT NULL,

-- DEFENSE SPECIFIC
  attack_speed_seconds NUMERIC CHECK (attack_speed_seconds > 0),
  attack_range non_negative_int,
  damage_per_shot non_negative_int[],
  damage_type damage_type,
  unit_target unit_target_type,

-- RESOURCE SPECIFIC
  resource_type resource_type,
  generation_rate_per_hour non_negative_int[],
  storage_capacity non_negative_int[],

-- ARMY SPECIFIC
  troop_capacity non_negative_int[],
  unlocked_troops VARCHAR[],

  CONSTRAINT match_building_levels CHECK (
      array_length(upgrade_profile, 1) = (array_length(health, 1))
      ),

  CONSTRAINT valid_building_hierarchy CHECK (
      (category = 'defense' AND specific_type IN ('cannon', 'archer_tower', 'air_defense')) OR
      (category = 'resource' AND specific_type IN ('gold_storage', 'gold_mine', 'elixir_collector', 'elixir_storage', 'dark_elixir_storage', 'dark_elixir_drill')) OR
      (category = 'army' AND specific_type = 'barracks') OR
      (category = 'townhall' AND specific_type = 'base_townhall')
      ),

  CONSTRAINT enforce_defense_data CHECK (
      (category = 'defense' AND attack_speed_seconds IS NOT NULL AND attack_range IS NOT NULL AND damage_per_shot IS NOT NULL AND damage_type IS NOT NULL AND unit_target IS NOT NULL) OR
      (category != 'defense' AND attack_speed_seconds IS NULL AND attack_range IS NULL AND damage_per_shot IS NULL AND damage_type IS NULL AND unit_target IS NULL)
      ),

  CONSTRAINT enforce_resource_data CHECK (
      (category = 'resource' AND resource_type IS NOT NULL) OR
      (category != 'resource' AND resource_type IS NULL AND generation_rate_per_hour IS NULL AND storage_capacity IS NULL)
      ),

  CONSTRAINT enforce_army_data CHECK (
      (category = 'army' AND troop_capacity IS NOT NULL) OR
      (category != 'army' AND troop_capacity IS NULL AND unlocked_troops IS NULL)
      )
);