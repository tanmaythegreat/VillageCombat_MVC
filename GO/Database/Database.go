package Database

import (
	"Village_combat/GO/Models"
	"encoding/json"
	"errors"
	"log"
	"math"
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
var BuildingID_Category map[string]Models.BuildingCategory
var BuildingSize map[string]struct {
	X int
	Y int
}

func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	err = LoadStaticBuildingIDsAndCategory()
	if err != nil {
		panic("Unable to load Buildings")
		return
	}
}

func LoadStaticBuildingIDsAndCategory() error {
	var buildings []struct {
		BuildingID string                  `gorm:"column:building_id"`
		Category   Models.BuildingCategory `gorm:"column:category"`
		GridSizeX  int                     `gorm:"column:grid_size_x"`
		GridSizeY  int                     `gorm:"column:grid_size_y"`
		Name       string                  `gorm:"column:name"`
	}

	err := DB.Table(Models.BuildingConfigBase{}.TableName()).Select("building_id,category,grid_size_x,grid_size_y, name").Find(&buildings).Error
	if err != nil {
		return err
	}
	BuildingID_Category = make(map[string]Models.BuildingCategory)
	BuildingSize = make(map[string]struct {
		X int
		Y int
	})
	for _, b := range buildings {
		BuildingID_Category[b.BuildingID] = b.Category
		BuildingSize[b.BuildingID] = struct {
			X int
			Y int
		}{
			X: b.GridSizeX, Y: b.GridSizeY,
		}
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
		UserID:     user.UserID,
		BuildingID: TownHall_ID,
		GridX:      0,
		GridY:      0,
		Level:      1,
	}

	if err := tx.Select("user_id", "building_id", "grid_x", "grid_y", "level").Create(&townhall).Error; err != nil {
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
func GetUserData(userId string) (Models.UserData, error) {
	var userData Models.UserData
	err := DB.Where("user_id = ?", userId).First(&userData).Error
	return userData, err
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

func GetAllBuildingConfigsJSON() (json.RawMessage, error) {
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
func GetAllDefenceBuildingConfigsJSON() (json.RawMessage, error) {
	var configs []Models.DefenseBuildingStats
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
	var configs []Models.ArmyBuildingStats
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
	var configs []Models.ResourceBuildingStats
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
func GetAllTroopConfigsJSON() (json.RawMessage, error) {
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
func GetPlacedBuildingJSON(userID string) (json.RawMessage, error) {
	var configs []Models.PlacedBuilding
	err := DB.Where("user_id = ?", userID).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(configs)
}
func GetPlacedBuilding(userID string, placedBuildingId string) (Models.PlacedBuilding, error) {
	var configs Models.PlacedBuilding
	err := DB.Where("user_id = ? AND id = ?", userID, placedBuildingId).Clauses(clause.Returning{}).First(&configs).Error
	return configs, err
}
func GetTrainedTroopsJSON(userID string) (json.RawMessage, error) {
	var configs []Models.TrainedTroop
	err := DB.Where("user_id = ?", userID).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(configs)
}
func GetPlacedBuilding_ID_Level(userID string) ([]Models.PlacedBuilding, error) {
	var placedBuildings []Models.PlacedBuilding
	err := DB.Table("placed_buildings").
		Select("building_id, level").
		Where("user_id = ?", userID).
		Find(&placedBuildings).Error

	if err != nil {
		return nil, err
	}
	return placedBuildings, nil
}
func GetBuildingDataOfLevelJSON(buildingID string, level int) (json.RawMessage, error) {
	response := struct {
		BuildingID  string                    `json:"building_id"`
		Level       int                       `json:"level"`
		BaseStats   Models.BuildingLevelStats `json:"base_stats"`
		UpgradeCost Models.UpgradeCost        `json:"upgrade_cost,omitempty"`
		Details     interface{}               `json:"details,omitempty"`
	}{
		BuildingID: buildingID,
		Level:      level,
	}

	if err := DB.Where("building_id = ? AND level = ?", buildingID, level).First(&response.BaseStats).Error; err != nil {
		return nil, err
	}

	err := DB.Where("building_id = ? AND upgrade_to_level = ?", buildingID, level+1).First(&response.UpgradeCost).Error
	if err != nil {
		response.UpgradeCost = Models.UpgradeCost{
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
	case Models.Resource:
		var stats Models.ResourceBuildingLevelStats
		if err := DB.Where("building_id = ? AND level = ?", buildingID, level).First(&stats).Error; err == nil {
			response.Details = stats
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

	case Models.Army:
		var stats Models.ArmyBuildingLevelStats
		if err := DB.Where("building_id = ? AND level = ?", buildingID, level).First(&stats).Error; err == nil {
			response.Details = stats
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

	case Models.Defense:
		var stats Models.DefenseBuildingLevelStats
		if err := DB.Where("building_id = ? AND level = ?", buildingID, level).First(&stats).Error; err == nil {
			response.Details = stats
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

	case Models.TownHall:
		response.Details = ""
	}

	return json.Marshal(response)
}
func GetPlacedBuildingLevel(userID string, placedBuildingID string) (int, error) {
	var placedBuilding Models.PlacedBuilding
	err := DB.Table("placed_building").
		Select("level").
		Where("user_id = ? AND building_id = ?", userID, placedBuildingID).
		First(&placedBuilding).Error

	if err != nil {
		return 0, err
	}
	return placedBuilding.Level, nil
}
func GetConstructionTasks(userID string) (json.RawMessage, error) {
	var constructionTasks []Models.ConstructionTask
	err := DB.Where("user_id = ?", userID).Find(&constructionTasks).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(constructionTasks)
}
func ConstructBuilding(userID string, buildingID string, x int, y int, tx *gorm.DB, duration int) (Models.PlacedBuilding, Models.ConstructionTask, error) {
	newBuilding := Models.PlacedBuilding{
		UserID:     userID,
		BuildingID: buildingID,
		GridX:      x,
		GridY:      y,
		Level:      0,
	}
	err := tx.Select("user_id", "building_id", "grid_x", "grid_y", "current_level").Clauses(clause.Returning{}).Create(&newBuilding).Error
	if err != nil {
		return newBuilding, Models.ConstructionTask{}, err
	}
	task, err := StartConstruction_Building(userID, Models.BuildingConstruction, newBuilding.ID, duration, tx)
	return newBuilding, task, err
}
func UpgradeBuilding(userID string, placed_buildingID string, tx *gorm.DB, duration int) (Models.ConstructionTask, error) {
	task, err := StartConstruction_Building(userID, Models.BuildingUpgrade, placed_buildingID, duration, tx)
	return task, err
}
func GetNearByBuildings(userId string, x int, y int) ([]struct {
	At_x   int
	At_y   int
	Size_x int
	Size_y int
}, error) {
	const radius = 10

	var placedBuildings []Models.PlacedBuilding

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
		}{
			At_x:   pb.GridX,
			At_y:   pb.GridY,
			Size_x: size.X,
			Size_y: size.Y,
		})
	}

	return nearbyBuildings, nil
}
func StartConstruction_Building(userID string, taskType Models.ConstructionType, placedBuildingId string, duration_seconds int, tx *gorm.DB) (Models.ConstructionTask, error) {

	newTask := Models.ConstructionTask{
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
func GetConstructionCost(building_id string, upgrade_to int) (Models.UpgradeCost, error) {
	var cost Models.UpgradeCost
	err := DB.Where("building_id = ? AND upgrade_to_level = ?", building_id, upgrade_to).
		First(&cost).Error
	return cost, err
}

func UserPurchase(userId string, cost Models.UpgradeCost, useGem bool) (*gorm.DB, error) {

	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var result *gorm.DB

	if useGem {
		result = tx.Model(&Models.UserData{}).
			Where("user_id = ? AND current_gems >= ?", userId, cost.OrGemRequired).
			Update("current_gems", gorm.Expr("current_gems - ?", cost.OrGemRequired))
	} else {
		result = tx.Model(&Models.UserData{}).
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
		log.Printf("Database error during purchase for user %s: %v\n", userId, result.Error)
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.New("insufficient resources")
	}

	return tx, nil
}
func CheckIsConstructionComplete(userId string) ([]Models.ConstructionTask, []Models.PlacedBuilding, error) {
	var completedTasks []Models.ConstructionTask
	var updatedBuildings []Models.PlacedBuilding

	tx := DB.Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	defer tx.Rollback()

	err := tx.Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Where("task_type IN ?", []Models.ConstructionType{Models.BuildingConstruction, Models.BuildingUpgrade}).
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
		buildingIDs = append(buildingIDs, task.PlacedBuildingID)
	}
	err = tx.Clauses(clause.Returning{}).
		Model(&Models.PlacedBuilding{}).
		Where("id IN ?", buildingIDs).
		Update("level", gorm.Expr("level + 1")).
		Scan(&updatedBuildings).Error

	if err != nil {
		log.Printf("Error updating building levels for user %s: %v\n", userId, err)
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction for user %s: %v\n", userId, err)
		return nil, nil, err
	}
	return completedTasks, updatedBuildings, nil
}
func IncrementUserTownHallLevel(userId string) error {
	return DB.Table("users").
		Where("user_id = ?", userId).
		Update("town_hall_level", gorm.Expr("town_hall_level + 1")).
		Error
}
func IsConstructionUnderProgress(userId string, placed_building_id string) (bool, error) {
	n := int64(0)
	err := DB.Table(Models.ConstructionTask{}.TableName()).Where("user_id = ? AND placed_building_id = ?").Count(&n).Error
	return n > 0, err

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
