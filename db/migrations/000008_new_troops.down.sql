-- =============================================================================
-- DOWN MIGRATION
-- =============================================================================

DELETE FROM upgrade_costs
WHERE troop_id IN (
    SELECT id
    FROM troop_configs
    WHERE name IN (
                   'Wizard',
                   'Wall Breaker',
                   'Balloon',
                   'Minion',
                   'Doraemon',
                   'Shinchan'
        )
);

DELETE FROM troop_level_stats
WHERE troop_id IN (
    SELECT id
    FROM troop_configs
    WHERE name IN (
                   'Wizard',
                   'Wall Breaker',
                   'Balloon',
                   'Minion',
                   'Doraemon',
                   'Shinchan'
        )
);

DELETE FROM troop_configs
WHERE name IN (
               'Wizard',
               'Wall Breaker',
               'Balloon',
               'Minion',
               'Doraemon',
               'Shinchan'
    );