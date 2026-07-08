-- =============================================================================
-- NEW TROOP CONFIGS
-- =============================================================================
INSERT INTO troop_configs (name, preferred_target, attack_type, movement_speed, attack_speed_seconds, attack_range, housing_space)
VALUES
    ('Wizard',       NULL,       'ranged'::attack_type, 1.0, 1.2, 3.5, 4),
    ('Wall Breaker', 'wall',     'melee'::attack_type,  1.4, 1.0, 1.0, 2),
    ('Balloon',      'defense',  'melee'::attack_type,  0.5, 2.0, 1.0, 5),
    ('Minion',       NULL,       'ranged'::attack_type, 1.6, 1.0, 2.0, 2);

-- =============================================================================
-- TROOP LEVEL STATS
-- =============================================================================

-- Wizard (glass cannon, high splash damage)
INSERT INTO troop_level_stats (troop_id, level, health, damage_per_shot)
SELECT id, 0,    0,   0 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 1,  130, 140 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 2,  150, 170 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 3,  180, 210 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 4,  220, 260 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 5,  270, 320 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 6,  330, 390 FROM troop_configs WHERE name = 'Wizard';

-- Wall Breaker (fragile, huge burst — meant to die after one hit vs walls in real game,
-- but keeping consistent per-shot damage here since your schema doesn't have "single-use")
INSERT INTO troop_level_stats (troop_id, level, health, damage_per_shot)
SELECT id, 0,    0,   0 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 1,   40, 300 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 2,   50, 380 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 3,   65, 480 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 4,   80, 600 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 5,  100, 750 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 6,  125, 940 FROM troop_configs WHERE name = 'Wall Breaker';

-- Balloon (air, high HP high damage, slow)
INSERT INTO troop_level_stats (troop_id, level, health, damage_per_shot)
SELECT id, 0,    0,   0 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 1, 1200, 150 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 2, 1450, 190 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 3, 1750, 240 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 4, 2100, 300 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 5, 2550, 375 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 6, 3100, 470 FROM troop_configs WHERE name = 'Balloon';

-- Minion (fast, fragile, ranged air)
INSERT INTO troop_level_stats (troop_id, level, health, damage_per_shot)
SELECT id, 0,    0,   0 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 1,  100,  55 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 2,  120,  70 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 3,  145,  90 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 4,  175, 115 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 5,  210, 145 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 6,  250, 180 FROM troop_configs WHERE name = 'Minion';


-- =============================================================================
-- UPGRADE COSTS
-- =============================================================================

-- Wizard (elixir + gems, pricier than Barbarian/Archer — strongest DPS unit)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT id, 1,      6000,         6,   1800, 2 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 2,     30000,        30,   7200, 2 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 3,    300000,       300,  28800, 3 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 4,   4500000,      4500,  72000, 4 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 5,  90000000,     90000, 129600, 5 FROM troop_configs WHERE name = 'Wizard'
UNION ALL SELECT id, 6, 2250000000,  2250000, 216000, 6 FROM troop_configs WHERE name = 'Wizard';

-- Wall Breaker (cheap elixir, unlocked early since it's a utility unit)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT id, 1,      1000,         1,    600, 1 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 2,      5000,         5,   2400, 1 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 3,     50000,        50,   9600, 2 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 4,    750000,       750,  28800, 3 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 5,  15000000,     15000,  57600, 4 FROM troop_configs WHERE name = 'Wall Breaker'
UNION ALL SELECT id, 6, 375000000,    375000, 115200, 5 FROM troop_configs WHERE name = 'Wall Breaker';

-- Balloon (elixir + gems, unlocked at TH4 like real game)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT id, 1,      8000,         8,   3600, 3 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 2,     40000,        40,  14400, 3 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 3,    400000,       400,  43200, 4 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 4,   6000000,      6000,  86400, 4 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 5, 120000000,    120000, 151200, 5 FROM troop_configs WHERE name = 'Balloon'
UNION ALL SELECT id, 6, 3000000000,  3000000, 259200, 6 FROM troop_configs WHERE name = 'Balloon';

-- Minion (dark elixir based, unlocked TH5+ like real game)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, dark_elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT id, 1,        20,         2,   3600, 5 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 2,        40,         4,  10800, 5 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 3,       400,        40,  36000, 5 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 4,      6000,       600,  72000, 5 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 5,    120000,     12000, 144000, 6 FROM troop_configs WHERE name = 'Minion'
UNION ALL SELECT id, 6,   3000000,    300000, 216000, 6 FROM troop_configs WHERE name = 'Minion';

-- =============================================================================
-- CUSTOM TROOP: Doraemon (fan-fun unit — supportive, tanky, gadget-themed)
-- =============================================================================
INSERT INTO troop_configs (name, preferred_target, attack_type, movement_speed, attack_speed_seconds, attack_range, housing_space)
VALUES
    ('Doraemon', 'defense', 'ranged'::attack_type, 0.8, 1.5, 4.0, 3);

-- Tanky support unit, moderate damage, decent range ("gadget" theme)
INSERT INTO troop_level_stats (troop_id, level, health, damage_per_shot)
SELECT id, 0,    0,   0 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 1,  600, 90 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 2,  720, 115 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 3,  860, 145 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 4, 1040, 180 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 5, 1250, 225 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 6, 1500, 280 FROM troop_configs WHERE name = 'Doraemon';

-- Upgrade costs (elixir + dark elixir, unlocked TH3, premium unit)
INSERT INTO upgrade_costs (troop_id, upgrade_to_level, elixir_required, dark_elixir_required, or_gem_required, time_required_seconds, town_hall_level_required)
SELECT id, 1,     20000,      0,        20,   3600, 3 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 2,    100000,      0,       100,  10800, 3 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 3,   1000000,      0,      1000,  36000, 4 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 4,  15000000,      0,     15000,  72000, 5 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 5,         0,   6000,       600, 144000, 6 FROM troop_configs WHERE name = 'Doraemon'
UNION ALL SELECT id, 6,         0,  60000,      6000, 216000, 6 FROM troop_configs WHERE name = 'Doraemon';
