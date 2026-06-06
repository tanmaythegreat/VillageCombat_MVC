CREATE TABLE battle_history
(
    battle_id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attacker_id        UUID NOT NULL,
    defender_id        UUID NOT NULL,
    elixir_looted      non_negative_int                      DEFAULT 0,
    gold_looted        non_negative_int                      DEFAULT 0,
    dark_elixir_looted non_negative_int                      DEFAULT 0,
    fought_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_attacker FOREIGN KEY (attacker_id) REFERENCES users (user_id),
    CONSTRAINT fk_defender FOREIGN KEY (defender_id) REFERENCES users (user_id)
);

CREATE TABLE battle_troop_losses (
     battle_id UUID NOT NULL,
     troop_id UUID NOT NULL,
     loss_count non_negative_int NOT NULL DEFAULT 0,

     PRIMARY KEY (battle_id, troop_id),
     CONSTRAINT fk_battle_troop FOREIGN KEY (battle_id) REFERENCES battle_history(battle_id),
     CONSTRAINT fk_troop_id FOREIGN KEY (troop_id) REFERENCES troop_configs(id)
);

CREATE TABLE buildings_broken (
  battle_id UUID NOT NULL,
  placed_building_id UUID NOT NULL,

  CONSTRAINT fk_battle_building FOREIGN KEY (battle_id) REFERENCES battle_history(battle_id) ON DELETE CASCADE,
  CONSTRAINT fk_placed_building FOREIGN KEY (placed_building_id) REFERENCES placed_buildings(id)
);

CREATE INDEX idx_battle_history_attacker_log ON battle_history (attacker_id, fought_at DESC);
CREATE INDEX idx_battle_history_defender_log ON battle_history (defender_id, fought_at DESC);
CREATE INDEX idx_buildings_broken_battle_id ON buildings_broken (battle_id);