DELETE FROM placed_buildings;
DELETE FROM army_building_level_stats;
DELETE FROM army_building_stats;

DELETE FROM resource_building_level_stats;
DELETE FROM resource_building_stats;

DELETE FROM defense_building_level_stats;
DELETE FROM defense_building_stats;

DELETE FROM building_level_stats;

DELETE FROM upgrade_costs;

DELETE FROM troop_level_stats;

DELETE FROM building_configs_base
WHERE name IN (
               'Town Hall',
               'Cannon',
               'Archer Tower',
               'Air Defense',
               'Gold Mine',
               'Gold Storage',
               'Elixir Collector',
               'Elixir Storage',
               'Dark Elixir Drill',
               'Dark Elixir Storage',
               'Barracks',
               'Wall'
    );

DELETE FROM troop_configs
WHERE name IN (
               'Barbarian',
               'Archer',
               'Goblin',
               'Giant'
    );