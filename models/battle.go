package models

import (
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SetUserBattleStatus(userID string, inBattle bool) error {
	return DB.Model(&UserStatus{}).
		Where("user_id = ?", userID).
		Update("in_battle", inBattle).Error
}
func FindOpponent(attackerID string, powerRange int) (*UserStatus, error) {
	var attacker UserStatus
	var opponent UserStatus

	err := DB.Where("user_id = ?", attackerID).First(&attacker).Error
	if err != nil {
		return nil, fmt.Errorf("attacker not found: %w", err)
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer tx.Rollback()

	subQuery := tx.Model(&UserStatus{}).
		Select("user_id").
		Where("user_id != ?", attackerID).
		Where("in_battle = ?", false).
		Where("defence_power BETWEEN ? AND ?", attacker.AttackPower-powerRange, attacker.AttackPower+powerRange).
		Order("last_defended ASC").
		Limit(10)

	err = tx.Model(&UserStatus{}).
		Where("user_id IN (?)", subQuery).
		Order("RANDOM()").
		Limit(1).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "SKIP LOCKED",
		}).
		First(&opponent).Error

	if err != nil {
		return nil, fmt.Errorf("no opponent found: %w", err)
	}

	if err = tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	logErr("StartMatch: could not clear attacker battle status", SetUserBattleStatus(attackerID, true))
	logErr("StartMatch: could not clear defender battle status", SetUserBattleStatus(opponent.UserID, true))

	return &opponent, nil
}

func InsertBattleHistory(history BattleHistory) (string, error) {
	result := DB.Create(&history)
	return history.BattleID, result.Error
}
func InsertBrokenBuildingBattleHistory(battleId, buildingId string, count int) error {
	record := BuildingsBroken{
		BattleID:   battleId,
		BuildingID: buildingId,
		Count:      count,
	}
	return DB.Create(&record).Error
}
func InsertTroopLoosesBattleHistory(battleId, troopId string, count int, isAttacker bool) error {
	loss := BattleTroopLoss{
		BattleID:   battleId,
		TroopID:    troopId,
		LossCount:  count,
		IsAttacker: isAttacker,
	}
	return DB.Create(&loss).Error
}
func SaveBattleRecord(record *BattleRecord) error {
	return DB.Save(record).Error
}
func GetBattleRecord(battleID string) (*BattleRecord, error) {
	var record BattleRecord
	result := DB.Where("battle_id = ?", battleID).First(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("battle record not found")
		}
		return nil, result.Error
	}
	return &record, nil
}
func GetBattleHistory(battleID string) (*BattleHistory, error) {
	var battle BattleHistory
	err := DB.Preload("TroopLosses").
		Preload("BrokenBuildings").
		First(&battle, "battle_id = ?", battleID).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &battle, nil
}
func GetBattleHistories(userId string, lastFoughtAt time.Time, toLoad int) []BattleHistory {
	var histories []BattleHistory
	if lastFoughtAt.IsZero() {
		lastFoughtAt = time.Now().Add(100 * time.Hour)
	}
	username, err := GetUsername(userId)
	if err != nil {
		log.Printf("Error occurred while finding battle history of user %s\n", userId)
		return histories
	}
	err = DB.Where("(attacker_name = ? OR defender_name = ?) AND (fought_at < ?)",
		username, username, lastFoughtAt).
		Order("fought_at DESC").
		Limit(toLoad).
		Find(&histories).Error
	if err != nil {
		log.Printf("Error occurred while finding battle history of user %s\n", userId)
		return histories
	}

	return histories
}
func AdjustAttackPower(userID string, delta int) error {
	return DB.Model(&UserStatus{}).
		Where("user_id = ?", userID).
		Update("attack_power", gorm.Expr("attack_power + ?", delta)).
		Error
}
func AdjustDefencePower(userID string, delta int) error {
	return DB.Model(&UserStatus{}).
		Where("user_id = ?", userID).
		Update("defence_power", gorm.Expr("defence_power + ?", delta)).
		Error
}
func MarkDefendedNow(userID string) error {
	return DB.Model(&UserStatus{}).
		Where("user_id = ?", userID).
		Update("last_defended", time.Now()).
		Error
}
