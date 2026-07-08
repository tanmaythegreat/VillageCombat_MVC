-- =============================================================================
-- MIGRATION: widen health + currency columns to non_negative_bigint
-- =============================================================================
DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1 FROM pg_type WHERE typname = 'non_negative_bigint'
        ) THEN
            CREATE DOMAIN non_negative_bigint AS BIGINT CHECK (VALUE >= 0);
        END IF;
    END $$;
-- ---------------------------------------------------------------------------
-- troop_level_stats.health
-- ---------------------------------------------------------------------------
ALTER TABLE troop_level_stats
    ALTER COLUMN health TYPE non_negative_bigint USING health::bigint;

-- ---------------------------------------------------------------------------
-- building_level_stats.health
-- ---------------------------------------------------------------------------
ALTER TABLE building_level_stats
    ALTER COLUMN health TYPE non_negative_bigint USING health::bigint;

-- ---------------------------------------------------------------------------
-- upgrade_costs — all currency columns
-- ---------------------------------------------------------------------------
ALTER TABLE upgrade_costs
    ALTER COLUMN gold_required        TYPE non_negative_bigint USING gold_required::bigint,
    ALTER COLUMN elixir_required      TYPE non_negative_bigint USING elixir_required::bigint,
    ALTER COLUMN dark_elixir_required TYPE non_negative_bigint USING dark_elixir_required::bigint,
    ALTER COLUMN or_gem_required      TYPE non_negative_bigint USING or_gem_required::bigint;

-- ---------------------------------------------------------------------------
-- resource_building_level_stats.storage_capacity
-- ---------------------------------------------------------------------------
ALTER TABLE resource_building_level_stats
    ALTER COLUMN storage_capacity TYPE non_negative_bigint USING storage_capacity::bigint;

-- ---------------------------------------------------------------------------
-- user_data — all currency columns
-- ---------------------------------------------------------------------------
ALTER TABLE user_data
    ALTER COLUMN current_gold               TYPE non_negative_bigint USING current_gold::bigint,
    ALTER COLUMN current_elixir             TYPE non_negative_bigint USING current_elixir::bigint,
    ALTER COLUMN current_dark_elixir        TYPE non_negative_bigint USING current_dark_elixir::bigint,
    ALTER COLUMN total_gold_capacity        TYPE non_negative_bigint USING total_gold_capacity::bigint,
    ALTER COLUMN total_elixir_capacity      TYPE non_negative_bigint USING total_elixir_capacity::bigint,
    ALTER COLUMN total_dark_elixir_capacity TYPE non_negative_bigint USING total_dark_elixir_capacity::bigint,
    ALTER COLUMN current_gems               TYPE non_negative_bigint USING current_gems::bigint;

-- ---------------------------------------------------------------------------
-- battle_history — looted currency columns
-- ---------------------------------------------------------------------------
ALTER TABLE battle_history
    ALTER COLUMN elixir_looted      TYPE non_negative_bigint USING elixir_looted::bigint,
    ALTER COLUMN gold_looted        TYPE non_negative_bigint USING gold_looted::bigint,
    ALTER COLUMN dark_elixir_looted TYPE non_negative_bigint USING dark_elixir_looted::bigint;