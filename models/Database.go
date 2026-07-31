package models

import (
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
var BuildingID_Category map[string]BuildingCategory
var BuildingSize map[string]struct {
	X int
	Y int
}
var TroopConfigs map[string]TroopConfig
var TroopLevelDetails map[struct {
	ID    string
	Level int
}]TroopLevelStats
var ResourceLevelDetails map[struct {
	ID    string
	Level int
}]ResourceBuildingLevelStats

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
	err = LoadStaticResourceData()
	if err != nil {
		panic("Unable to load Resource data.")
	}
}
func LoadStaticTroopData() error {
	var config []TroopConfig
	var lvlstat []TroopLevelStats

	err := DB.Table(TroopConfig{}.TableName()).Find(&config).Error
	if err != nil {
		return err
	}
	err = DB.Table(TroopLevelStats{}.TableName()).Find(&lvlstat).Error
	if err != nil {
		return err
	}
	TroopConfigs = make(map[string]TroopConfig)
	TroopLevelDetails = make(map[struct {
		ID    string
		Level int
	}]TroopLevelStats)
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
func LoadStaticResourceData() error {
	var lvlstat []ResourceBuildingLevelStats
	err := DB.Table(ResourceBuildingLevelStats{}.TableName()).Find(&lvlstat).Error
	if err != nil {
		return err
	}
	ResourceLevelDetails = make(map[struct {
		ID    string
		Level int
	}]ResourceBuildingLevelStats)
	for _, stats := range lvlstat {
		ResourceLevelDetails[struct {
			ID    string
			Level int
		}{ID: stats.BuildingID, Level: stats.Level}] = stats
	}
	return nil
}
func LoadStaticBuildingIDsAndCategory() error {
	var buildings []struct {
		BuildingID string           `gorm:"column:building_id"`
		Category   BuildingCategory `gorm:"column:category"`
		GridSizeX  int              `gorm:"column:grid_size_x"`
		GridSizeY  int              `gorm:"column:grid_size_y"`
		Name       string           `gorm:"column:name"`
	}

	err := DB.Table(BuildingConfigBase{}.TableName()).Select("building_id,category,grid_size_x,grid_size_y, name").Find(&buildings).Error
	if err != nil {
		return err
	}
	BuildingID_Category = make(map[string]BuildingCategory)
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
func RegisterUser(username string, passwordHash string, email string) (*User, *UserData, error) {

	tx := DB.Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}

	user := User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}
	if err := tx.Select("username", "email", "password_hash").Clauses(clause.Returning{}).Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	userData := UserData{
		UserID: user.UserID,
	}
	if err := tx.Select("user_id").Clauses(clause.Returning{}).Create(&userData).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	townhall := PlacedBuilding{
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
	userStatus := &UserStatus{
		UserID:       user.UserID,
		LastDefended: time.Now(),
		InBattle:     false,
		AttackPower:  100,
		DefencePower: 100,
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
func GetUserByName(username string) (*User, error) {
	var user User
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
func GetUserByEmail(email string) (*User, error) {
	var user User
	err := DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
func GetUserData(userId string) (UserData, error) {
	var userData UserData
	err := DB.Where("user_id = ?", userId).First(&userData).Error
	return userData, err
}
func AddRefreshToken(userID string, tokenHash string, ipAddress string, userAgent string, expireTime time.Time) error {
	var existing RefreshToken
	err := DB.Where("user_id = ?", userID).First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newToken := RefreshToken{
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

	return DB.Model(&existing).Select("jwt_token_hash", "ip_address", "user_agent", "expires_at", "is_used").Updates(RefreshToken{
		JWTTokenHash: tokenHash,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    expireTime,
		IsUsed:       false,
	}).Error
}
func RemoveRefreshToken(userID string) error {
	err := DB.Where("user_id = ?", userID).Delete(&RefreshToken{}).Error
	if err != nil {
		return err
	}
	return nil
}
func GetRefreshTokenByUserID(userID string) (*RefreshToken, error) {
	var token RefreshToken
	err := DB.Where("user_id = ?", userID).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}
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

	logErr("StartMatch: could not clear attacker battle status", SetUserBattleStatus(attackerID, true))
	logErr("StartMatch: could not clear defender battle status", SetUserBattleStatus(opponent.UserID, true))

	if err = tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &opponent, nil
}

func logErr(context string, err error) {
	if err != nil {
		log.Printf("%s: %v", context, err)
	}
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
func GetUsername(userId string) (string, error) {
	var username string
	err := DB.Model(&User{}).
		Where("user_id = ?", userId).
		Select("username").
		Row().
		Scan(&username)
	return username, err
}
func GetUser(userId string) (User, error) {
	var user User
	err := DB.Model(&User{}).
		Where("user_id = ?", userId).
		First(&user).Error
	return user, err
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
