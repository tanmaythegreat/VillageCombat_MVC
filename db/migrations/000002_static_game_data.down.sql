DROP TABLE IF EXISTS army_building_level_stats CASCADE;
DROP TABLE IF EXISTS army_building_stats CASCADE;
DROP TABLE IF EXISTS resource_building_level_stats CASCADE;
DROP TABLE IF EXISTS resource_building_stats CASCADE;
DROP TABLE IF EXISTS defense_building_level_stats CASCADE;
DROP TABLE IF EXISTS defense_building_stats CASCADE;
DROP TABLE IF EXISTS upgrade_costs CASCADE;
DROP TABLE IF EXISTS building_level_stats CASCADE;
DROP TABLE IF EXISTS troop_level_stats CASCADE;
DROP TABLE IF EXISTS building_configs_base CASCADE;
DROP TABLE IF EXISTS troop_configs CASCADE;

DROP TYPE IF EXISTS resource_type CASCADE;
DROP TYPE IF EXISTS unit_target_type CASCADE;
DROP TYPE IF EXISTS damage_type CASCADE;
DROP TYPE IF EXISTS building_category CASCADE;
DROP TYPE IF EXISTS attack_type CASCADE;

DROP DOMAIN IF EXISTS non_negative_numeric CASCADE;
DROP DOMAIN IF EXISTS town_hall_range CASCADE;
DROP DOMAIN IF EXISTS non_negative_int CASCADE;