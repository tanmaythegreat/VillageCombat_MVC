package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
