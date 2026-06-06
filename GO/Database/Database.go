package Database

import (
	"Village_combat/GO/Models"
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
}

func RegisterUser(username string, passwordHash string, email string) (*Models.User, *Models.UserData, error) {
	user := Models.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}
	if err := DB.Select("username", "email", "password_hash").Create(&user).Error; err != nil {
		return nil, nil, err
	}
	userData := Models.UserData{
		UserID: user.UserID,
	}
	if err := DB.Select("user_id").Create(&userData).Error; err != nil {
		return nil, nil, err
	}
	return &user, &userData, nil
}

func GetUserByName(username string) (*Models.User, error) {
	var user Models.User
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func GetUserByEmail(email string) (*Models.User, error) {
	var user Models.User
	err := DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func AddRefreshToken(userID string, tokenHash string, ipAddress string, userAgent string, expireTime time.Time) error {
	tokenRecord := Models.RefreshToken{
		UserID:       userID,
		JWTTokenHash: tokenHash,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    expireTime,
	}
	return DB.Select("user_id", "jwt_token_hash", "ip_address", "user_agent", "expires_at").Create(&tokenRecord).Error
}

func GetRefreshTokenByUserID(userID string) (*Models.RefreshToken, error) {
	var token Models.RefreshToken
	err := DB.Where("user_id = ?", userID).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func DeleteRefreshToken(userID string) error {
	return DB.Where("user_id = ?", userID).Delete(&Models.RefreshToken{}).Error
}

// region Unused Functions
//func UpdateUserResources(userData *Models.UserData) error {
//	return DB.Select("current_gold", "current_elixir", "current_dark_elixir", "current_gems", "updated_at").Save(userData).Error
//}
//func UpdatePlacedBuilding(building *Models.PlacedBuilding) error {
//	return DB.Select("current_level", "dynamic_state", "last_updated_at").Save(building).Error
//}
//func UpdateUserTownHallLevel(userData *Models.UserData) error {
//	return DB.Select("town_hall_level", "updated_at").Save(userData).Error
//}
//func SaveBattleResult(battle *Models.BattleHistory) error {
//	return DB.Select(
//		"attacker_id",
//		"defender_id",
//		"elixir_looted",
//		"gold_looted",
//		"dark_elixir_looted",
//	).Create(battle).Error
//}
//func SaveBattleTroopLosses(losses []Models.BattleTroopLoss) error {
//	return DB.Select("battle_id", "troop_id", "loss_count").Create(&losses).Error
//}
//func SaveBrokenBuildings(broken []Models.BuildingsBroken) error {
//	return DB.Select("battle_id", "placed_building_id").Create(&broken).Error
//}
//func GetAttackerBattleHistory(userID string) ([]Models.BattleHistory, error) {
//	var battles []Models.BattleHistory
//	err := DB.Omit("TroopLosses", "BrokenBuildings").
//		Where("attacker_id = ?", userID).
//		Order("fought_at DESC").
//		Find(&battles).Error
//	return battles, err
//}
//func GetDefenderBattleHistory(userID string) ([]Models.BattleHistory, error) {
//	var battles []Models.BattleHistory
//	err := DB.Omit("TroopLosses", "BrokenBuildings").
//		Where("defender_id = ?", userID).
//		Order("fought_at DESC").
//		Find(&battles).Error
//	return battles, err
//}
//func GetFullBattleDetail(battleID string) (*Models.BattleHistory, error) {
//	var battle Models.BattleHistory
//	err := DB.
//		Preload("TroopLosses").
//		Preload("BrokenBuildings").
//		Where("battle_id = ?", battleID).
//		First(&battle).Error
//	return &battle, err
//}
//func GetUsersPlacedBuildings(user *Models.User) ([]Models.PlacedBuilding, error) {
//	var buildings []Models.PlacedBuilding
//	err := DB.Omit("dynamic_state").Where("user_id = ?", user.UserID).Find(&buildings).Error
//	return buildings, err
//}
//
// endregion
