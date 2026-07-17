package battle

import (
	"bytes"
	"errors"
	"log"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"Village_combat/models"
)

// ---------------------------------------------------------------------------
// Test helpers
//
// BattleState's slice fields use anonymous struct element types. Go treats
// two independently-written anonymous struct types as identical as long as
// their field names, types, tags, and order match exactly — so these helpers
// copy those definitions verbatim from battle.go to stay assignable.
// ---------------------------------------------------------------------------

func newEmptyState() *BattleState {
	return &BattleState{
		mu:                 sync.Mutex{},
		TroopSpawns:        []models.TroopSpawn{},
		AliveTroopAttacker: nil,
		AliveTroopDefender: nil,
		Buildings:          nil,
		AliveBuildings:     nil,
		DeathMap:           []int{},
		StartTime:          time.Now(),
	}
}

// addBuilding registers a building (optionally a defense building) and adds
// it to AliveBuildings with the given starting health. Returns its index
// within state.Buildings (== the BuildingIndex recorded in AliveBuildings).
func addBuilding(
	state *BattleState,
	buildingID string,
	gridX, gridY, level int,
	isDefense bool,
	defStat models.DefenseBuildingStats,
	defLevelStat models.DefenseBuildingLevelStats,
	aliveHealth int64,
) int {
	var defender *struct {
		LevelStat models.DefenseBuildingLevelStats
		Stat      models.DefenseBuildingStats
	}
	if isDefense {
		defender = &struct {
			LevelStat models.DefenseBuildingLevelStats
			Stat      models.DefenseBuildingStats
		}{LevelStat: defLevelStat, Stat: defStat}
	}

	state.Buildings = append(state.Buildings, struct {
		Placed_Building models.PlacedBuilding
		Defender        *struct {
			LevelStat models.DefenseBuildingLevelStats
			Stat      models.DefenseBuildingStats
		}
	}{
		Placed_Building: models.PlacedBuilding{
			ID:         buildingID + "-placed",
			BuildingID: buildingID,
			GridX:      gridX,
			GridY:      gridY,
			Level:      level,
		},
		Defender: defender,
	})

	buildingIndex := len(state.Buildings) - 1
	state.AliveBuildings = append(state.AliveBuildings, struct {
		BuildingIndex   int   `json:"BuildingIndex"`
		HealthRemaining int64 `json:"HealthRemaining"`
	}{BuildingIndex: buildingIndex, HealthRemaining: aliveHealth})

	return buildingIndex
}

func addAttackerTroop(
	state *BattleState,
	troopID string, level int,
	x, y float64,
	config models.TroopConfig,
	levelStat models.TroopLevelStats,
	health int64,
) int {
	spawnIndex := len(state.TroopSpawns)
	state.TroopSpawns = append(state.TroopSpawns, models.TroopSpawn{
		TroopID:           troopID,
		TroopLevel:        level,
		SpawnedByAttacker: true,
		SpawnedAt_X:       int(x),
		SpawnedAt_Y:       int(y),
	})
	state.AliveTroopAttacker = append(state.AliveTroopAttacker, struct {
		TroopIndex      int
		Current_X       float64
		Current_Y       float64
		Config          models.TroopConfig
		LevelStat       models.TroopLevelStats
		HealthRemaining int64
	}{
		TroopIndex:      spawnIndex,
		Current_X:       x,
		Current_Y:       y,
		Config:          config,
		LevelStat:       levelStat,
		HealthRemaining: health,
	})
	return len(state.AliveTroopAttacker) - 1
}

func addDefenderTroop(
	state *BattleState,
	troopID string, level int,
	x, y float64,
	config models.TroopConfig,
	levelStat models.TroopLevelStats,
	health int64,
) int {
	spawnIndex := len(state.TroopSpawns)
	state.TroopSpawns = append(state.TroopSpawns, models.TroopSpawn{
		TroopID:           troopID,
		TroopLevel:        level,
		SpawnedByAttacker: false,
		SpawnedAt_X:       int(x),
		SpawnedAt_Y:       int(y),
	})
	state.AliveTroopDefender = append(state.AliveTroopDefender, struct {
		TroopIndex      int
		Current_X       float64
		Current_Y       float64
		Config          models.TroopConfig
		LevelStat       models.TroopLevelStats
		HealthRemaining int64
	}{
		TroopIndex:      spawnIndex,
		Current_X:       x,
		Current_Y:       y,
		Config:          config,
		LevelStat:       levelStat,
		HealthRemaining: health,
	})
	return len(state.AliveTroopDefender) - 1
}

// ---------------------------------------------------------------------------
// simulate(): targeting, movement, damage
// ---------------------------------------------------------------------------

func TestSimulate_AttackerDamagesBuildingInRange(t *testing.T) {
	models.BuildingSize = map[string]struct{ X, Y int }{"res-1": {X: 1, Y: 1}}
	models.BuildingID_Category = map[string]models.BuildingCategory{"res-1": models.Resource}

	state := newEmptyState()
	addBuilding(state, "res-1", 0, 0, 5, false, models.DefenseBuildingStats{}, models.DefenseBuildingLevelStats{}, 1000)
	addAttackerTroop(state, "barb", 1, 0, 0,
		models.TroopConfig{AttackRange: 1, AttackSpeedSeconds: 1, MovementSpeed: 1},
		models.TroopLevelStats{DamagePerShot: 50, Health: 100}, 100)

	update := simulate(state)

	if state.AliveBuildings[0].HealthRemaining != 950 {
		t.Errorf("expected building health 950, got %d", state.AliveBuildings[0].HealthRemaining)
	}
	dmg, ok := update["building_damage"].([]int64)
	if !ok || len(dmg) != 1 || dmg[0] != 50 {
		t.Errorf("unexpected building_damage: %#v", update["building_damage"])
	}
}

func TestSimulate_AttackerMovesTowardOutOfRangeBuilding(t *testing.T) {
	models.BuildingSize = map[string]struct{ X, Y int }{"res-1": {X: 1, Y: 1}}
	models.BuildingID_Category = map[string]models.BuildingCategory{"res-1": models.Resource}

	state := newEmptyState()
	addBuilding(state, "res-1", 10, 10, 5, false, models.DefenseBuildingStats{}, models.DefenseBuildingLevelStats{}, 1000)
	addAttackerTroop(state, "barb", 1, 0, 0,
		models.TroopConfig{AttackRange: 1, AttackSpeedSeconds: 1, MovementSpeed: 2},
		models.TroopLevelStats{DamagePerShot: 50, Health: 100}, 100)

	simulate(state)

	dist := math.Sqrt(10*10 + 10*10)
	ratio := 2.0 / dist
	expectedX := 10 * ratio
	expectedY := 10 * ratio

	gotX := state.AliveTroopAttacker[0].Current_X
	gotY := state.AliveTroopAttacker[0].Current_Y
	const eps = 1e-9
	if math.Abs(gotX-expectedX) > eps || math.Abs(gotY-expectedY) > eps {
		t.Errorf("expected position (%v,%v), got (%v,%v)", expectedX, expectedY, gotX, gotY)
	}
	// Building should be untouched — the troop moved instead of attacking.
	if state.AliveBuildings[0].HealthRemaining != 1000 {
		t.Errorf("expected building untouched while troop is out of range, got %d", state.AliveBuildings[0].HealthRemaining)
	}
}

func TestSimulate_PreferredTargetPrioritizesCategoryOverCloserBuilding(t *testing.T) {
	defenseCat := models.Defense
	models.BuildingSize = map[string]struct{ X, Y int }{
		"resource-1": {X: 1, Y: 1},
		"defense-1":  {X: 1, Y: 1},
	}
	models.BuildingID_Category = map[string]models.BuildingCategory{
		"resource-1": models.Resource,
		"defense-1":  defenseCat,
	}

	state := newEmptyState()
	addBuilding(state, "resource-1", 1, 0, 5, false, models.DefenseBuildingStats{}, models.DefenseBuildingLevelStats{}, 1000) // very close
	addBuilding(state, "defense-1", 10, 0, 5, false, models.DefenseBuildingStats{}, models.DefenseBuildingLevelStats{}, 1000) // far, but preferred

	cfg := models.TroopConfig{
		AttackRange:        1,
		AttackSpeedSeconds: 1,
		MovementSpeed:      1,
		PreferredTarget:    &defenseCat,
	}
	addAttackerTroop(state, "giant", 1, 0, 0, cfg, models.TroopLevelStats{DamagePerShot: 10, Health: 500}, 500)

	simulate(state)

	if state.AliveBuildings[0].HealthRemaining != 1000 {
		t.Errorf("expected the closer, non-preferred resource building to be untouched, health=%d", state.AliveBuildings[0].HealthRemaining)
	}
	if state.AliveTroopAttacker[0].Current_X <= 0 {
		t.Errorf("expected troop to move toward the preferred (further) defense building, X=%v", state.AliveTroopAttacker[0].Current_X)
	}
}

func TestSimulate_DefenseBuildingDamagesClosestAttacker(t *testing.T) {
	models.BuildingSize = map[string]struct{ X, Y int }{"cannon": {X: 2, Y: 2}}
	models.BuildingID_Category = map[string]models.BuildingCategory{"cannon": models.Defense}

	state := newEmptyState()
	defStat := models.DefenseBuildingStats{AttackSpeedSeconds: 1, AttackRange: 10, DamageType: models.SingleTarget}
	defLevelStat := models.DefenseBuildingLevelStats{DamagePerShot: 40}
	addBuilding(state, "cannon", 0, 0, 5, true, defStat, defLevelStat, 1000)

	// Center of a 2x2 building at (0,0) is (1,1) — place the troop right there.
	addAttackerTroop(state, "barb", 1, 1, 1,
		models.TroopConfig{AttackRange: 0.01, AttackSpeedSeconds: 1, MovementSpeed: 1},
		models.TroopLevelStats{DamagePerShot: 5, Health: 100}, 100)

	simulate(state)

	if state.AliveTroopAttacker[0].HealthRemaining != 60 {
		t.Errorf("expected attacker health 60 after cannon hit (100-40), got %d", state.AliveTroopAttacker[0].HealthRemaining)
	}
}

func TestSimulate_SplashDamageHitsNearbyAttackersOnly(t *testing.T) {
	models.BuildingSize = map[string]struct{ X, Y int }{"wizard-tower": {X: 2, Y: 2}}
	models.BuildingID_Category = map[string]models.BuildingCategory{"wizard-tower": models.Defense}

	state := newEmptyState()
	defStat := models.DefenseBuildingStats{AttackSpeedSeconds: 1, AttackRange: 10, DamageType: models.Splash}
	defLevelStat := models.DefenseBuildingLevelStats{DamagePerShot: 30}
	addBuilding(state, "wizard-tower", 0, 0, 5, true, defStat, defLevelStat, 1000)

	cfg := models.TroopConfig{AttackRange: 0.01, AttackSpeedSeconds: 1, MovementSpeed: 1}
	lvl := models.TroopLevelStats{DamagePerShot: 5, Health: 100}

	// Center of the building is (1,1).
	addAttackerTroop(state, "barb", 1, 1, 1, cfg, lvl, 100) // idx0 — primary target (distance 0 from center)
	addAttackerTroop(state, "barb", 1, 2, 1, cfg, lvl, 100) // idx1 — distance 1 from idx0, within 1.5 splash radius
	addAttackerTroop(state, "barb", 1, 5, 5, cfg, lvl, 100) // idx2 — far from idx0, outside splash radius

	simulate(state)

	if state.AliveTroopAttacker[0].HealthRemaining != 70 {
		t.Errorf("expected primary target health 70 (100-30), got %d", state.AliveTroopAttacker[0].HealthRemaining)
	}
	if state.AliveTroopAttacker[1].HealthRemaining != 70 {
		t.Errorf("expected splash-hit neighbor health 70 (100-30), got %d", state.AliveTroopAttacker[1].HealthRemaining)
	}
	if state.AliveTroopAttacker[2].HealthRemaining != 100 {
		t.Errorf("expected distant troop untouched by splash, got %d", state.AliveTroopAttacker[2].HealthRemaining)
	}
}

func TestSimulate_DefenderTroopAttacksClosestAttacker(t *testing.T) {
	models.BuildingSize = map[string]struct{ X, Y int }{}
	models.BuildingID_Category = map[string]models.BuildingCategory{}

	state := newEmptyState()
	defCfg := models.TroopConfig{AttackRange: 5, AttackSpeedSeconds: 1, MovementSpeed: 1}
	defLvl := models.TroopLevelStats{DamagePerShot: 20, Health: 100}
	addDefenderTroop(state, "def-troop", 1, 0, 0, defCfg, defLvl, 100)

	// Both attackers have an attack range too small to act back this tick,
	// so we isolate the assertion to the defender troop's own targeting.
	atkCfg := models.TroopConfig{AttackRange: 0.01, AttackSpeedSeconds: 1, MovementSpeed: 1}
	addAttackerTroop(state, "atk-close", 1, 1, 0, atkCfg, models.TroopLevelStats{DamagePerShot: 1, Health: 100}, 100) // idx0, distance 1
	addAttackerTroop(state, "atk-far", 1, 4, 0, atkCfg, models.TroopLevelStats{DamagePerShot: 1, Health: 100}, 100)   // idx1, distance 4

	simulate(state)

	if state.AliveTroopAttacker[0].HealthRemaining != 80 {
		t.Errorf("expected closer attacker to take 20 damage (100-20), got %d", state.AliveTroopAttacker[0].HealthRemaining)
	}
	if state.AliveTroopAttacker[1].HealthRemaining != 100 {
		t.Errorf("expected farther attacker untouched, got %d", state.AliveTroopAttacker[1].HealthRemaining)
	}
}

func TestSimulate_MinimumDamageGuaranteeAppliesWhenComputedDamageTruncatesToZero(t *testing.T) {
	models.BuildingSize = map[string]struct{ X, Y int }{"res-1": {X: 1, Y: 1}}
	models.BuildingID_Category = map[string]models.BuildingCategory{"res-1": models.Resource}

	state := newEmptyState()
	addBuilding(state, "res-1", 0, 0, 5, false, models.DefenseBuildingStats{}, models.DefenseBuildingLevelStats{}, 1000)

	// DamagePerShot / AttackSpeedSeconds = 1/10 = 0.1 -> truncates to 0, but
	// DamagePerShot > 0 so the minimum-damage floor of 1 should apply.
	cfg := models.TroopConfig{AttackRange: 1, AttackSpeedSeconds: 10, MovementSpeed: 1}
	addAttackerTroop(state, "barb", 1, 0, 0, cfg, models.TroopLevelStats{DamagePerShot: 1, Health: 100}, 100)

	simulate(state)

	if state.AliveBuildings[0].HealthRemaining != 999 {
		t.Errorf("expected minimum-damage floor of 1 to apply (1000-1), got health %d", state.AliveBuildings[0].HealthRemaining)
	}
}

// ---------------------------------------------------------------------------
// simulate(): death / removal bookkeeping
// ---------------------------------------------------------------------------

func TestSimulate_DeadBuildingIsRemovedAndRecorded(t *testing.T) {
	models.BuildingSize = map[string]struct{ X, Y int }{"res-1": {X: 1, Y: 1}}
	models.BuildingID_Category = map[string]models.BuildingCategory{"res-1": models.Resource}

	state := newEmptyState()
	idx := addBuilding(state, "res-1", 0, 0, 5, false, models.DefenseBuildingStats{}, models.DefenseBuildingLevelStats{}, 0) // already at 0 health

	update := simulate(state)

	if len(state.AliveBuildings) != 0 {
		t.Fatalf("expected dead building to be removed from AliveBuildings, still has %d entries", len(state.AliveBuildings))
	}
	if len(state.DeathMap) != 1 || state.DeathMap[0] != idx {
		t.Errorf("expected DeathMap to record building index %d, got %v", idx, state.DeathMap)
	}
	died, ok := update["building_died"].([]int)
	if !ok || len(died) != 1 || died[0] != 0 {
		t.Errorf("unexpected building_died: %#v", update["building_died"])
	}
}

func TestSimulate_DeadAttackerTroopIsRemoved(t *testing.T) {
	models.BuildingSize = map[string]struct{ X, Y int }{}
	models.BuildingID_Category = map[string]models.BuildingCategory{}

	state := newEmptyState()
	addAttackerTroop(state, "barb", 1, 0, 0,
		models.TroopConfig{AttackRange: 1, AttackSpeedSeconds: 1, MovementSpeed: 1},
		models.TroopLevelStats{DamagePerShot: 1, Health: 100}, 0) // already dead

	update := simulate(state)

	if len(state.AliveTroopAttacker) != 0 {
		t.Fatalf("expected dead attacker troop to be removed, still has %d entries", len(state.AliveTroopAttacker))
	}
	died, ok := update["attacker_troop_died"].([]int)
	if !ok || len(died) != 1 || died[0] != 0 {
		t.Errorf("unexpected attacker_troop_died: %#v", update["attacker_troop_died"])
	}
}

func TestSimulate_DeadDefenderTroopIsRemoved(t *testing.T) {
	models.BuildingSize = map[string]struct{ X, Y int }{}
	models.BuildingID_Category = map[string]models.BuildingCategory{}

	state := newEmptyState()
	addDefenderTroop(state, "def-troop", 1, 0, 0,
		models.TroopConfig{AttackRange: 1, AttackSpeedSeconds: 1, MovementSpeed: 1},
		models.TroopLevelStats{DamagePerShot: 1, Health: 100}, 0) // already dead

	update := simulate(state)

	if len(state.AliveTroopDefender) != 0 {
		t.Fatalf("expected dead defender troop to be removed, still has %d entries", len(state.AliveTroopDefender))
	}
	died, ok := update["defender_troop_died"].([]int)
	if !ok || len(died) != 1 || died[0] != 0 {
		t.Errorf("unexpected defender_troop_died: %#v", update["defender_troop_died"])
	}
}

// ---------------------------------------------------------------------------
// logErr
// ---------------------------------------------------------------------------

func TestLogErr_NoopWhenNil(t *testing.T) {
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	logErr("some-context", nil)

	if buf.Len() != 0 {
		t.Errorf("expected no log output for a nil error, got: %q", buf.String())
	}
}

func TestLogErr_LogsWhenErrorPresent(t *testing.T) {
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	logErr("my-context", errors.New("boom"))

	out := buf.String()
	if !strings.Contains(out, "my-context") || !strings.Contains(out, "boom") {
		t.Errorf("expected log output to mention context and error, got: %q", out)
	}
}
