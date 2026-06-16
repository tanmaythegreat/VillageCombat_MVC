DELETE FROM placed_buildings
WHERE building_id IN (
    SELECT building_id FROM building_configs_base
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
                   'Barracks'
        )
);

DELETE FROM upgrade_costs
WHERE building_id IN (
    SELECT building_id FROM building_configs_base
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
                   'Barracks'
        )
);

DELETE FROM upgrade_costs
WHERE troop_id IN (
    SELECT id FROM troop_configs
    WHERE name IN ('Barbarian', 'Archer', 'Goblin', 'Giant')
);

DELETE FROM army_building_level_stats
WHERE building_id IN (
    SELECT building_id FROM building_configs_base WHERE name = 'Barracks'
);

DELETE FROM army_building_stats
WHERE building_id IN (
    SELECT building_id FROM building_configs_base WHERE name = 'Barracks'
);

DELETE FROM resource_building_level_stats
WHERE building_id IN (
    SELECT building_id FROM building_configs_base
    WHERE name IN (
                   'Gold Mine',
                   'Gold Storage',
                   'Elixir Collector',
                   'Elixir Storage',
                   'Dark Elixir Drill',
                   'Dark Elixir Storage'
        )
);

DELETE FROM resource_building_stats
WHERE building_id IN (
    SELECT building_id FROM building_configs_base
    WHERE name IN (
                   'Gold Mine',
                   'Gold Storage',
                   'Elixir Collector',
                   'Elixir Storage',
                   'Dark Elixir Drill',
                   'Dark Elixir Storage'
        )
);

DELETE FROM defense_building_level_stats
WHERE building_id IN (
    SELECT building_id FROM building_configs_base
    WHERE name IN ('Cannon', 'Archer Tower', 'Air Defense')
);

DELETE FROM defense_building_stats
WHERE building_id IN (
    SELECT building_id FROM building_configs_base
    WHERE name IN ('Cannon', 'Archer Tower', 'Air Defense')
);

DELETE FROM building_level_stats
WHERE building_id IN (
    SELECT building_id FROM building_configs_base
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
                   'Barracks'
        )
);

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
               'Barracks'
    );

DELETE FROM troop_level_stats
WHERE troop_id IN (
    SELECT id FROM troop_configs
    WHERE name IN ('Barbarian', 'Archer', 'Goblin', 'Giant')
);

DELETE FROM troop_configs
WHERE name IN ('Barbarian', 'Archer', 'Goblin', 'Giant');