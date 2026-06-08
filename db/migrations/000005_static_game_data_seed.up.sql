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


-- =============================================================================
-- 3. TROOP UPGRADE COSTS
--    (elixir-based; dark elixir kicks in at level 5-6 for Giant)
--    upgrade_to_level N = cost to go FROM level N-1 TO level N
-- =============================================================================

-- Barbarian (pure elixir)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, time_required_seconds, town_hall_level_required)
SELECT id, 2,    10000,   3600, 1 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 3,    50000,  14400, 2 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 4,   150000,  43200, 3 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 5,   500000,  86400, 4 FROM troop_configs WHERE name = 'Barbarian'
UNION ALL
SELECT id, 6,  1500000, 172800, 5 FROM troop_configs WHERE name = 'Barbarian';

-- Archer (pure elixir)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, time_required_seconds, town_hall_level_required)
SELECT id, 2,    15000,   3600, 1 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 3,    75000,  18000, 2 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 4,   200000,  50400, 3 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 5,   600000,  93600, 4 FROM troop_configs WHERE name = 'Archer'
UNION ALL
SELECT id, 6,  1800000, 180000, 5 FROM troop_configs WHERE name = 'Archer';

-- Goblin (elixir, slightly cheaper)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, time_required_seconds, town_hall_level_required)
SELECT id, 2,     8000,   3600, 1 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 3,    40000,  14400, 2 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 4,   120000,  43200, 3 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 5,   400000,  86400, 4 FROM troop_configs WHERE name = 'Goblin'
UNION ALL
SELECT id, 6,  1200000, 172800, 5 FROM troop_configs WHERE name = 'Goblin';

-- Giant (elixir up to 4, dark elixir for 5-6)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, dark_elixir_required, time_required_seconds, town_hall_level_required)
SELECT id, 2,   50000,      0,  14400, 1 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 3,  200000,      0,  43200, 2 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 4,  700000,      0,  86400, 3 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 5,       0,   3000, 172800, 5 FROM troop_configs WHERE name = 'Giant'
UNION ALL
SELECT id, 6,       0,   8000, 259200, 6 FROM troop_configs WHERE name = 'Giant';


-- =============================================================================
-- 4. BUILDING CONFIGS BASE
-- =============================================================================
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
INSERT INTO defense_building_stats (building_id, building_type, attack_speed_seconds, attack_range, damage_type, unit_target)
SELECT building_id, 'cannon',        1.0, 9.0, 'single_target'::damage_type, 'ground'::unit_target_type         FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 'archer_tower',  0.5, 9.5, 'single_target'::damage_type, 'ground_and_air'::unit_target_type FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 'air_defense',   0.8, 9.0, 'single_target'::damage_type, 'air'::unit_target_type            FROM building_configs_base WHERE name = 'Air Defense';


-- =============================================================================
-- 7. DEFENSE BUILDING LEVEL STATS (damage per shot per level)
-- =============================================================================

-- Cannon
INSERT INTO defense_building_level_stats (building_id, level, damage_per_shot)
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
INSERT INTO resource_building_stats (building_id, building_type, resource_type)
SELECT building_id, 'gold_mine',           'gold'::resource_type        FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 'gold_storage',        'gold'::resource_type        FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 'elixir_collector',    'elixir'::resource_type      FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 'elixir_storage',      'elixir'::resource_type      FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 'dark_elixir_drill',   'dark_elixir'::resource_type FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 'dark_elixir_storage', 'dark_elixir'::resource_type FROM building_configs_base WHERE name = 'Dark Elixir Storage';


-- =============================================================================
-- 9. RESOURCE BUILDING LEVEL STATS
--    Collectors/mines/drill: low generation, modest storage (~20% of dedicated storage)
--    Storages: zero generation, large storage capacity
-- =============================================================================

-- Gold Mine (generates gold; small storage buffer)
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
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

-- Gold Storage (no generation; large storage)
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
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

-- Elixir Collector (generates elixir; small storage buffer)
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
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

-- Elixir Storage (no generation; large storage)
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
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

-- Dark Elixir Drill (unlocked at TH5; generates dark elixir; small storage buffer)
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
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

-- Dark Elixir Storage (no generation; large storage; unlocked at TH5)
INSERT INTO resource_building_level_stats (building_id, level, generation_rate_per_hour, storage_capacity)
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
INSERT INTO army_building_stats (building_id, building_type)
SELECT building_id, 'barracks' FROM building_configs_base WHERE name = 'Barracks';


-- =============================================================================
-- 11. ARMY BUILDING LEVEL STATS
-- =============================================================================
INSERT INTO army_building_level_stats (building_id, level, troop_capacity)
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
-- 12. BUILDING UPGRADE COSTS
--     Defense  -> gold (primary) + small elixir top-up at higher levels
--     Resource -> gold for gold buildings, elixir for elixir buildings,
--                 dark elixir for dark elixir buildings (levels 5-6 heavy)
--     Army     -> elixir
--     Town Hall-> gold (primary) + elixir top-up
-- =============================================================================

-- ── Town Hall ─────────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, elixir_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,        0,        0,       0, 1 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 2,    25000,     5000,   14400, 1 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 3,   100000,    20000,   86400, 2 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 4,   400000,    80000,  259200, 3 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 5,  1000000,   200000,  432000, 4 FROM building_configs_base WHERE name = 'Town Hall'
UNION ALL
SELECT building_id, 6,  3000000,   500000,  604800, 5 FROM building_configs_base WHERE name = 'Town Hall';

-- ── Cannon ────────────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, elixir_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,        0,     0,      0, 1 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 2,     4000,     0,   1800, 1 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 3,    12000,     0,   7200, 2 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 4,    40000,     0,  21600, 3 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 5,   120000,  5000,  57600, 4 FROM building_configs_base WHERE name = 'Cannon'
UNION ALL
SELECT building_id, 6,   360000, 15000, 129600, 5 FROM building_configs_base WHERE name = 'Cannon';

-- ── Archer Tower ──────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, elixir_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,        0,     0,      0, 1 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 2,     6000,     0,   3600, 1 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 3,    18000,     0,  10800, 2 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 4,    60000,     0,  28800, 3 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 5,   180000,  8000,  72000, 4 FROM building_configs_base WHERE name = 'Archer Tower'
UNION ALL
SELECT building_id, 6,   500000, 25000, 151200, 5 FROM building_configs_base WHERE name = 'Archer Tower';

-- ── Air Defense ───────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, elixir_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,        0,      0,      0, 2 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 2,    15000,      0,   7200, 2 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 3,    45000,      0,  21600, 3 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 4,   120000,      0,  57600, 4 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 5,   320000,  12000,  86400, 4 FROM building_configs_base WHERE name = 'Air Defense'
UNION ALL
SELECT building_id, 6,   800000,  30000, 172800, 5 FROM building_configs_base WHERE name = 'Air Defense';

-- ── Gold Mine ─────────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,        0,      0, 1 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 2,      500,    900, 1 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 3,     2000,   3600, 1 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 4,     8000,  14400, 2 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 5,    30000,  43200, 3 FROM building_configs_base WHERE name = 'Gold Mine'
UNION ALL
SELECT building_id, 6,   100000, 108000, 4 FROM building_configs_base WHERE name = 'Gold Mine';

-- ── Gold Storage ──────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, gold_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,        0,      0, 1 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 2,     1000,   1800, 1 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 3,     5000,   7200, 2 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 4,    20000,  28800, 3 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 5,    80000,  86400, 4 FROM building_configs_base WHERE name = 'Gold Storage'
UNION ALL
SELECT building_id, 6,   300000, 172800, 5 FROM building_configs_base WHERE name = 'Gold Storage';

-- ── Elixir Collector ──────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, elixir_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,        0,      0, 1 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 2,      500,    900, 1 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 3,     2000,   3600, 1 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 4,     8000,  14400, 2 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 5,    30000,  43200, 3 FROM building_configs_base WHERE name = 'Elixir Collector'
UNION ALL
SELECT building_id, 6,   100000, 108000, 4 FROM building_configs_base WHERE name = 'Elixir Collector';

-- ── Elixir Storage ────────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, elixir_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,        0,      0, 1 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 2,     1000,   1800, 1 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 3,     5000,   7200, 2 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 4,    20000,  28800, 3 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 5,    80000,  86400, 4 FROM building_configs_base WHERE name = 'Elixir Storage'
UNION ALL
SELECT building_id, 6,   300000, 172800, 5 FROM building_configs_base WHERE name = 'Elixir Storage';

-- ── Dark Elixir Drill (unlocked at TH5; levels 5-6 cost heavy dark elixir) ──
INSERT INTO upgrade_costs (building_id, upgrade_to_level, dark_elixir_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,      0,      0, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 2,     50,   7200, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 3,    150,  21600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 4,    400,  57600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 5,   1200, 129600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Drill'
UNION ALL
SELECT building_id, 6,   3500, 259200, 6 FROM building_configs_base WHERE name = 'Dark Elixir Drill';

-- ── Dark Elixir Storage (unlocked at TH5) ─────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, dark_elixir_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,      0,      0, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 2,     60,   7200, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 3,    200,  21600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 4,    600,  57600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 5,   1800, 129600, 5 FROM building_configs_base WHERE name = 'Dark Elixir Storage'
UNION ALL
SELECT building_id, 6,   5000, 259200, 6 FROM building_configs_base WHERE name = 'Dark Elixir Storage';

-- ── Barracks (elixir) ─────────────────────────────────────────────────────────
INSERT INTO upgrade_costs (building_id, upgrade_to_level, elixir_required, time_required_seconds, town_hall_level_required)
SELECT building_id, 1,        0,      0, 1 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 2,     1500,   1800, 1 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 3,     6000,   7200, 2 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 4,    25000,  21600, 3 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 5,   100000,  57600, 4 FROM building_configs_base WHERE name = 'Barracks'
UNION ALL
SELECT building_id, 6,   400000, 129600, 5 FROM building_configs_base WHERE name = 'Barracks';