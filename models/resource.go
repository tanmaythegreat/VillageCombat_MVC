package models

import (
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetCapacityDifference(building_id string, level1 int, level2 int) (int64, error) {
	var stats []ResourceBuildingLevelStats

	result := DB.Select("level, storage_capacity").
		Where("building_id = ? AND level IN (?, ?)", building_id, level1, level2).
		Find(&stats)

	if result.Error != nil {
		return 0, result.Error
	}

	var cap1, cap2 int64
	found1, found2 := false, false
	for _, s := range stats {
		if s.Level == level1 {
			cap1 = s.StorageCapacity
			found1 = true
		} else if s.Level == level2 {
			cap2 = s.StorageCapacity
			found2 = true
		}
	}
	if !found1 || !found2 {
		log.Printf("Warning: Missing records for building %s (level1 found: %v, level2 found: %v)\n", building_id, found1, found2)
	}
	return cap2 - cap1, nil
}
func AddUserGold(userId string, gold int64) (UserData, error) {
	var updatedUser UserData
	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"current_gold": gorm.Expr("GREATEST(0, LEAST(current_gold + ?, total_gold_capacity))", gold),
			"updated_at":   time.Now(),
		}).Error
	return updatedUser, err
}
func AddUserElixir(userId string, elixir int64) (UserData, error) {
	var updatedUser UserData
	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"current_elixir": gorm.Expr("GREATEST(0, LEAST(current_elixir + ?, total_elixir_capacity))", elixir),
			"updated_at":     time.Now(),
		}).Error
	return updatedUser, err
}
func AddUserDarkElixir(userId string, darkElixir int64) (UserData, error) {
	var updatedUser UserData
	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"current_dark_elixir": gorm.Expr("GREATEST(0, LEAST(current_dark_elixir + ?, total_dark_elixir_capacity))", darkElixir),
			"updated_at":          time.Now(),
		}).Error
	return updatedUser, err
}
func AddUserGoldGetRemaining(userId string, gold int64) (UserData, error, int64) {
	var updatedUser UserData
	var extraGold int64

	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).First(&updatedUser).Error; err != nil {
			return err
		}
		newGold := updatedUser.CurrentGold + gold
		if newGold > updatedUser.TotalGoldCapacity {
			extraGold = newGold - updatedUser.TotalGoldCapacity
			updatedUser.CurrentGold = updatedUser.TotalGoldCapacity
		} else {
			extraGold = 0
			updatedUser.CurrentGold = newGold
		}
		return tx.Model(&updatedUser).Updates(map[string]interface{}{
			"current_gold": updatedUser.CurrentGold,
			"updated_at":   time.Now(),
		}).Error
	})

	return updatedUser, err, extraGold
}
func AddUserElixirGetRemaining(userId string, elixir int64) (UserData, error, int64) {
	var updatedUser UserData
	var extraElixir int64

	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).First(&updatedUser).Error; err != nil {
			return err
		}

		newElixir := updatedUser.CurrentElixir + elixir

		if newElixir > updatedUser.TotalElixirCapacity {
			extraElixir = newElixir - updatedUser.TotalElixirCapacity
			updatedUser.CurrentElixir = updatedUser.TotalElixirCapacity
		} else {
			extraElixir = 0
			updatedUser.CurrentElixir = newElixir
		}

		return tx.Model(&updatedUser).Updates(map[string]interface{}{
			"current_elixir": updatedUser.CurrentElixir,
			"updated_at":     time.Now(),
		}).Error
	})

	return updatedUser, err, extraElixir
}
func AddUserDarkElixirGetRemaining(userId string, darkElixir int64) (UserData, error, int64) {
	var updatedUser UserData
	var extraDarkElixir int64

	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userId).First(&updatedUser).Error; err != nil {
			return err
		}

		newDarkElixir := updatedUser.CurrentDarkElixir + darkElixir

		if newDarkElixir > updatedUser.TotalDarkElixirCapacity {
			extraDarkElixir = newDarkElixir - updatedUser.TotalDarkElixirCapacity
			updatedUser.CurrentDarkElixir = updatedUser.TotalDarkElixirCapacity
		} else {
			extraDarkElixir = 0
			updatedUser.CurrentDarkElixir = newDarkElixir
		}

		return tx.Model(&updatedUser).Updates(map[string]interface{}{
			"current_dark_elixir": updatedUser.CurrentDarkElixir,
			"updated_at":          time.Now(),
		}).Error
	})

	return updatedUser, err, extraDarkElixir
}
func AddUserGems(userId string, gems int64) (UserData, error) {
	var updatedUser UserData
	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"current_gems": gorm.Expr("GREATEST(current_gems + ?, 0)", gems),
			"updated_at":   time.Now(),
		}).Error
	return updatedUser, err
}
func AddUserCapacity(userId string, gold_capacity int64, elixir_capacity int64, dark_elixir_capacity int64, troop_capacity int) (UserData, error) {
	var updatedUser UserData

	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"total_gold_capacity":        gorm.Expr("total_gold_capacity + ?", gold_capacity),
			"total_elixir_capacity":      gorm.Expr("total_elixir_capacity + ?", elixir_capacity),
			"total_dark_elixir_capacity": gorm.Expr("total_dark_elixir_capacity + ?", dark_elixir_capacity),
			"total_troop_capacity":       gorm.Expr("total_troop_capacity + ?", troop_capacity),
			"updated_at":                 time.Now(),
		}).Error

	return updatedUser, err
}
func GetGenerationRate(resourceBuildingId string, level int) (float64, error) {
	var stats ResourceBuildingLevelStats
	err := DB.Select("generation_rate_per_hour").
		Where("building_id = ? AND level = ?", resourceBuildingId, level).
		First(&stats).Error
	return stats.GenerationRatePerHour, err
}
func UserPurchase(userId string, cost UpgradeCost, useGem bool) (*gorm.DB, error) {

	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var result *gorm.DB

	if useGem {
		result = tx.Model(&UserData{}).
			Where("user_id = ? AND current_gems >= ?", userId, cost.OrGemRequired).
			Update("current_gems", gorm.Expr("current_gems - ?", cost.OrGemRequired))
	} else {
		result = tx.Model(&UserData{}).
			Where("user_id = ? AND current_gold >= ? AND current_elixir >= ? AND current_dark_elixir >= ?",
				userId, cost.GoldRequired, cost.ElixirRequired, cost.DarkElixirRequired).
			Updates(map[string]interface{}{
				"current_gold":        gorm.Expr("current_gold - ?", cost.GoldRequired),
				"current_elixir":      gorm.Expr("current_elixir - ?", cost.ElixirRequired),
				"current_dark_elixir": gorm.Expr("current_dark_elixir - ?", cost.DarkElixirRequired),
			})
	}

	if result.Error != nil {
		tx.Rollback()
		log.Printf("controllers error during purchase for user %s: %v\n", userId, result.Error)
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.New("insufficient resources")
	}

	return tx, nil
}
