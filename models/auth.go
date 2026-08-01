package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

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
