package models

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetUserTrainedTroops(userId string) (json.RawMessage, error) {
	var troops []TrainedTroop
	err := DB.Where("user_id = ?", userId).Find(&troops).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(troops)
}
func GetTroopUpgradeCost(Troop_id string, level_upgrade_to int) (UpgradeCost, error) {
	var cost UpgradeCost
	err := DB.Where("troop_id = ? AND upgrade_to_level = ?", Troop_id, level_upgrade_to).
		First(&cost).Error
	return cost, err
}
func AddTroopsToUser(userId string, troopId string, level int, count int, tx *gorm.DB) error {
	var troop TrainedTroop
	err := tx.Where("user_id = ? AND troop_id = ? AND level = ?", userId, troopId, level).First(&troop).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newTroop := TrainedTroop{
				UserID:        userId,
				TroopID:       troopId,
				Level:         level,
				Count:         count,
				LastUpdatedAt: time.Now(),
			}
			return tx.Select("user_id", "troop_id", "level", "count", "last_updated_at").
				Clauses(clause.Returning{}).
				Create(&newTroop).Error
		}
		return err
	}
	return tx.Model(&troop).Updates(map[string]interface{}{
		"count":           gorm.Expr("count + ?", count),
		"last_updated_at": time.Now(),
	}).Error
}
func SubtractTroopsOfUser(userId string, troopId string, level int, count int, tx *gorm.DB) (bool, error) {
	if count < 0 {
		return false, errors.New("cannot subtract a negative amount of troops")
	}
	if count == 0 {
		return true, nil
	}
	result := tx.Model(&TrainedTroop{}).
		Where(
			"user_id = ? AND troop_id = ? AND level = ? AND count >= ?",
			userId, troopId, level, count,
		).
		Updates(map[string]interface{}{
			"count":           gorm.Expr("count - ?", count),
			"last_updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}
func GetAllTroopsDataJSON() (json.RawMessage, error) {
	var troops []TroopConfig
	err := DB.Preload("LevelStats").Preload("UpgradeCosts").Find(&troops).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(troops)
}
func StartTrainingTroops(userID string, troopId string, count int, placedBuildingId string, duration_seconds int, level_to int, tx *gorm.DB) (ConstructionTask, error) {

	newTask := ConstructionTask{
		UserID:           userID,
		TaskType:         TroopTraining,
		PlacedBuildingID: placedBuildingId,
		TroopID:          &troopId,
		TroopCount:       &count,
		DurationSeconds:  duration_seconds,
		TroopLevelTo:     &level_to,
	}

	err := tx.Select(
		"user_id",
		"task_type",
		"placed_building_id",
		"duration_seconds",
		"troop_id",
		"troop_count",
		"troop_level_to",
	).Clauses(clause.Returning{}).Create(&newTask).Error
	return newTask, err
}
func GetTroopCapacityDifference(level1 int, level2 int) (int, error) {
	var stats []ArmyBuildingLevelStats

	result := DB.Select("level, troop_capacity").
		Where("building_id = ? AND level IN (?, ?)", Barracks_ID, level1, level2).
		Find(&stats)

	if result.Error != nil {
		return 0, result.Error
	}

	var cap1, cap2 int
	found1, found2 := false, false

	for _, s := range stats {
		if s.Level == level1 {
			cap1 = s.TroopCapacity
			found1 = true
		} else if s.Level == level2 {
			cap2 = s.TroopCapacity
			found2 = true
		}
	}

	if !found1 || !found2 {
		log.Printf(
			"Warning: Missing troop (level1 found: %v, level2 found: %v)\n",
			found1, found2,
		)
	}

	return cap2 - cap1, nil
}
func HasTroopTrainingTask(userID string) (bool, error) {
	var count int64
	err := DB.
		Model(&ConstructionTask{}).
		Where("user_id = ? AND task_type = ?", userID, TroopTraining).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
func CanAddTroops(userID string, count int) (bool, error) {
	var canAdd bool

	err := DB.Model(&UserData{}).
		Select(`
			COALESCE(SUM(trained_troops.count), 0) + ? <= user_data.total_troop_capacity AS can_add
		`, count).
		Joins(`
			LEFT JOIN trained_troops 
			ON trained_troops.user_id = user_data.user_id
		`).
		Where("user_data.user_id = ?", userID).
		Group("user_data.user_id, user_data.total_troop_capacity").
		Scan(&canAdd).Error

	if err != nil {
		return false, err
	}

	return canAdd, nil
}
