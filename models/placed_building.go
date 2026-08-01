package models

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetPlacedBuildingJSON(userID string) (json.RawMessage, error) {
	var configs []PlacedBuilding
	err := DB.Where("user_id = ?", userID).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(configs)
}
func GetPlacedBuildings(userID string) ([]PlacedBuilding, error) {
	var configs []PlacedBuilding
	err := DB.Where("user_id = ?", userID).Find(&configs).Error
	return configs, err
}
func GetPlacedBuilding(userID string, placedBuildingId string) (PlacedBuilding, error) {
	var configs PlacedBuilding
	err := DB.Where("user_id = ? AND id = ?", userID, placedBuildingId).Clauses(clause.Returning{}).First(&configs).Error
	return configs, err
}

// TODO : i forgot to use this function at many placed
func UpdatePlacedBuilding(userId string, placedBuildingID string) (PlacedBuilding, error) {
	var oldBuilding PlacedBuilding
	err := DB.Where("id = ? AND user_id = ?", placedBuildingID, userId).First(&oldBuilding).Error
	if err != nil {
		return oldBuilding, err
	}
	err = DB.Model(&PlacedBuilding{}).
		Where("id = ?", placedBuildingID).
		Update("last_updated_at", time.Now()).Error
	return oldBuilding, err
}
func DecreaseUpdateTime(userId string, placedBuildingID string, hours float64) error {
	var updatedBuilding PlacedBuilding
	secondsToSubtract := hours * 3600
	err := DB.Model(&updatedBuilding).
		Clauses(clause.Returning{}).
		Where("id = ? AND user_id = ?", placedBuildingID, userId).
		Update("last_updated_at", gorm.Expr("last_updated_at - ? * INTERVAL '1 second'", secondsToSubtract)).
		Error
	return err
}
func GetPlacedBuilding_ID_Level(userID string) ([]PlacedBuilding, error) {
	var placedBuildings []PlacedBuilding
	err := DB.Table("placed_buildings").
		Select("building_id, level").
		Where("user_id = ?", userID).
		Find(&placedBuildings).Error

	if err != nil {
		return nil, err
	}
	return placedBuildings, nil
}
func GetPlacedBuildingLevel(userID string, placedBuildingID string) (int, error) {
	var placedBuilding PlacedBuilding
	err := DB.Table("placed_building").
		Select("level").
		Where("user_id = ? AND building_id = ?", userID, placedBuildingID).
		First(&placedBuilding).Error

	if err != nil {
		return 0, err
	}
	return placedBuilding.Level, nil
}
func UpdatePlacedBuildingPosition(userId string, placed_building_id string, gridx, gridy int) (PlacedBuilding, error) {
	var updatedBuilding PlacedBuilding

	result := DB.Model(&updatedBuilding).
		Clauses(clause.Returning{}).
		Where("id = ? AND user_id = ?", placed_building_id, userId).
		Updates(map[string]interface{}{
			"grid_x":          gridx,
			"grid_y":          gridy,
			"last_updated_at": time.Now(),
		})

	if result.Error != nil {
		return PlacedBuilding{}, result.Error
	}
	if result.RowsAffected == 0 {
		return PlacedBuilding{}, errors.New("building not found or user unauthorized")
	}
	return updatedBuilding, nil
}
func GetNearByBuildings(userId string, x int, y int) ([]struct {
	At_x   int
	At_y   int
	Size_x int
	Size_y int
	Id     string
}, error) {
	const radius = 10

	var placedBuildings []PlacedBuilding

	result := DB.Where(
		"user_id = ? AND grid_x BETWEEN ? AND ? AND grid_y BETWEEN ? AND ?",
		userId,
		x-radius,
		x+radius,
		y-radius,
		y+radius,
	).Find(&placedBuildings)

	if result.Error != nil {
		return nil, result.Error
	}

	nearbyBuildings := make([]struct {
		At_x   int
		At_y   int
		Size_x int
		Size_y int
		Id     string
	}, 0, len(placedBuildings))

	for _, pb := range placedBuildings {
		size, exists := BuildingSize[pb.BuildingID]
		if !exists {
			log.Printf("Warning: Size not found for BuildingID %s. Defaulting to 1x1.\n", pb.BuildingID)
			size = struct {
				X int
				Y int
			}{X: 1, Y: 1}
		}

		nearbyBuildings = append(nearbyBuildings, struct {
			At_x   int
			At_y   int
			Size_x int
			Size_y int
			Id     string
		}{
			At_x:   pb.GridX,
			At_y:   pb.GridY,
			Size_x: size.X,
			Size_y: size.Y,
			Id:     pb.ID,
		})
	}

	return nearbyBuildings, nil
}
func SetBrokenBuilding(userID string, placedBuildingID string, isBroken bool, tx *gorm.DB) error {
	result := tx.Model(&PlacedBuilding{}).
		Where("id = ? AND user_id = ?", placedBuildingID, userID).
		Update("is_broken", isBroken)

	return result.Error
}
func SetBrokenBuildings(userID string, placedBuildingIDs []string, isBroken bool) error {
	if len(placedBuildingIDs) == 0 {
		return nil
	}
	result := DB.Model(&PlacedBuilding{}).
		Where("id IN ? AND user_id = ?", placedBuildingIDs, userID).
		Update("is_broken", isBroken)

	return result.Error
}
func IsBuildingBroken(userID string, placedBuildingID string) (bool, error) { // Note: fixed the "Borken" typo!
	var isBroken bool
	result := DB.Model(&PlacedBuilding{}).
		Select("is_broken").
		Where("id = ? AND user_id = ?", placedBuildingID, userID).
		Scan(&isBroken)

	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, gorm.ErrRecordNotFound
	}
	return isBroken, nil
}
