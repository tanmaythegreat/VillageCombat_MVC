package Database

import (
	"Village_combat/GO/Models"
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

func GetUsersPlacedBuildings(user *Models.User) ([]Models.PlacedBuilding, error) {
	var buildings []Models.PlacedBuilding
	err := DB.Omit("dynamic_state").Where("user_id = ?", user.UserID).Find(&buildings).Error
	return buildings, err
}

func AddRefreshToken(userID string, tokenHash string, ipAddress string, userAgent string) error {
	tokenRecord := Models.RefreshToken{
		UserID:       userID,
		JWTTokenHash: tokenHash,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(24 * 7 * time.Hour),
	}
	return DB.Select("user_id", "jwt_token_hash", "ip_address", "user_agent", "expires_at").Create(&tokenRecord).Error
}

func UpdateUserResources(userData *Models.UserData) error {
	return DB.Select("current_gold", "current_elixir", "current_dark_elixir", "current_gems", "updated_at").Save(userData).Error
}

func UpdatePlacedBuilding(building *Models.PlacedBuilding) error {
	return DB.Select("current_level", "dynamic_state", "last_updated_at").Save(building).Error
}

func UpdateUserTownHallLevel(userData *Models.UserData) error {
	return DB.Select("town_hall_level", "updated_at").Save(userData).Error
}
