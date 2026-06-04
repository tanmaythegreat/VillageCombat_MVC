package Models

import (
	"encoding/json"
	"time"
)

type AttackType string

const (
	Melee  AttackType = "melee"
	Ranged AttackType = "ranged"
)

type BuildingCategory string

const (
	TownHall BuildingCategory = "townhall"
	Defense  BuildingCategory = "defense"
	Resource BuildingCategory = "resource"
	Army     BuildingCategory = "army"
)

type DamageType string

const (
	SingleTarget DamageType = "single_target"
	Splash       DamageType = "splash"
)

type UnitTargetType string

const (
	Ground       UnitTargetType = "ground"
	GroundAndAir UnitTargetType = "ground_and_air"
	Air          UnitTargetType = "air"
)

type ResourceType string

const (
	Gold       ResourceType = "gold"
	Elixir     ResourceType = "elixir"
	DarkElixir ResourceType = "dark_elixir"
)

type ConstructionType string

const (
	BuildingConstruction ConstructionType = "building_construction"
	BuildingUpgrade      ConstructionType = "building_upgrade"
	TroopTraining        ConstructionType = "troop_training"
)

type User struct {
	UserID       string    `gorm:"column:user_id;primaryKey" json:"user_id"`
	Username     string    `gorm:"column:username" json:"username"`
	Email        string    `gorm:"column:email" json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type RefreshToken struct {
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	UserID       string    `gorm:"column:user_id" json:"user_id"`
	JWTTokenHash string    `gorm:"column:jwt_token_hash" json:"-"`
	IPAddress    string    `gorm:"column:ip_address" json:"ip_address"`
	UserAgent    string    `gorm:"column:user_agent" json:"user_agent"`
	ExpiresAt    time.Time `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	IsUsed       bool      `gorm:"column:is_used" json:"is_used"`
}

type TroopConfig struct {
	ID                 string            `gorm:"column:id;primaryKey" json:"id"`
	Name               string            `gorm:"column:name" json:"name"`
	PreferredTarget    *BuildingCategory `gorm:"column:preferred_target" json:"preferred_target,omitempty"`
	UnlockAtLevel      int               `gorm:"column:unlock_at_level" json:"unlock_at_level"`
	AttackType         AttackType        `gorm:"column:attack_type" json:"attack_type"`
	MovementSpeed      float64           `gorm:"column:movement_speed" json:"movement_speed"`
	AttackSpeedSeconds float64           `gorm:"column:attack_speed_seconds" json:"attack_speed_seconds"`
	AttackRange        float64           `gorm:"column:attack_range" json:"attack_range"`
	HousingSpace       int               `gorm:"column:housing_space" json:"housing_space"`
	LevelStats         []TroopLevelStats `gorm:"foreignKey:TroopID" json:"level_stats,omitempty"`
	UpgradeCosts       []UpgradeCost     `gorm:"foreignKey:TroopID" json:"upgrade_costs,omitempty"`
}

type TroopLevelStats struct {
	TroopID       string `gorm:"column:troop_id;primaryKey" json:"troop_id"`
	Level         int    `gorm:"column:level;primaryKey" json:"level"`
	Health        int    `gorm:"column:health" json:"health"`
	DamagePerShot int    `gorm:"column:damage_per_shot" json:"damage_per_shot"`
}

type UpgradeCost struct {
	ID                    string  `gorm:"column:id;primaryKey" json:"id"`
	TroopID               *string `gorm:"column:troop_id" json:"troop_id,omitempty"`
	BuildingID            *string `gorm:"column:building_id" json:"building_id,omitempty"`
	UpgradeToLevel        int     `gorm:"column:upgrade_to_level" json:"upgrade_to_level"`
	GoldRequired          int     `gorm:"column:gold_required" json:"gold_required"`
	ElixirRequired        int     `gorm:"column:elixir_required" json:"elixir_required"`
	DarkElixirRequired    int     `gorm:"column:dark_elixir_required" json:"dark_elixir_required"`
	OrGemRequired         int     `gorm:"column:or_gem_required" json:"or_gem_required"`
	TimeRequiredSeconds   int     `gorm:"column:time_required_seconds" json:"time_required_seconds"`
	TownHallLevelRequired int     `gorm:"column:town_hall_level_required" json:"town_hall_level_required"`
}

type BuildingConfigBase struct {
	BuildingID   string               `gorm:"column:building_id;primaryKey" json:"building_id"`
	Name         string               `gorm:"column:name" json:"name"`
	Category     BuildingCategory     `gorm:"column:category" json:"category"`
	GridSizeX    int                  `gorm:"column:grid_size_x" json:"grid_size_x"`
	GridSizeY    int                  `gorm:"column:grid_size_y" json:"grid_size_y"`
	LevelStats   []BuildingLevelStats `gorm:"foreignKey:BuildingID" json:"level_stats,omitempty"`
	UpgradeCosts []UpgradeCost        `gorm:"foreignKey:BuildingID" json:"upgrade_costs,omitempty"`
}

type BuildingLevelStats struct {
	BuildingID string `gorm:"column:building_id;primaryKey" json:"building_id"`
	Level      int    `gorm:"column:level;primaryKey" json:"level"`
	Health     int    `gorm:"column:health" json:"health"`
}

type DefenseBuildingStats struct {
	BuildingID         string                      `gorm:"column:building_id;primaryKey" json:"building_id"`
	BuildingType       string                      `gorm:"column:building_type" json:"building_type"`
	AttackSpeedSeconds float64                     `gorm:"column:attack_speed_seconds" json:"attack_speed_seconds"`
	AttackRange        float64                     `gorm:"column:attack_range" json:"attack_range"`
	DamageType         DamageType                  `gorm:"column:damage_type" json:"damage_type"`
	UnitTarget         UnitTargetType              `gorm:"column:unit_target" json:"unit_target"`
	LevelStats         []DefenseBuildingLevelStats `gorm:"foreignKey:BuildingID" json:"level_stats,omitempty"`
}

type DefenseBuildingLevelStats struct {
	BuildingID    string `gorm:"column:building_id;primaryKey" json:"building_id"`
	Level         int    `gorm:"column:level;primaryKey" json:"level"`
	DamagePerShot int    `gorm:"column:damage_per_shot" json:"damage_per_shot"`
}

type ResourceBuildingStats struct {
	BuildingID   string                       `gorm:"column:building_id;primaryKey" json:"building_id"`
	BuildingType string                       `gorm:"column:building_type" json:"building_type"`
	ResourceType ResourceType                 `gorm:"column:resource_type" json:"resource_type"`
	LevelStats   []ResourceBuildingLevelStats `gorm:"foreignKey:BuildingID" json:"level_stats,omitempty"`
}

type ResourceBuildingLevelStats struct {
	BuildingID            string  `gorm:"column:building_id;primaryKey" json:"building_id"`
	Level                 int     `gorm:"column:level;primaryKey" json:"level"`
	GenerationRatePerHour float64 `gorm:"column:generation_rate_per_hour" json:"generation_rate_per_hour"`
	StorageCapacity       int     `gorm:"column:storage_capacity" json:"storage_capacity"`
}

type ArmyBuildingStats struct {
	BuildingID   string                   `gorm:"column:building_id;primaryKey" json:"building_id"`
	BuildingType string                   `gorm:"column:building_type" json:"building_type"`
	LevelStats   []ArmyBuildingLevelStats `gorm:"foreignKey:BuildingID" json:"level_stats,omitempty"`
}

type ArmyBuildingLevelStats struct {
	BuildingID    string `gorm:"column:building_id;primaryKey" json:"building_id"`
	Level         int    `gorm:"column:level;primaryKey" json:"level"`
	TroopCapacity int    `gorm:"column:troop_capacity" json:"troop_capacity"`
}

type PlacedBuilding struct {
	ID            string          `gorm:"column:id;primaryKey" json:"id"`
	UserID        string          `gorm:"column:user_id" json:"user_id"`
	BuildingID    string          `gorm:"column:building_id" json:"building_id"`
	GridX         int             `gorm:"column:grid_x" json:"grid_x"`
	GridY         int             `gorm:"column:grid_y" json:"grid_y"`
	CurrentLevel  int             `gorm:"column:current_level" json:"current_level"`
	DynamicState  json.RawMessage `gorm:"column:dynamic_state;type:jsonb" json:"dynamic_state"`
	ConstructedAt time.Time       `gorm:"column:constructed_at" json:"constructed_at"`
	LastUpdatedAt time.Time       `gorm:"column:last_updated_at" json:"last_updated_at"`
}

type TrainedTroop struct {
	ID            string          `gorm:"column:id;primaryKey" json:"id"`
	UserID        string          `gorm:"column:user_id" json:"user_id"`
	TroopID       string          `gorm:"column:troop_id" json:"troop_id"`
	CurrentLevel  int             `gorm:"column:current_level" json:"current_level"`
	DynamicState  json.RawMessage `gorm:"column:dynamic_state;type:jsonb" json:"dynamic_state"`
	LastUpdatedAt time.Time       `gorm:"column:last_updated_at" json:"last_updated_at"`
	Count         int             `gorm:"count" json:"count"`
}

type ConstructionTask struct {
	ID               string           `gorm:"column:id;primaryKey" json:"id"`
	UserID           string           `gorm:"column:user_id" json:"user_id"`
	TaskType         ConstructionType `gorm:"column:task_type" json:"task_type"`
	PlacedBuildingID string           `gorm:"column:placed_building_id" json:"placed_building_id"`
	TroopID          *string          `gorm:"column:troop_id" json:"troop_id,omitempty"`
	StartedAt        time.Time        `gorm:"column:started_at" json:"started_at"`
	DurationSeconds  int              `gorm:"column:duration_seconds" json:"duration_seconds"`
}

type UserData struct {
	UserID            string    `gorm:"column:user_id;primaryKey" json:"user_id"`
	TownHallLevel     int       `gorm:"column:town_hall_level" json:"town_hall_level"`
	CurrentGold       int       `gorm:"column:current_gold" json:"current_gold"`
	CurrentElixir     int       `gorm:"column:current_elixir" json:"current_elixir"`
	CurrentDarkElixir int       `gorm:"column:current_dark_elixir" json:"current_dark_elixir"`
	CurrentGems       int       `gorm:"column:current_gems" json:"current_gems"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type BattleHistory struct {
	BattleID         string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:battle_id"`
	AttackerID       string    `gorm:"type:uuid;not null;column:attacker_id"`
	DefenderID       string    `gorm:"type:uuid;not null;column:defender_id"`
	ElixirLooted     int       `gorm:"default:0;column:elixir_looted"`
	GoldLooted       int       `gorm:"default:0;column:gold_looted"`
	DarkElixirLooted int       `gorm:"default:0;column:dark_elixir_looted"`
	FoughtAt         time.Time `gorm:"default:CURRENT_TIMESTAMP;column:fought_at"`

	TroopLosses     []BattleTroopLoss `gorm:"foreignKey:BattleID"`
	BrokenBuildings []BuildingsBroken `gorm:"foreignKey:BattleID;constraint:OnDelete:CASCADE"`
}

type BattleTroopLoss struct {
	BattleID  string `gorm:"type:uuid;primaryKey;column:battle_id"`
	TroopID   string `gorm:"type:uuid;primaryKey;column:troop_id"`
	LossCount int    `gorm:"not null;default:0;column:loss_count"`
}

type BuildingsBroken struct {
	BattleID         string `gorm:"type:uuid;primaryKey;column:battle_id"`
	PlacedBuildingID string `gorm:"type:uuid;primaryKey;column:placed_building_id"`
}
