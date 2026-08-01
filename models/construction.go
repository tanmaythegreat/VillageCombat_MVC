package models

import (
	"encoding/json"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetConstructionTasks(userID string) (json.RawMessage, error) {
	var constructionTasks []ConstructionTask
	err := DB.Where("user_id = ?", userID).Find(&constructionTasks).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(constructionTasks)
}
func ConstructBuilding(userID string, buildingID string, x int, y int, tx *gorm.DB, duration int) (PlacedBuilding, ConstructionTask, error) {
	newBuilding := PlacedBuilding{
		UserID:     userID,
		BuildingID: buildingID,
		GridX:      x,
		GridY:      y,
		Level:      0,
	}
	err := tx.Select("user_id", "building_id", "grid_x", "grid_y", "level").Clauses(clause.Returning{}).Create(&newBuilding).Error
	if err != nil {
		return newBuilding, ConstructionTask{}, err
	}
	task, err := StartConstruction_Building(userID, BuildingConstruction, newBuilding.ID, duration, tx)
	return newBuilding, task, err
}
func UpgradeBuilding(userID string, placed_buildingID string, tx *gorm.DB, duration int) (ConstructionTask, error) {
	task, err := StartConstruction_Building(userID, BuildingUpgrade, placed_buildingID, duration, tx)
	return task, err
}
func StartConstruction_Building(userID string, taskType ConstructionType, placedBuildingId string, duration_seconds int, tx *gorm.DB) (ConstructionTask, error) {

	newTask := ConstructionTask{
		UserID:           userID,
		TaskType:         taskType,
		PlacedBuildingID: placedBuildingId,
		DurationSeconds:  duration_seconds,
	}

	err := tx.Select(
		"user_id",
		"task_type",
		"placed_building_id",
		"duration_seconds",
	).Clauses(clause.Returning{}).Create(&newTask).Error
	return newTask, err
}
func GetConstructionCost(building_id string, upgrade_to int) (UpgradeCost, error) {
	var cost UpgradeCost
	err := DB.Where("building_id = ? AND upgrade_to_level = ?", building_id, upgrade_to).
		First(&cost).Error
	return cost, err
}
func CheckIsConstructionComplete(userId string) ([]ConstructionTask, []PlacedBuilding, error) {
	var completedTasks []ConstructionTask
	var updatedBuildings []PlacedBuilding

	tx := DB.Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	defer tx.Rollback()

	err := tx.Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Where("started_at + (duration_seconds * INTERVAL '1 second') <= NOW()").
		Delete(&completedTasks).Error

	if err != nil {
		log.Printf("Error deleting completed tasks for user %s: %v\n", userId, err)
		return nil, nil, err
	}

	if len(completedTasks) == 0 {
		tx.Commit()
		return completedTasks, updatedBuildings, nil
	}

	var buildingIDs []string
	for _, task := range completedTasks {
		if task.TaskType == TroopTraining {
			err := AddTroopsToUser(userId, *task.TroopID, *task.TroopLevelTo, *task.TroopCount, tx)
			if err != nil {
				tx.Rollback()
				return nil, nil, err
			}
		} else if task.TaskType == BuildingRepair {
			err = SetBrokenBuilding(userId, task.PlacedBuildingID, false, tx)
			if err != nil {
				tx.Rollback()
				return nil, nil, err
			}
		} else {
			buildingIDs = append(buildingIDs, task.PlacedBuildingID)
		}
	}
	if len(buildingIDs) > 0 {
		err = tx.Clauses(clause.Returning{}).
			Model(&updatedBuildings).
			Where("id IN ?", buildingIDs).
			Update("level", gorm.Expr("level + 1")).Error

		if err != nil {
			log.Printf("Error updating building levels for user %s: %v\n", userId, err)
			return nil, nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction for user %s: %v\n", userId, err)
		return nil, nil, err
	}
	return completedTasks, updatedBuildings, nil
}
func IncrementUserTownHallLevel(userId string) error {
	return DB.Table("user_data").
		Where("user_id = ?", userId).
		Update("town_hall_level", gorm.Expr("town_hall_level + 1")).
		Error
}
func IsConstructionUnderProgress(userId string, placed_building_id string) (bool, error) {
	n := int64(0)
	err := DB.Table(ConstructionTask{}.TableName()).Where("user_id = ? AND placed_building_id = ?", userId, placed_building_id).Count(&n).Error
	return n > 0, err
}
