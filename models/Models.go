package models

import (
	"database/sql/driver"
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
	BuildingRepair       ConstructionType = "building_repair"
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
	Health        int64  `gorm:"column:health" json:"health"`
	DamagePerShot int    `gorm:"column:damage_per_shot" json:"damage_per_shot"`
}

func (TroopLevelStats) TableName() string { return "troop_level_stats" }

type UpgradeCost struct {
	ID                    string  `gorm:"column:id;primaryKey" json:"id"`
	TroopID               *string `gorm:"column:troop_id" json:"troop_id,omitempty"`
	BuildingID            *string `gorm:"column:building_id" json:"building_id,omitempty"`
	UpgradeToLevel        int     `gorm:"column:upgrade_to_level" json:"upgrade_to_level"`
	GoldRequired          int64   `gorm:"column:gold_required" json:"gold_required"`
	ElixirRequired        int64   `gorm:"column:elixir_required" json:"elixir_required"`
	DarkElixirRequired    int64   `gorm:"column:dark_elixir_required" json:"dark_elixir_required"`
	OrGemRequired         int64   `gorm:"column:or_gem_required" json:"or_gem_required"`
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
	Health     int64  `gorm:"column:health" json:"health"`
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
	StorageCapacity       int64   `gorm:"column:storage_capacity" json:"storage_capacity"`
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
	CurrentGold             int64     `gorm:"column:current_gold" json:"current_gold"`
	CurrentElixir           int64     `gorm:"column:current_elixir" json:"current_elixir"`
	CurrentDarkElixir       int64     `gorm:"column:current_dark_elixir" json:"current_dark_elixir"`
	TotalGoldCapacity       int64     `gorm:"column:total_gold_capacity" json:"total_gold_capacity"`
	TotalElixirCapacity     int64     `gorm:"column:total_elixir_capacity" json:"total_elixir_capacity"`
	TotalDarkElixirCapacity int64     `gorm:"column:total_dark_elixir_capacity" json:"total_dark_elixir_capacity"`
	TotalTroopCapacity      int       `gorm:"column:total_troop_capacity" json:"total_troop_capacity"`
	CurrentGems             int64     `gorm:"column:current_gems" json:"current_gems"`
	UpdatedAt               time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserData) TableName() string { return "user_data" }

type BattleHistory struct {
	BattleID         string            `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:battle_id" json:"battle_id"`
	AttackerName     string            `gorm:"type:uuid;not null;column:attacker_name" json:"attacker_name"`
	DefenderName     string            `gorm:"type:uuid;not null;column:defender_name" json:"defender_name"`
	ElixirLooted     int64             `gorm:"default:0;column:elixir_looted" json:"elixir_looted"`
	GoldLooted       int64             `gorm:"default:0;column:gold_looted" json:"gold_looted"`
	DarkElixirLooted int64             `gorm:"default:0;column:dark_elixir_looted" json:"dark_elixir_looted"`
	FoughtAt         time.Time         `gorm:"default:CURRENT_TIMESTAMP;column:fought_at" json:"fought_at"`
	BattleDuration   int               `gorm:"column:battle_duration" json:"battle_duration"`
	DoDefenderKnow   bool              `gorm:"column:do_defender_know" json:"do_defender_know"`
	WinnerAttacker   bool              `gorm:"column:winner_attacker" json:"winner_attacker"`
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
	AttackPower  int       `gorm:"column:attack_power;default:0" json:"attack_power"`
	DefencePower int       `gorm:"column:defence_power;default:0" json:"defence_power"`
}

func (UserStatus) TableName() string { return "user_status" }

type TroopSpawn struct {
	TroopID           string  `gorm:"column:troop_id" json:"troop_id"`
	TroopLevel        int     `gorm:"column:troop_level" json:"troop_level"`
	SpawnedByAttacker bool    `gorm:"column:spawned_by_attacker" json:"spawned_by_attacker"`
	SpawnedAt_X       int     `gorm:"column:spawned_at_x" json:"spawnedAt_X"`
	SpawnedAt_Y       int     `gorm:"column:spawned_at_y" json:"spawnedAt_Y"`
	SpawnTime         float64 `gorm:"column:spawn_time" json:"spawn_time"`
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

// ---------------------------------------------------------------------------
// Shared helpers for Postgres composite-type / array-of-composite encoding
// ---------------------------------------------------------------------------

// pgQuote wraps s in double quotes and backslash-escapes any embedded
// double quotes or backslashes, per Postgres text-format rules. It is used
// both for quoting individual composite fields and for quoting a whole
// composite literal as an element of an array.
func pgQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' || c == '\\' {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	b.WriteByte('"')
	return b.String()
}

func pgBool(v bool) string {
	if v {
		return "t"
	}
	return "f"
}

func toScanString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("unsupported scan type %T", value)
	}
}

// parsePgArrayElements splits a Postgres array literal like
// {"(...)","(...)"}  or  {NULL,"(...)"}
// into its top-level elements, honoring quoting/escaping so that commas
// and parentheses *inside* quoted elements are not treated as separators.
// A nil entry in the returned slice represents a SQL NULL array element.
func parsePgArrayElements(src string) ([]*string, error) {
	src = strings.TrimSpace(src)
	if len(src) < 2 || src[0] != '{' || src[len(src)-1] != '}' {
		return nil, fmt.Errorf("invalid array literal: %q", src)
	}
	inner := src[1 : len(src)-1]
	n := len(inner)
	if n == 0 {
		return nil, nil
	}

	var elements []*string
	i := 0
	for i < n {
		for i < n && inner[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}

		if inner[i] == '"' {
			var sb strings.Builder
			i++ // skip opening quote
			for i < n {
				c := inner[i]
				if c == '\\' && i+1 < n {
					sb.WriteByte(inner[i+1])
					i += 2
					continue
				}
				if c == '"' {
					i++
					break
				}
				sb.WriteByte(c)
				i++
			}
			s := sb.String()
			elements = append(elements, &s)
		} else {
			start := i
			for i < n && inner[i] != ',' {
				i++
			}
			val := inner[start:i]
			if val == "NULL" {
				elements = append(elements, nil)
			} else {
				v := val
				elements = append(elements, &v)
			}
		}

		// advance to next element
		for i < n && inner[i] != ',' {
			i++
		}
		if i < n && inner[i] == ',' {
			i++
		}
	}
	return elements, nil
}

// parseCompositeFields parses a single composite-type literal, e.g.
// ("some,id",1,t,2,3,4.5)
// into its raw field strings, honoring quoting/escaping of individual
// fields. An empty (unquoted, zero-length) field represents SQL NULL for
// that column.
func parseCompositeFields(tuple string) ([]string, error) {
	tuple = strings.TrimSpace(tuple)
	if len(tuple) < 2 || tuple[0] != '(' || tuple[len(tuple)-1] != ')' {
		return nil, fmt.Errorf("invalid composite literal: %q", tuple)
	}
	inner := tuple[1 : len(tuple)-1]
	n := len(inner)

	var fields []string
	i := 0
	for {
		if i >= n {
			fields = append(fields, "")
			break
		}
		if inner[i] == '"' {
			var sb strings.Builder
			i++
			for i < n {
				c := inner[i]
				if c == '\\' && i+1 < n {
					sb.WriteByte(inner[i+1])
					i += 2
					continue
				}
				if c == '"' {
					i++
					break
				}
				sb.WriteByte(c)
				i++
			}
			fields = append(fields, sb.String())
		} else {
			start := i
			for i < n && inner[i] != ',' {
				i++
			}
			fields = append(fields, inner[start:i])
		}

		if i < n && inner[i] == ',' {
			i++
			continue
		}
		break
	}
	return fields, nil
}
func (a TroopSpawnArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	elements := make([]string, 0, len(a))
	for _, ts := range a {
		tuple := fmt.Sprintf("(%s,%d,%s,%d,%d,%s)",
			pgQuote(ts.TroopID),
			ts.TroopLevel,
			pgBool(ts.SpawnedByAttacker),
			ts.SpawnedAt_X,
			ts.SpawnedAt_Y,
			strconv.FormatFloat(ts.SpawnTime, 'f', -1, 64),
		)
		elements = append(elements, pgQuote(tuple))
	}
	return "{" + strings.Join(elements, ",") + "}", nil
}

func (a *TroopSpawnArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	str, err := toScanString(value)
	if err != nil {
		return fmt.Errorf("TroopSpawnArray.Scan: %w", err)
	}

	elements, err := parsePgArrayElements(str)
	if err != nil {
		return fmt.Errorf("TroopSpawnArray.Scan: %w", err)
	}

	res := make([]TroopSpawn, 0, len(elements))
	for idx, el := range elements {
		if el == nil {
			continue // NULL element in the array ,nothing to populate
		}
		fields, err := parseCompositeFields(*el)
		if err != nil {
			return fmt.Errorf("TroopSpawnArray.Scan: element %d: %w", idx, err)
		}
		if len(fields) != 6 {
			return fmt.Errorf("TroopSpawnArray.Scan: element %d: expected 6 fields, got %d (%q)", idx, len(fields), *el)
		}

		level, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("TroopSpawnArray.Scan: element %d: invalid troop_level %q: %w", idx, fields[1], err)
		}
		x, err := strconv.Atoi(fields[3])
		if err != nil {
			return fmt.Errorf("TroopSpawnArray.Scan: element %d: invalid spawned_at_x %q: %w", idx, fields[3], err)
		}
		y, err := strconv.Atoi(fields[4])
		if err != nil {
			return fmt.Errorf("TroopSpawnArray.Scan: element %d: invalid spawned_at_y %q: %w", idx, fields[4], err)
		}
		spawnTime, err := strconv.ParseFloat(fields[5], 64)
		if err != nil {
			return fmt.Errorf("TroopSpawnArray.Scan: element %d: invalid spawn_time %q: %w", idx, fields[5], err)
		}

		res = append(res, TroopSpawn{
			TroopID:           fields[0],
			TroopLevel:        level,
			SpawnedByAttacker: fields[2] == "t" || fields[2] == "true",
			SpawnedAt_X:       x,
			SpawnedAt_Y:       y,
			SpawnTime:         spawnTime,
		})
	}
	*a = res
	return nil
}

func (a InitialBuildingArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	elements := make([]string, 0, len(a))
	for _, b := range a {
		tuple := fmt.Sprintf("(%s,%d,%d,%d,%s)",
			pgQuote(b.BuildingID),
			b.Grid_X,
			b.Grid_Y,
			b.Level,
			pgBool(b.IsBroken),
		)
		elements = append(elements, pgQuote(tuple))
	}
	return "{" + strings.Join(elements, ",") + "}", nil
}

func (a *InitialBuildingArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	str, err := toScanString(value)
	if err != nil {
		return fmt.Errorf("InitialBuildingArray.Scan: %w", err)
	}

	elements, err := parsePgArrayElements(str)
	if err != nil {
		return fmt.Errorf("InitialBuildingArray.Scan: %w", err)
	}

	res := make([]InitialBattleBuilding, 0, len(elements))
	for idx, el := range elements {
		if el == nil {
			continue
		}
		fields, err := parseCompositeFields(*el)
		if err != nil {
			return fmt.Errorf("InitialBuildingArray.Scan: element %d: %w", idx, err)
		}
		if len(fields) != 5 {
			return fmt.Errorf("InitialBuildingArray.Scan: element %d: expected 5 fields, got %d (%q)", idx, len(fields), *el)
		}

		x, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("InitialBuildingArray.Scan: element %d: invalid grid_x %q: %w", idx, fields[1], err)
		}
		y, err := strconv.Atoi(fields[2])
		if err != nil {
			return fmt.Errorf("InitialBuildingArray.Scan: element %d: invalid grid_y %q: %w", idx, fields[2], err)
		}
		level, err := strconv.Atoi(fields[3])
		if err != nil {
			return fmt.Errorf("InitialBuildingArray.Scan: element %d: invalid level %q: %w", idx, fields[3], err)
		}

		res = append(res, InitialBattleBuilding{
			BuildingID: fields[0],
			Grid_X:     x,
			Grid_Y:     y,
			Level:      level,
			IsBroken:   fields[4] == "t" || fields[4] == "true",
		})
	}
	*a = res
	return nil
}
