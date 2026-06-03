DROP TABLE IF EXISTS building_configs CASCADE;

DROP TYPE IF EXISTS resource_type;
DROP TYPE IF EXISTS unit_target_type;
DROP TYPE IF EXISTS damage_type;
DROP TYPE IF EXISTS building_type;
DROP TYPE IF EXISTS building_category;
DROP TYPE IF EXISTS level_up_config;

DROP DOMAIN IF EXISTS town_hall_range;
DROP DOMAIN IF EXISTS non_negative_int;