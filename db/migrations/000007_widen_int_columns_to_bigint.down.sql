DROP DOMAIN IF EXISTS non_negative_bigint CASCADE;

-- ---------------------------------------------------------------------------
-- troop_level_stats.health
-- ---------------------------------------------------------------------------
ALTER TABLE troop_level_stats
    ALTER COLUMN health TYPE non_negative_int USING health::int;

-- ---------------------------------------------------------------------------
-- building_level_stats.health
-- ---------------------------------------------------------------------------
ALTER TABLE building_level_stats
    ALTER COLUMN health TYPE non_negative_int USING health::int;

-- ---------------------------------------------------------------------------
-- upgrade_costs — all currency columns
-- ---------------------------------------------------------------------------
ALTER TABLE upgrade_costs
    ALTER COLUMN gold_required        TYPE non_negative_int USING gold_required::int,
    ALTER COLUMN elixir_required      TYPE non_negative_int USING elixir_required::int,
    ALTER COLUMN dark_elixir_required TYPE non_negative_int USING dark_elixir_required::int,
    ALTER COLUMN or_gem_required      TYPE non_negative_int USING or_gem_required::int;

-- ---------------------------------------------------------------------------
-- resource_building_level_stats.storage_capacity
-- ---------------------------------------------------------------------------
ALTER TABLE resource_building_level_stats
    ALTER COLUMN storage_capacity TYPE non_negative_int USING storage_capacity::int;

-- ---------------------------------------------------------------------------
-- user_data — all currency columns
-- ---------------------------------------------------------------------------
ALTER TABLE user_data
    ALTER COLUMN current_gold               TYPE non_negative_int USING current_gold::int,
    ALTER COLUMN current_elixir             TYPE non_negative_int USING current_elixir::int,
    ALTER COLUMN current_dark_elixir        TYPE non_negative_int USING current_dark_elixir::int,
    ALTER COLUMN total_gold_capacity        TYPE non_negative_int USING total_gold_capacity::int,
    ALTER COLUMN total_elixir_capacity      TYPE non_negative_int USING total_elixir_capacity::int,
    ALTER COLUMN total_dark_elixir_capacity TYPE non_negative_int USING total_dark_elixir_capacity::int,
    ALTER COLUMN current_gems               TYPE non_negative_int USING current_gems::int;

-- ---------------------------------------------------------------------------
-- battle_history — looted currency columns
-- ---------------------------------------------------------------------------
ALTER TABLE battle_history
    ALTER COLUMN elixir_looted      TYPE non_negative_int USING elixir_looted::int,
    ALTER COLUMN gold_looted        TYPE non_negative_int USING gold_looted::int,
    ALTER COLUMN dark_elixir_looted TYPE non_negative_int USING dark_elixir_looted::int;