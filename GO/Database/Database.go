package Database

import (
	"Village_combat/GO/Models"
	"encoding/json"
	"errors"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DB *gorm.DB

// Global variables to hold the dynamic UUIDs for each building type
var (
	TownHall_ID          string
	Cannon_ID            string
	ArcherTower_ID       string
	AirDefense_ID        string
	GoldMine_ID          string
	GoldStorage_ID       string
	ElixirCollector_ID   string
	ElixirStorage_ID     string
	DarkElixirDrill_ID   string
	DarkElixirStorage_ID string
	Barracks_ID          string
)

func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	err = LoadStaticBuildingIDs()
	if err != nil {
		panic("Unable to load Buildings")
		return
	}
}

func LoadStaticBuildingIDs() error {
	var buildings []struct {
		BuildingID string `gorm:"column:building_id"`
		Name       string `gorm:"column:name"`
	}

	err := DB.Table(Models.BuildingConfigBase{}.TableName()).Select("building_id, name").Find(&buildings).Error
	if err != nil {
		return err
	}

	for _, b := range buildings {
		switch b.Name {
		case "Town Hall":
			TownHall_ID = b.BuildingID
		case "Cannon":
			Cannon_ID = b.BuildingID
		case "Archer Tower":
			ArcherTower_ID = b.BuildingID
		case "Air Defense":
			AirDefense_ID = b.BuildingID
		case "Gold Mine":
			GoldMine_ID = b.BuildingID
		case "Gold Storage":
			GoldStorage_ID = b.BuildingID
		case "Elixir Collector":
			ElixirCollector_ID = b.BuildingID
		case "Elixir Storage":
			ElixirStorage_ID = b.BuildingID
		case "Dark Elixir Drill":
			DarkElixirDrill_ID = b.BuildingID
		case "Dark Elixir Storage":
			DarkElixirStorage_ID = b.BuildingID
		case "Barracks":
			Barracks_ID = b.BuildingID
		}
	}

	log.Println("Static building IDs successfully loaded into memory.")
	return nil
}
func RegisterUser(username string, passwordHash string, email string) (*Models.User, *Models.UserData, error) {

	tx := DB.Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}

	user := Models.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}
	if err := tx.Select("username", "email", "password_hash").Clauses(clause.Returning{}).Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	userData := Models.UserData{
		UserID: user.UserID,
	}
	if err := tx.Select("user_id").Clauses(clause.Returning{}).Create(&userData).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	townhall := Models.PlacedBuilding{
		UserID:       user.UserID,
		BuildingID:   TownHall_ID,
		GridX:        0,
		GridY:        0,
		CurrentLevel: 1,
	}

	if err := tx.Select("user_id", "building_id", "grid_x", "grid_y", "current_level").Create(&townhall).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
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
	var existing Models.RefreshToken
	err := DB.Where("user_id = ?", userID).First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newToken := Models.RefreshToken{
			UserID:       userID,
			JWTTokenHash: tokenHash,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			ExpiresAt:    expireTime,
		}
		return DB.Select("user_id", "jwt_token_hash", "ip_address", "user_agent", "expires_at").Create(&newToken).Error
	}

	if err != nil {
		return err
	}

	return DB.Model(&existing).Select("jwt_token_hash", "ip_address", "user_agent", "expires_at", "is_used").Updates(Models.RefreshToken{
		JWTTokenHash: tokenHash,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    expireTime,
		IsUsed:       false,
	}).Error
}

func GetRefreshTokenByUserID(userID string) (*Models.RefreshToken, error) {
	var token Models.RefreshToken
	err := DB.Where("user_id = ?", userID).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func UpdatePlacedBuilding(building *Models.PlacedBuilding) error {
	return DB.Select("current_level", "dynamic_state", "last_updated_at").Save(building).Error
}
func GetAllBuildingConfigsJSON() ([]byte, error) {
	var configs []Models.BuildingConfigBase
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
func GetAllTroopConfigsJSON() ([]byte, error) {
	var configs []Models.TroopConfig
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

// region Unused Functions
//func UpdateUserResources(userData *Models.UserData) error {
//	return DB.Select("current_gold", "current_elixir", "current_dark_elixir", "current_gems", "updated_at").Save(userData).Error
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
