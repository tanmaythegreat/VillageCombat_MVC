package models

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
