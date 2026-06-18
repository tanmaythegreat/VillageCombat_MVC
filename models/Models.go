package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	Wall     BuildingCategory = "wall"
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
	BauildingRepair      ConstructionType = "building_repair"
)

type User struct {
	UserID       string    `gorm:"column:user_id;primaryKey" json:"user_id"`
	Username     string    `gorm:"column:username" json:"username"`
	Email        string    `gorm:"column:email" json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (User) TableName() string { return "users" }

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

func (RefreshToken) TableName() string { return "refresh_tokens" }

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

func (TroopConfig) TableName() string { return "troop_configs" }

type TroopLevelStats struct {
	TroopID       string `gorm:"column:troop_id;primaryKey" json:"troop_id"`
	Level         int    `gorm:"column:level;primaryKey" json:"level"`
	Health        int    `gorm:"column:health" json:"health"`
	DamagePerShot int    `gorm:"column:damage_per_shot" json:"damage_per_shot"`
}

func (TroopLevelStats) TableName() string { return "troop_level_stats" }

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

func (UpgradeCost) TableName() string { return "upgrade_costs" }

type BuildingConfigBase struct {
	BuildingID   string               `gorm:"column:building_id;primaryKey" json:"building_id"`
	Name         string               `gorm:"column:name" json:"name"`
	Category     BuildingCategory     `gorm:"column:category" json:"category"`
	GridSizeX    int                  `gorm:"column:grid_size_x" json:"grid_size_x"`
	GridSizeY    int                  `gorm:"column:grid_size_y" json:"grid_size_y"`
	LevelStats   []BuildingLevelStats `gorm:"foreignKey:BuildingID" json:"level_stats,omitempty"`
	UpgradeCosts []UpgradeCost        `gorm:"foreignKey:BuildingID" json:"upgrade_costs,omitempty"`
}

func (BuildingConfigBase) TableName() string { return "building_configs_base" }

type BuildingLevelStats struct {
	BuildingID string `gorm:"column:building_id;primaryKey" json:"building_id"`
	Level      int    `gorm:"column:level;primaryKey" json:"level"`
	Health     int    `gorm:"column:health" json:"health"`
}

func (BuildingLevelStats) TableName() string { return "building_level_stats" }

type DefenseBuildingStats struct {
	BuildingID         string                      `gorm:"column:building_id;primaryKey" json:"building_id"`
	AttackSpeedSeconds float64                     `gorm:"column:attack_speed_seconds" json:"attack_speed_seconds"`
	AttackRange        float64                     `gorm:"column:attack_range" json:"attack_range"`
	DamageType         DamageType                  `gorm:"column:damage_type" json:"damage_type"`
	UnitTarget         UnitTargetType              `gorm:"column:unit_target" json:"unit_target"`
	LevelStats         []DefenseBuildingLevelStats `gorm:"foreignKey:BuildingID" json:"level_stats,omitempty"`
}

func (DefenseBuildingStats) TableName() string { return "defense_building_stats" }

type DefenseBuildingLevelStats struct {
	BuildingID    string `gorm:"column:building_id;primaryKey" json:"building_id"`
	Level         int    `gorm:"column:level;primaryKey" json:"level"`
	DamagePerShot int    `gorm:"column:damage_per_shot" json:"damage_per_shot"`
}

func (DefenseBuildingLevelStats) TableName() string { return "defense_building_level_stats" }

type ResourceBuildingStats struct {
	BuildingID   string                       `gorm:"column:building_id;primaryKey" json:"building_id"`
	ResourceType ResourceType                 `gorm:"column:resource_type" json:"resource_type"`
	LevelStats   []ResourceBuildingLevelStats `gorm:"foreignKey:BuildingID" json:"level_stats,omitempty"`
}

func (ResourceBuildingStats) TableName() string { return "resource_building_stats" }

type ResourceBuildingLevelStats struct {
	BuildingID            string  `gorm:"column:building_id;primaryKey" json:"building_id"`
	Level                 int     `gorm:"column:level;primaryKey" json:"level"`
	GenerationRatePerHour float64 `gorm:"column:generation_rate_per_hour" json:"generation_rate_per_hour"`
	StorageCapacity       int     `gorm:"column:storage_capacity" json:"storage_capacity"`
}

func (ResourceBuildingLevelStats) TableName() string { return "resource_building_level_stats" }

type ArmyBuildingStats struct {
	BuildingID string                   `gorm:"column:building_id;primaryKey" json:"building_id"`
	LevelStats []ArmyBuildingLevelStats `gorm:"foreignKey:BuildingID" json:"level_stats,omitempty"`
}

func (ArmyBuildingStats) TableName() string { return "army_building_stats" }

type ArmyBuildingLevelStats struct {
	BuildingID    string `gorm:"column:building_id;primaryKey" json:"building_id"`
	Level         int    `gorm:"column:level;primaryKey" json:"level"`
	TroopCapacity int    `gorm:"column:troop_capacity" json:"troop_capacity"`
}

func (ArmyBuildingLevelStats) TableName() string { return "army_building_level_stats" }

type PlacedBuilding struct {
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
	UserID        string    `gorm:"column:user_id" json:"user_id"`
	BuildingID    string    `gorm:"column:building_id" json:"building_id"`
	GridX         int       `gorm:"column:grid_x" json:"grid_x"`
	GridY         int       `gorm:"column:grid_y" json:"grid_y"`
	Level         int       `gorm:"column:level" json:"level"`
	IsBroken      bool      `gorm:"column:is_broken" json:"is_broken"`
	ConstructedAt time.Time `gorm:"column:constructed_at" json:"constructed_at"`
	LastUpdatedAt time.Time `gorm:"column:last_updated_at" json:"last_updated_at"`
}

func (PlacedBuilding) TableName() string { return "placed_buildings" }

type TrainedTroop struct {
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
	UserID        string    `gorm:"column:user_id" json:"user_id"`
	TroopID       string    `gorm:"column:troop_id" json:"troop_id"`
	Level         int       `gorm:"column:level" json:"level"`
	Count         int       `gorm:"column:count" json:"count"`
	LastUpdatedAt time.Time `gorm:"column:last_updated_at" json:"last_updated_at"`
}

func (TrainedTroop) TableName() string { return "trained_troops" }

type ConstructionTask struct {
	ID               string           `gorm:"column:id;primaryKey" json:"id"`
	UserID           string           `gorm:"column:user_id" json:"user_id"`
	TaskType         ConstructionType `gorm:"column:task_type" json:"task_type"`
	PlacedBuildingID string           `gorm:"column:placed_building_id" json:"placed_building_id"`
	TroopID          *string          `gorm:"column:troop_id" json:"troop_id,omitempty"`
	TroopCount       *int             `gorm:"column:troop_count" json:"troop_count"`
	TroopLevelTo     *int             `gorm:"column:troop_level_to" json:"troop_level_to"`
	StartedAt        time.Time        `gorm:"column:started_at" json:"started_at"`
	DurationSeconds  int              `gorm:"column:duration_seconds" json:"duration_seconds"`
}

func (ConstructionTask) TableName() string { return "construction_tasks" }

type UserData struct {
	UserID                  string    `gorm:"column:user_id;primaryKey" json:"user_id"`
	TownHallLevel           int       `gorm:"column:town_hall_level" json:"town_hall_level"`
	CurrentGold             int       `gorm:"column:current_gold" json:"current_gold"`
	CurrentElixir           int       `gorm:"column:current_elixir" json:"current_elixir"`
	CurrentDarkElixir       int       `gorm:"column:current_dark_elixir" json:"current_dark_elixir"`
	TotalGoldCapacity       int       `gorm:"column:total_gold_capacity" json:"total_gold_capacity"`
	TotalElixirCapacity     int       `gorm:"column:total_elixir_capacity" json:"total_elixir_capacity"`
	TotalDarkElixirCapacity int       `gorm:"column:total_dark_elixir_capacity" json:"total_dark_elixir_capacity"`
	CurrentGems             int       `gorm:"column:current_gems" json:"current_gems"`
	UpdatedAt               time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserData) TableName() string { return "user_data" }

type BattleHistory struct {
	BattleID         string            `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:battle_id" json:"battle_id"`
	AttackerID       string            `gorm:"type:uuid;not null;column:attacker_id" json:"attacker_id"`
	DefenderID       string            `gorm:"type:uuid;not null;column:defender_id" json:"defender_id"`
	ElixirLooted     int               `gorm:"default:0;column:elixir_looted" json:"elixir_looted"`
	GoldLooted       int               `gorm:"default:0;column:gold_looted" json:"gold_looted"`
	DarkElixirLooted int               `gorm:"default:0;column:dark_elixir_looted" json:"dark_elixir_looted"`
	FoughtAt         time.Time         `gorm:"default:CURRENT_TIMESTAMP;column:fought_at" json:"fought_at"`
	BattleDuration   int               `gorm:"column:battle_duration" json:"battle_duration"`
	DoDefenderKnow   bool              `gorm:"column:do_defender_know" json:"do_defender_know"`
	TroopLosses      []BattleTroopLoss `gorm:"foreignKey:BattleID" json:"troop_losses,omitempty"`
	BrokenBuildings  []BuildingsBroken `gorm:"foreignKey:BattleID;constraint:OnDelete:CASCADE" json:"broken_buildings,omitempty"`
}

func (BattleHistory) TableName() string { return "battle_history" }

type BattleTroopLoss struct {
	BattleID   string `gorm:"type:uuid;primaryKey;column:battle_id" json:"battle_id"`
	TroopID    string `gorm:"type:uuid;primaryKey;column:troop_id" json:"troop_id"`
	LossCount  int    `gorm:"not null;default:0;column:loss_count" json:"loss_count"`
	IsAttacker bool   `gorm:"column:is_attacker" json:"is_attacker"`
}

func (BattleTroopLoss) TableName() string { return "battle_troop_losses" }

type BuildingsBroken struct {
	BattleID   string `gorm:"type:uuid;primaryKey;column:battle_id" json:"battle_id"`
	BuildingID string `gorm:"type:uuid;column:building_id" json:"building_id"`
	Count      int    `gorm:"column:count" json:"count"`
}

func (BuildingsBroken) TableName() string { return "buildings_broken" }

type UserStatus struct {
	UserID       string    `gorm:"column:user_id;primaryKey" json:"user_id"`
	LastDefended time.Time `gorm:"column:last_defended" json:"last_defended"`
	InBattle     bool      `gorm:"column:in_battle;default:false" json:"in_battle"`
	Power        int       `gorm:"column:power;default:0" json:"power"`
}

func (UserStatus) TableName() string { return "user_status" }

type TroopSpawn struct {
	TroopID           string  `gorm:"column:troop_id"`
	TroopLevel        int     `gorm:"column:troop_level"`
	SpawnedByAttacker bool    `gorm:"column:spawned_by_attacker"`
	SpawnedAt_X       int     `gorm:"column:spawned_at_x"`
	SpawnedAt_Y       int     `gorm:"column:spawned_at_y"`
	SpawnTime         float64 `gorm:"column:spawn_time"`
}

type InitialBattleBuilding struct {
	BuildingID string `gorm:"column:building_id"` // Maps to UUID in building_configs_base
	Grid_X     int    `gorm:"column:grid_x"`
	Grid_Y     int    `gorm:"column:grid_y"`
	Level      int    `gorm:"column:level"`
	IsBroken   bool   `gorm:"column:is_broken"`
}

type TroopSpawnArray []TroopSpawn
type InitialBuildingArray []InitialBattleBuilding

type BattleRecord struct {
	BattleID         string               `gorm:"column:battle_id;primaryKey;type:uuid"`
	TroopSpawns      TroopSpawnArray      `gorm:"column:troop_spawns;type:troop_spawn[]"`
	InitialBuildings InitialBuildingArray `gorm:"column:initial_buildings;type:initial_battle_building[]"`
}

func (BattleRecord) TableName() string {
	return "battle_record"
}
func (a TroopSpawnArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	var elements []string
	for _, ts := range a {
		row := fmt.Sprintf(`("%s",%d,%t,%d,%d,%f)`, ts.TroopID, ts.TroopLevel, ts.SpawnedByAttacker, ts.SpawnedAt_X, ts.SpawnedAt_Y, ts.SpawnTime)
		elements = append(elements, row)
	}
	return "{" + strings.Join(elements, ",") + "}", nil
}

func (a InitialBuildingArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	var elements []string
	for _, b := range a {
		row := fmt.Sprintf(`("%s",%d,%d,%d,%t)`, b.BuildingID, b.Grid_X, b.Grid_Y, b.Level, b.IsBroken)
		elements = append(elements, row)
	}
	return "{" + strings.Join(elements, ",") + "}", nil
}

func parsePostgresArray(src string) []string {
	src = strings.Trim(src, "{}")
	if src == "" {
		return nil
	}
	var results []string
	var inTuple bool
	var current strings.Builder
	for i := 0; i < len(src); i++ {
		ch := src[i]
		if ch == '(' {
			inTuple = true
			continue
		}
		if ch == ')' {
			inTuple = false
			results = append(results, current.String())
			current.Reset()
			continue
		}
		if inTuple {
			current.WriteByte(ch)
		}
	}
	return results
}
func (a *TroopSpawnArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	str, ok := value.(string)
	if !ok {
		bytes, ok := value.([]byte)
		if !ok {
			return errors.New("invalid data type for TroopSpawnArray")
		}
		str = string(bytes)
	}
	tuples := parsePostgresArray(str)
	var res []TroopSpawn
	for _, t := range tuples {
		parts := strings.Split(t, ",")
		if len(parts) < 6 {
			continue
		}
		id := strings.Trim(parts[0], "\"")
		level, _ := strconv.Atoi(parts[1])
		attacker := parts[2] == "t" || parts[2] == "true"
		x, _ := strconv.Atoi(parts[3])
		y, _ := strconv.Atoi(parts[4])
		time, _ := strconv.ParseFloat(parts[5], 64)
		res = append(res, TroopSpawn{
			TroopID:           id,
			TroopLevel:        level,
			SpawnedByAttacker: attacker,
			SpawnedAt_X:       x,
			SpawnedAt_Y:       y,
			SpawnTime:         time,
		})
	}
	*a = res
	return nil
}

func (a *InitialBuildingArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	str, ok := value.(string)
	if !ok {
		bytes, ok := value.([]byte)
		if !ok {
			return errors.New("invalid data type for InitialBuildingArray")
		}
		str = string(bytes)
	}
	tuples := parsePostgresArray(str)
	var res []InitialBattleBuilding
	for _, t := range tuples {
		parts := strings.Split(t, ",")
		if len(parts) < 5 {
			continue
		}
		id := strings.Trim(parts[0], "\"")
		x, _ := strconv.Atoi(parts[1])
		y, _ := strconv.Atoi(parts[2])
		level, _ := strconv.Atoi(parts[3])
		broken := parts[4] == "t" || parts[4] == "true"
		res = append(res, InitialBattleBuilding{
			BuildingID: id,
			Grid_X:     x,
			Grid_Y:     y,
			Level:      level,
			IsBroken:   broken,
		})
	}
	*a = res
	return nil
}
