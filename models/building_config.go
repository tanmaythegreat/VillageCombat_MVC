package models

import (
	"encoding/json"
	"errors"
	"math"

	"gorm.io/gorm"
)

func GetAllBuildingConfigsJSON() (json.RawMessage, error) {
	var configs []BuildingConfigBase
	err := DB.Find(&configs).Error
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(configs)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}
func GetAllDefenceBuildingConfigsJSON() (json.RawMessage, error) {
	var configs []DefenseBuildingStats
	err := DB.Find(&configs).Error
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(configs)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
func GetAllArmyBuildingConfigsJSON() (json.RawMessage, error) {
	var configs []ArmyBuildingStats
	err := DB.Find(&configs).Error
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(configs)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
func GetAllResourceBuildingConfigsJSON() (json.RawMessage, error) {
	var configs []ResourceBuildingStats
	err := DB.Find(&configs).Error
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(configs)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
func GetBuildingDataOfLevelJSON(buildingID string, level int) (json.RawMessage, error) {
	response := struct {
		BuildingID  string             `json:"building_id"`
		Level       int                `json:"level"`
		BaseStats   BuildingLevelStats `json:"base_stats"`
		UpgradeCost UpgradeCost        `json:"upgrade_cost,omitempty"`
		Details     interface{}        `json:"details,omitempty"`
	}{
		BuildingID: buildingID,
		Level:      level,
	}
	if level != 0 {
		if err := DB.Where("building_id = ? AND level = ?", buildingID, level).First(&response.BaseStats).Error; err != nil {
			return nil, err
		}
	}

	err := DB.Where("building_id = ? AND upgrade_to_level = ?", buildingID, level+1).First(&response.UpgradeCost).Error
	if err != nil {
		response.UpgradeCost = UpgradeCost{
			ID:                    "",
			TroopID:               nil,
			BuildingID:            &buildingID,
			UpgradeToLevel:        level + 1,
			GoldRequired:          math.MaxInt32,
			ElixirRequired:        math.MaxInt32,
			DarkElixirRequired:    math.MaxInt32,
			OrGemRequired:         math.MaxInt32,
			TimeRequiredSeconds:   math.MaxInt32,
			TownHallLevelRequired: level + 1,
		}
	}

	category := BuildingID_Category[buildingID]

	switch category {
	case Resource:
		var stats ResourceBuildingLevelStats
		if err := DB.Where("building_id = ? AND level = ?", buildingID, level).First(&stats).Error; err == nil {
			response.Details = stats
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

	case Army:
		var stats ArmyBuildingLevelStats
		if err := DB.Where("building_id = ? AND level = ?", buildingID, level).First(&stats).Error; err == nil {
			response.Details = stats
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

	case Defense:
		var stats DefenseBuildingLevelStats
		if err := DB.Where("building_id = ? AND level = ?", buildingID, level).First(&stats).Error; err == nil {
			response.Details = stats
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

	case TownHall:
		response.Details = ""
	}

	return json.Marshal(response)
}
func GetDefenceBuildingStatAndLevelStat(buildingId string, level int) (DefenseBuildingLevelStats, DefenseBuildingStats, error) {
	var buildingStats DefenseBuildingStats
	var levelStats DefenseBuildingLevelStats

	if err := DB.Where("building_id = ?", buildingId).First(&buildingStats).Error; err != nil {
		return levelStats, buildingStats, err
	}

	if err := DB.Where("building_id = ? AND level = ?", buildingId, level).First(&levelStats).Error; err != nil {
		return levelStats, buildingStats, err
	}

	return levelStats, buildingStats, nil
}
func GetBuildingHealth(buildingID string, level int) (int64, error) {
	var health int64
	err := DB.Model(&BuildingLevelStats{}).
		Where("building_id = ? AND level = ?", buildingID, level).
		Pluck("health", &health).
		Error

	if err != nil {
		return 0, err
	}

	return health, nil
}
