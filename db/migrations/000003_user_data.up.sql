    CREATE TYPE construction_type AS ENUM ('building_construction', 'building_upgrade', 'troop_training','building_repair');

CREATE TABLE placed_buildings
(
    id              UUID PRIMARY KEY          DEFAULT uuid_generate_v4(),
    user_id         UUID             NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    building_id     UUID             NOT NULL REFERENCES building_configs_base (building_id),
    grid_x          INT              NOT NULL,
    grid_y          INT              NOT NULL,
    level           non_negative_int NOT NULL DEFAULT 1,
    is_broken        BOOLEAN          NOT NULL DEFAULT FALSE,
    constructed_at  TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated_at TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE trained_troops
(
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID             NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    troop_id        UUID             NOT NULL REFERENCES troop_configs (id),
    level   non_negative_int NOT NULL DEFAULT 1,
    last_updated_at TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    count           non_negative_int NOT NULL DEFAULT 0
);

CREATE TABLE construction_tasks
(
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id            UUID              NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    task_type          construction_type NOT NULL,
    placed_building_id UUID              NOT NULL REFERENCES placed_buildings (id) ON DELETE CASCADE,
    troop_id           UUID,
    troop_count        non_negative_int,
    troop_level_to non_negative_int,
    started_at         TIMESTAMPTZ       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    duration_seconds   non_negative_int  NOT NULL,
    CONSTRAINT validate_task_context CHECK (
        (task_type = 'troop_training' AND troop_id IS NOT NULL) OR
        (task_type IN ('building_construction', 'building_upgrade','building_repair') AND troop_id IS NULL)
        ),
    FOREIGN KEY (troop_id) REFERENCES troop_configs (id) ON DELETE CASCADE
);

CREATE TABLE user_data
(
    user_id                    UUID PRIMARY KEY REFERENCES users (user_id) ON DELETE CASCADE,
    town_hall_level            town_hall_range  NOT NULL DEFAULT 1,
    current_gold               non_negative_int NOT NULL DEFAULT 0,
    current_elixir             non_negative_int NOT NULL DEFAULT 0,
    current_dark_elixir        non_negative_int NOT NULL DEFAULT 0,
    total_gold_capacity        non_negative_int NOT NULL DEFAULT 0,
    total_elixir_capacity      non_negative_int NOT NULL DEFAULT 0,
    total_dark_elixir_capacity non_negative_int NOT NULL DEFAULT 0,
    current_gems               non_negative_int NOT NULL DEFAULT 250,
    updated_at                 TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_placed_buildings_user ON placed_buildings (user_id);
CREATE INDEX idx_construction_tasks_user ON construction_tasks (user_id);
CREATE INDEX idx_construction_tasks_building_id ON construction_tasks (placed_building_id);
CREATE INDEX idx_trained_troops_user ON trained_troops (user_id);