package Database

import (
	"Village_combat/GO/Models"
	"encoding/json"
	"errors"
	"fmt"
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
var TroopConfigs map[string]Models.TroopConfig
var TroopLevelDetails map[struct {
	ID    string
	Level int
}]Models.TroopLevelStats

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
	err = LoadStaticTroopData()
	if err != nil {
		panic("Unable to load Troops data.")
	}
}
func LoadStaticTroopData() error {
	var config []Models.TroopConfig
	var lvlstat []Models.TroopLevelStats

	err := DB.Table(Models.TroopConfig{}.TableName()).Find(&config).Error
	if err != nil {
		return err
	}
	err = DB.Table(Models.TroopLevelStats{}.TableName()).Find(&lvlstat).Error
	if err != nil {
		return err
	}
	TroopConfigs = make(map[string]Models.TroopConfig)
	TroopLevelDetails = make(map[struct {
		ID    string
		Level int
	}]Models.TroopLevelStats)
	for _, troopConfig := range config {
		TroopConfigs[troopConfig.ID] = troopConfig
	}
	for _, stats := range lvlstat {
		TroopLevelDetails[struct {
			ID    string
			Level int
		}{ID: stats.TroopID, Level: stats.Level}] = stats
	}
	return nil
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
	userStatus := &Models.UserStatus{
		UserID:       user.UserID,
		LastDefended: time.Now(),
		InBattle:     false,
		Power:        100,
	}

	if err := tx.Create(userStatus).Error; err != nil {
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
func GetPlacedBuildingJSON(userID string) (json.RawMessage, error) {
	var configs []Models.PlacedBuilding
	err := DB.Where("user_id = ?", userID).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(configs)
}
func GetPlacedBuildings(userID string) ([]Models.PlacedBuilding, error) {
	var configs []Models.PlacedBuilding
	err := DB.Where("user_id = ?", userID).Find(&configs).Error
	return configs, err
}
func GetPlacedBuilding(userID string, placedBuildingId string) (Models.PlacedBuilding, error) {
	var configs Models.PlacedBuilding
	err := DB.Where("user_id = ? AND id = ?", userID, placedBuildingId).Clauses(clause.Returning{}).First(&configs).Error
	return configs, err
}

// TODO : i forgot to use this function at many placed
func UpdatePlacedBuilding(userId string, placedBuildingID string) (Models.PlacedBuilding, error) {
	var oldBuilding Models.PlacedBuilding
	err := DB.Where("id = ? AND user_id = ?", placedBuildingID, userId).First(&oldBuilding).Error
	if err != nil {
		return oldBuilding, err
	}
	err = DB.Model(&Models.PlacedBuilding{}).
		Where("id = ?", placedBuildingID).
		Update("last_updated_at", time.Now()).Error
	return oldBuilding, err
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
func GetDefenceBuildingStatAndLevelStat(buildingId string, level int) (Models.DefenseBuildingLevelStats, Models.DefenseBuildingStats, error) {
	var buildingStats Models.DefenseBuildingStats
	var levelStats Models.DefenseBuildingLevelStats

	if err := DB.Where("building_id = ?", buildingId).First(&buildingStats).Error; err != nil {
		return levelStats, buildingStats, err
	}

	if err := DB.Where("building_id = ? AND level = ?", buildingId, level).First(&levelStats).Error; err != nil {
		return levelStats, buildingStats, err
	}

	return levelStats, buildingStats, nil
}
func GetBuildingHealth(buildingID string, level int) (int, error) {
	var health int
	err := DB.Model(&Models.BuildingLevelStats{}).
		Where("building_id = ? AND level = ?", buildingID, level).
		Pluck("health", &health).
		Error

	if err != nil {
		return 0, err
	}

	return health, nil
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
	err := tx.Select("user_id", "building_id", "grid_x", "grid_y", "level").Clauses(clause.Returning{}).Create(&newBuilding).Error
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
		if task.TaskType == Models.TroopTraining {
			err := AddTroopsToUser(userId, *task.TroopID, *task.TroopLevelTo, *task.TroopCount, tx)
			if err != nil {
				tx.Rollback()
				return nil, nil, err
			}
		} else if task.TaskType == Models.BauildingRepair {
			err = SetBrokenBuilding(userId, task.PlacedBuildingID, false, tx)
			if err != nil {
				tx.Rollback()
				return nil, nil, err
			}
		} else {
			buildingIDs = append(buildingIDs, task.PlacedBuildingID)
		}
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
	return DB.Table("user_data").
		Where("user_id = ?", userId).
		Update("town_hall_level", gorm.Expr("town_hall_level + 1")).
		Error
}
func IsConstructionUnderProgress(userId string, placed_building_id string) (bool, error) {
	n := int64(0)
	err := DB.Table(Models.ConstructionTask{}.TableName()).Where("user_id = ? AND placed_building_id = ?", userId, placed_building_id).Count(&n).Error
	return n > 0, err
}

func GetUserTrainedTroops(userId string) (json.RawMessage, error) {
	var troops []Models.TrainedTroop
	err := DB.Where("user_id = ?", userId).Find(&troops).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(troops)
}
func GetTroopUpgradeCost(Troop_id string, level_upgrade_to int) (Models.UpgradeCost, error) {
	var cost Models.UpgradeCost
	err := DB.Where("troop_id = ? AND upgrade_to_level = ?", Troop_id, level_upgrade_to).
		First(&cost).Error
	return cost, err
}
func AddTroopsToUser(userId string, troopId string, level int, count int, tx *gorm.DB) error {
	var troop Models.TrainedTroop
	err := tx.Where("user_id = ? AND troop_id = ? AND level = ?", userId, troopId, level).First(&troop).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newTroop := Models.TrainedTroop{
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
func SubtractTroopsOfUser(
	userId string,
	troopId string,
	level int,
	count int,
	tx *gorm.DB,
) (bool, error) {
	if count < 0 {
		return false, errors.New("cannot subtract a negative amount of troops")
	}
	if count == 0 {
		return true, nil
	}
	result := tx.Model(&Models.TrainedTroop{}).
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
	var troops []Models.TroopConfig
	err := DB.Preload("LevelStats").Preload("UpgradeCosts").Find(&troops).Error
	if err != nil {
		return nil, err
	}
	return json.Marshal(troops)
}
func StartTrainingTroops(userID string, troopId string, count int, placedBuildingId string, duration_seconds int, level_to int, tx *gorm.DB) (Models.ConstructionTask, error) {

	newTask := Models.ConstructionTask{
		UserID:           userID,
		TaskType:         Models.TroopTraining,
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
func GetCapacityDifference(building_id string, level1 int, level2 int) (int, error) {
	var stats []Models.ResourceBuildingLevelStats

	result := DB.Select("level, storage_capacity").
		Where("building_id = ? AND level IN (?, ?)", building_id, level1, level2).
		Find(&stats)

	if result.Error != nil {
		return 0, result.Error
	}

	var cap1, cap2 int
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
func AddUserGold(userId string, gold int) (Models.UserData, error) {
	var updatedUser Models.UserData

	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"current_gold": gorm.Expr("LEAST(current_gold + ?, total_gold_capacity)", gold),
			"updated_at":   time.Now(),
		}).Error

	return updatedUser, err
}

func AddUserElixir(userId string, elixir int) (Models.UserData, error) {
	var updatedUser Models.UserData
	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"current_elixir": gorm.Expr("LEAST(current_elixir + ?, total_elixir_capacity)", elixir),
			"updated_at":     time.Now(),
		}).Error
	return updatedUser, err
}
func AddUserDarkElixir(userId string, darkElixir int) (Models.UserData, error) {
	var updatedUser Models.UserData
	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"current_dark_elixir": gorm.Expr("LEAST(current_dark_elixir + ?, total_dark_elixir_capacity)", darkElixir),
			"updated_at":          time.Now(),
		}).Error
	return updatedUser, err
}
func AddUserGems(userId string, gems int) (Models.UserData, error) {
	var updatedUser Models.UserData

	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"current_gems": gorm.Expr("current_gems + ?", gems),
			"updated_at":   time.Now(),
		}).Error

	return updatedUser, err
}
func AddUserCapacity(userId string, gold_capacity int, elixir_capacity int, dark_elixir_capacity int) (Models.UserData, error) {
	var updatedUser Models.UserData

	err := DB.Model(&updatedUser).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"total_gold_capacity":        gorm.Expr("total_gold_capacity + ?", gold_capacity),
			"total_elixir_capacity":      gorm.Expr("total_elixir_capacity + ?", elixir_capacity),
			"total_dark_elixir_capacity": gorm.Expr("total_dark_elixir_capacity + ?", dark_elixir_capacity),
			"updated_at":                 time.Now(),
		}).Error

	return updatedUser, err
}
func GetGenerationRate(resourceBuildingId string, level int) (float64, error) {
	var stats Models.ResourceBuildingLevelStats
	err := DB.Select("generation_rate_per_hour").
		Where("building_id = ? AND level = ?", resourceBuildingId, level).
		First(&stats).Error
	return stats.GenerationRatePerHour, err
}

func SetUserBattleStatus(userID string, inBattle bool) error {
	return DB.Model(&Models.UserStatus{}).
		Where("user_id = ?", userID).
		Update("in_battle", inBattle).Error
}
func FindOpponent(attackerID string, powerRange int) (*Models.UserStatus, error) {
	var attacker Models.UserStatus
	var opponent Models.UserStatus

	err := DB.Where("user_id = ?", attackerID).First(&attacker).Error
	if err != nil {
		return nil, fmt.Errorf("attacker not found: %w", err)
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer tx.Rollback()

	subQuery := tx.Model(&Models.UserStatus{}).
		Select("user_id").
		Where("user_id != ?", attackerID).
		Where("in_battle = ?", false).
		Where("power BETWEEN ? AND ?", attacker.Power-powerRange, attacker.Power+powerRange).
		Order("last_defended ASC").
		Limit(10)

	err = tx.Model(&Models.UserStatus{}).
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

	err = tx.Model(&Models.UserStatus{}).
		Where("user_id = ?", opponent.UserID).
		Update("in_battle", true).Error

	if err != nil {
		return nil, fmt.Errorf("failed to mark in_battle: %w", err)
	}

	if err = tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &opponent, nil
}
func SetBrokenBuilding(userID string, placedBuildingID string, isBroken bool, tx *gorm.DB) error {
	result := tx.Model(&Models.PlacedBuilding{}).
		Where("id = ? AND user_id = ?", placedBuildingID, userID).
		Update("is_broken", isBroken)

	return result.Error
}
func SetBrokenBuildings(userID string, placedBuildingIDs []string, isBroken bool) error {
	if len(placedBuildingIDs) == 0 {
		return nil
	}
	result := DB.Model(&Models.PlacedBuilding{}).
		Where("id IN ? AND user_id = ?", placedBuildingIDs, userID).
		Update("is_broken", isBroken)

	return result.Error
}
func IsBuildingBroken(userID string, placedBuildingID string) (bool, error) { // Note: fixed the "Borken" typo!
	var isBroken bool
	result := DB.Model(&Models.PlacedBuilding{}).
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
func InsertBattleHistory(history Models.BattleHistory) (string, error) {
	result := DB.Create(&history)
	return history.BattleID, result.Error
}

func InsertBrokenBuildingBattleHistory(battleId, buildingId string, count int) error {
	record := Models.BuildingsBroken{
		BattleID:   battleId,
		BuildingID: buildingId,
		Count:      count,
	}
	return DB.Create(&record).Error
}
func InsertTroopLoosesBattleHistory(battleId, troopId string, count int) error {
	loss := Models.BattleTroopLoss{
		BattleID:  battleId,
		TroopID:   troopId,
		LossCount: count,
	}
	return DB.Create(&loss).Error
}
func GetUsername(userId string) (string, error) {
	var username string
	err := DB.Model(&Models.User{}).
		Where("user_id = ?", userId).
		Select("username").
		Row().
		Scan(&username)
	return username, err
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
