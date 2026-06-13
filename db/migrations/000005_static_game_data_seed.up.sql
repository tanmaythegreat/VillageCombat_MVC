--this file is AI generated!
-- =============================================================================
-- SEED: static_game_data
-- Extends town_hall_range domain to 1-6, then inserts all static game config.
-- UUIDs are generated automatically by the database (uuid_generate_v4() default).
-- FK references resolved by name via subqueries / CTEs.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 0. Patch domain to support levels 1-6
-- ---------------------------------------------------------------------------


-- =============================================================================
-- 1. TROOP CONFIGS
-- =============================================================================
INSERT INTO troop_configs (name, preferred_target, unlock_at_level, attack_type, movement_speed, attack_speed_seconds, attack_range, housing_space)
VALUES
    ('Barbarian', 'defense',  1, 'melee'::attack_type,  1.0, 1.0, 1.0, 1),
    ('Archer',    'defense',  1, 'ranged'::attack_type, 0.9, 1.0, 3.5, 1),
    ('Goblin',    'resource', 1, 'melee'::attack_type,  1.4, 1.0, 1.0, 1),
    ('Giant',     'defense',  1, 'melee'::attack_type,  0.6, 2.0, 1.0, 5);


-- =============================================================================
-- 2. TROOP LEVEL STATS
-- =============================================================================

-- Barbarian
INSERT INTO troop_level_stats (troop_id, level, health, damage_per_shot, training_time_seconds)
SELECT id, 0,    0,   0,   0 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 1,  300,  75,  20 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 2,  360,  95,  20 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 3,  430, 120,  20 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 4,  520, 150,  20 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 5,  620, 185,  20 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 6,  740, 225,  20 FROM troop_configs WHERE name = 'Barbarian';

-- Archer
INSERT INTO troop_level_stats (troop_id, level, health, damage_per_shot, training_time_seconds)
SELECT id, 0,    0,   0,   0 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 1,  200,  60,  25 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 2,  240,  75,  25 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 3,  285,  95,  25 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 4,  340, 115,  25 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 5,  405, 140,  25 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 6,  480, 170,  25 FROM troop_configs WHERE name = 'Archer';

-- Goblin
INSERT INTO troop_level_stats (troop_id, level, health, damage_per_shot, training_time_seconds)
SELECT id, 0,    0,   0,   0 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 1,  200,  80,  30 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 2,  240, 100,  30 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 3,  285, 125,  30 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 4,  340, 155,  30 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 5,  405, 190,  30 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 6,  480, 230,  30 FROM troop_configs WHERE name = 'Goblin';

-- Giant
INSERT INTO troop_level_stats (troop_id, level, health, damage_per_shot, training_time_seconds)
SELECT id, 0,    0,   0,   0 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 1, 1500, 110, 120 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 2, 1900, 140, 120 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 3, 2300, 175, 120 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 4, 2800, 215, 120 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 5, 3400, 260, 120 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 6, 4100, 310, 120 FROM troop_configs WHERE name = 'Giant';


-- ---------------------------------------------------------------------------
-- 3. TROOP UPGRADE COSTS (super-exponential resources + proportional gems)
-- ---------------------------------------------------------------------------

-- Barbarian (pure elixir + gems)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT id, 2,     10000,        10,   3600, 1 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 3,    100000,       100,  14400, 2 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 4,   1500000,      1500,  43200, 3 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 5,  30000000,     30000,  86400, 4 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 6, 750000000,    750000, 172800, 5 FROM troop_configs WHERE name = 'Barbarian';

-- Archer (pure elixir + gems)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT id, 2,     15000,        15,   3600, 1 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 3,    150000,       150,  18000, 2 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 4,   2250000,      2250,  50400, 3 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 5,  45000000,     45000,  93600, 4 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 6, 1125000000,  1125000, 180000, 5 FROM troop_configs WHERE name = 'Archer';

-- Goblin (elixir + gems)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT id, 2,      8000,         8,   3600, 1 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 3,     80000,        80,  14400, 2 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 4,   1200000,      1200,  43200, 3 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 5,  24000000,     24000,  86400, 4 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 6, 600000000,    600000, 172800, 5 FROM troop_configs WHERE name = 'Goblin';

-- Giant (elixir/dark elixir + gems)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, dark_elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT id, 2,     50000,      0,        50,  14400, 1 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 3,    500000,      0,       500,  43200, 2 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 4,   7500000,      0,      7500,  86400, 3 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 5,         0,   3000,       300, 172800, 5 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 6,         0,  30000,      3000, 259200, 6 FROM troop_configs WHERE name = 'Giant';

-- ---------------------------------------------------------------------------
-- 4. BUILDING CONFIGS BASE
-- ---------------------------------------------------------------------------
INSERT INTO building_configs_base (name, category, grid_size_x, grid_size_y)
VALUES
    -- Town Hall
    ('Town Hall',              'townhall'::building_category, 4, 4),

    -- Defense
    ('Cannon',                 'defense'::building_category,  3, 3),
    ('Archer Tower',           'defense'::building_category,  3, 3),
    ('Air Defense',            'defense'::building_category,  3, 3),

    -- Resource
    ('Gold Mine',              'resource'::building_category, 3, 3),
    ('Gold Storage',           'resource'::building_category, 3, 3),
    ('Elixir Collector',       'resource'::building_category, 3, 3),
    ('Elixir Storage',         'resource'::building_category, 3, 3),
    ('Dark Elixir Drill',      'resource'::building_category, 3, 3),
    ('Dark Elixir Storage',    'resource'::building_category, 3, 3),

    -- Army
    ('Barracks',               'army'::building_category,     3, 3);


-- =============================================================================
-- 5. BUILDING LEVEL STATS (health per level)
-- =============================================================================

-- Town Hall
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 1,  1500 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 2,  2000 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 3,  2600 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 4,  3200 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 5,  4000 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 6,  5000 FROM building_configs_base WHERE name = 'Town Hall';

-- Cannon
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 1,   420 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 2,   500 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 3,   590 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 4,   700 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 5,   840 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 6,  1000 FROM building_configs_base WHERE name = 'Cannon';

-- Archer Tower
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 1,   400 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 2,   480 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 3,   580 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 4,   700 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 5,   850 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 6,  1050 FROM building_configs_base WHERE name = 'Archer Tower';

-- Air Defense
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 1,   800 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 2,   950 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 3,  1100 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 4,  1300 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 5,  1550 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 6,  1850 FROM building_configs_base WHERE name = 'Air Defense';

-- Gold Mine
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 1,   400 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 2,   480 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 3,   560 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 4,   650 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 5,   750 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 6,   860 FROM building_configs_base WHERE name = 'Gold Mine';

-- Gold Storage
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 1,   400 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 2,   500 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 3,   600 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 4,   700 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 5,   800 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 6,  1000 FROM building_configs_base WHERE name = 'Gold Storage';

-- Elixir Collector
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 1,   400 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 2,   480 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 3,   560 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 4,   650 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 5,   750 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 6,   860 FROM building_configs_base WHERE name = 'Elixir Collector';

-- Elixir Storage
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 1,   400 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 2,   500 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 3,   600 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 4,   700 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 5,   800 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 6,  1000 FROM building_configs_base WHERE name = 'Elixir Storage';

-- Dark Elixir Drill
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 1,   500 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 2,   600 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 3,   700 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 4,   820 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 5,   960 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 6,  1100 FROM building_configs_base WHERE name = 'Dark Elixir Drill';

-- Dark Elixir Storage
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 1,   600 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 2,   700 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 3,   830 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 4,   970 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 5,  1150 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 6,  1400 FROM building_configs_base WHERE name = 'Dark Elixir Storage';

-- Barracks
INSERT INTO building_level_stats (building_id, level, health)
SELECT building_id, 0,     0 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 1,   250 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 2,   310 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 3,   380 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 4,   460 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 5,   550 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 6,   650 FROM building_configs_base WHERE name = 'Barracks';


-- =============================================================================
-- 6. DEFENSE BUILDING STATS (static per building, not per level)
-- =============================================================================
INSERT INTO defense_building_stats (building_id, attack_speed_seconds, attack_range, damage_type, unit_target)
SELECT building_id,        1.0, 9.0, 'single_target'::damage_type, 'ground'::unit_target_type         FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 0.5, 9.5, 'single_target'::damage_type, 'ground_and_air'::unit_target_type FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id,   0.8, 9.0, 'single_target'::damage_type, 'air'::unit_target_type            FROM building_configs_base WHERE name = 'Air Defense';


-- =============================================================================
-- 7. DEFENSE BUILDING LEVEL STATS (damage per shot per level)
-- =============================================================================

-- Cannon
INSERT INTO defense_building_level_stats (building_id, level, damage_per_shot)
SELECT building_id, 0,   0 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 1,  40 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 2,  55 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 3,  72 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 4,  92 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 5, 116 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 6, 144 FROM building_configs_base WHERE name = 'Cannon';

-- Archer Tower
INSERT INTO defense_building_level_stats (building_id, level, damage_per_shot)
SELECT building_id, 0,   0 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 1,  20 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 2,  28 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 3,  38 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 4,  50 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 5,  64 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 6,  80 FROM building_configs_base WHERE name = 'Archer Tower';

-- Air Defense
INSERT INTO defense_building_level_stats (building_id, level, damage_per_shot)
SELECT building_id, 0,   0 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 1, 100 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 2, 130 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 3, 165 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 4, 205 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 5, 250 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 6, 300 FROM building_configs_base WHERE name = 'Air Defense';


-- =============================================================================
-- 8. RESOURCE BUILDING STATS
-- =============================================================================
INSERT INTO resource_building_stats (building_id, resource_type)
SELECT building_id, 'gold'::resource_type        FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 'gold'::resource_type        FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 'elixir'::resource_type      FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id,    'elixir'::resource_type      FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id,    'dark_elixir'::resource_type FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id,  'dark_elixir'::resource_type FROM building_configs_base WHERE name = 'Dark Elixir Storage';


-- =============================================================================
-- 9. RESOURCE BUILDING LEVEL STATS
-- =============================================================================

-- Gold Mine
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
SELECT building_id, 0,     0,      0 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 1,   200,   1000 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 2,   400,   2500 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 3,   700,   5000 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 4,  1100,  10000 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 5,  1600,  20000 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 6,  2200,  35000 FROM building_configs_base WHERE name = 'Gold Mine';

-- Gold Storage
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
SELECT building_id, 0,     0,      0 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 1,     0,  10000 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 2,     0,  25000 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 3,     0,  75000 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 4,     0, 200000 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 5,     0, 450000 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 6,     0, 900000 FROM building_configs_base WHERE name = 'Gold Storage';

-- Elixir Collector
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
SELECT building_id, 0,     0,      0 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 1,   200,   1000 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 2,   400,   2500 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 3,   700,   5000 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 4,  1100,  10000 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 5,  1600,  20000 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 6,  2200,  35000 FROM building_configs_base WHERE name = 'Elixir Collector';

-- Elixir Storage
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
SELECT building_id, 0,     0,      0 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 1,     0,  10000 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 2,     0,  25000 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 3,     0,  75000 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 4,     0, 200000 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 5,     0, 450000 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 6,     0, 900000 FROM building_configs_base WHERE name = 'Elixir Storage';

-- Dark Elixir Drill
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
SELECT building_id, 0,     0,      0 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 1,    20,     80 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 2,    40,    160 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 3,    65,    280 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 4,    95,    440 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 5,   130,    640 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 6,   180,    900 FROM building_configs_base WHERE name = 'Dark Elixir Drill';

-- Dark Elixir Storage
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
SELECT building_id, 0,     0,      0 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 1,     0,   1000 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 2,     0,   2500 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 3,     0,   5000 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 4,     0,  10000 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 5,     0,  20000 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 6,     0,  45000 FROM building_configs_base WHERE name = 'Dark Elixir Storage';


-- =============================================================================
-- 10. ARMY BUILDING STATS
-- =============================================================================
INSERT INTO army_building_stats (building_id)
SELECT building_id FROM building_configs_base WHERE name = 'Barracks';


-- =============================================================================
-- 11. ARMY BUILDING LEVEL STATS
-- =============================================================================
INSERT INTO army_building_level_stats (building_id, level, troop_capacity)
SELECT building_id, 0,   0 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 1,  20 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 2,  30 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 3,  45 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 4,  60 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 5,  80 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 6, 100 FROM building_configs_base WHERE name = 'Barracks';


-- =============================================================================
-- 12. BUILDING UPGRADE COSTS (super-exponential resources + proportional gems)
-- =============================================================================

-- ── Town Hall (5x Gem Multiplier) ─────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           0,          5,       0, 1 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 1,           0,           0,         10,       0, 1 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 2,       25000,        5000,        150,   14400, 1 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 3,      250000,       50000,       1500,   86400, 2 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 4,     3750000,      750000,      22500,  259200, 3 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 5,    75000000,    15000000,     450000,  432000, 4 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 6,  1875000000,   375000000,   11250000,  604800, 5 FROM building_configs_base WHERE name = 'Town Hall';

-- ── Cannon ────────────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           0,          1,       0, 1 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 1,           0,           0,          2,       0, 1 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 2,        4000,           0,          4,   1800, 1 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 3,       40000,           0,         40,   7200, 2 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 4,      600000,           0,        600,  21600, 3 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 5,    12000000,        5000,      12005,  57600, 4 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 6,   300000000,       75000,     300075, 129600, 5 FROM building_configs_base WHERE name = 'Cannon';

-- ── Archer Tower ──────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           0,          1,       0, 1 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 1,           0,           0,          2,       0, 1 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 2,        6000,           0,          6,   3600, 1 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 3,       60000,           0,         60,  10800, 2 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 4,      900000,           0,        900,  28800, 3 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 5,    18000000,        8000,      18008,  72000, 4 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 6,   450000000,      120000,     450120, 151200, 5 FROM building_configs_base WHERE name = 'Archer Tower';

-- ── Air Defense ───────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           0,          1,       0, 1 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 1,           0,           0,          2,       0, 2 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 2,       15000,           0,         15,   7200, 2 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 3,      150000,           0,        150,  21600, 3 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 4,     2250000,           0,       2250,  57600, 4 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 5,    45000000,       12000,      45012,  86400, 4 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 6,  1125000000,      180000,    1125180, 172800, 5 FROM building_configs_base WHERE name = 'Air Defense';

-- ── Gold Mine ─────────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           1,          0, 1 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 1,           0,           2,          0, 1 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 2,         500,           1,        900, 1 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 3,        5000,           5,       3600, 1 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 4,       75000,          75,      14400, 2 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 5,     1500000,        1500,      43200, 3 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 6,    37500000,       37500,     108000, 4 FROM building_configs_base WHERE name = 'Gold Mine';

-- ── Gold Storage ──────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           1,          0, 1 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 1,           0,           2,          0, 1 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 2,        1000,           1,       1800, 1 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 3,       10000,          10,       7200, 2 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 4,      150000,         150,      28800, 3 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 5,     3000000,        3000,      86400, 4 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 6,    75000000,       75000,     172800, 5 FROM building_configs_base WHERE name = 'Gold Storage';

-- ── Elixir Collector ──────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           1,          0, 1 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 1,           0,           2,          0, 1 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 2,         500,           1,        900, 1 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 3,        5000,           5,       3600, 1 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 4,       75000,          75,      14400, 2 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 5,     1500000,        1500,      43200, 3 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 6,    37500000,       37500,     108000, 4 FROM building_configs_base WHERE name = 'Elixir Collector';

-- ── Elixir Storage ────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           1,          0, 1 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 1,           0,           2,          0, 1 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 2,        1000,           1,       1800, 1 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 3,       10000,          10,       7200, 2 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 4,      150000,         150,      28800, 3 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 5,     3000000,        3000,      86400, 4 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 6,    75000000,       75000,     172800, 5 FROM building_configs_base WHERE name = 'Elixir Storage';

-- ── Dark Elixir Drill (unlocked at TH5) ───────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, dark_elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           1,          0, 1 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 1,           0,           2,          0, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 2,          50,           5,       7200, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 3,         500,          50,      21600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 4,        7500,         750,      57600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 5,      150000,       15000,     129600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 6,     3750000,      375000,     259200, 6 FROM building_configs_base WHERE name = 'Dark Elixir Drill';

-- ── Dark Elixir Storage (unlocked at TH5) ─────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, dark_elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           1,          0, 1 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 1,           0,           2,          0, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 2,          60,           6,       7200, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 3,         600,          60,      21600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 4,        9000,         900,      57600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 5,      180000,       18000,     129600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 6,     4500000,      450000,     259200, 6 FROM building_configs_base WHERE name = 'Dark Elixir Storage';

-- ── Barracks (elixir) ─────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 0,           0,           1,          0, 1 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 1,           0,           2,          0, 1 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 2,        1500,           2,       1800, 1 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 3,       15000,          15,       7200, 2 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 4,      225000,         225,      21600, 3 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 5,     4500000,        4500,      57600, 4 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 6,   112500000,      112500,     129600, 5 FROM building_configs_base WHERE name = 'Barracks';