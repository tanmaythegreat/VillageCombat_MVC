package auth

import (
	"Village_combat/models"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserLoginCreds struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordText string `json:"password_text"`
}

type JWT_Token struct {
	AccessToken     string    `json:"access_token"`
	RefreshTokenB64 string    `json:"refresh_token_b64"`
	ExpiresAt       time.Time `json:"expires_at"`
}

var jwtSecretKey []byte

func init() {
	jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtSecretKey) == 0 {
		panic("JWT_SECRET_KEY environment variable is not set")
	}
}

func createAccessToken(userID string, duration time.Duration) (string, time.Time, error) {
	header := map[string]string{
		"algorithm": "HS256",
		"type":      "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	expireTime := time.Now().Add(duration)
	payload := map[string]interface{}{
		"user_id":   userID,
		"issued_at": time.Now().Unix(),
		"expire_at": expireTime.Unix(),
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	unsignedToken := headerB64 + "." + payloadB64

	mac := hmac.New(sha256.New, jwtSecretKey)
	mac.Write([]byte(unsignedToken))
	signatureBytes := mac.Sum(nil)

	signatureB64 := base64.RawURLEncoding.EncodeToString(signatureBytes)

	jwt := unsignedToken + "." + signatureB64
	return jwt, expireTime, nil
}
func GenerateJWT_Token(userId string, ipAddress string, userAgent string) (JWT_Token, error) {

	accessToken, expireTime, err := createAccessToken(userId, 15*time.Minute)
	if err != nil {
		return JWT_Token{}, err
	}

	plainRefreshToken := make([]byte, 64)
	_, err = rand.Read(plainRefreshToken)
	if err != nil {
		return JWT_Token{}, err
	}
	hashedToken, err := bcrypt.GenerateFromPassword(plainRefreshToken, bcrypt.DefaultCost)
	if err != nil {
		return JWT_Token{}, err
	}

	err = models.AddRefreshToken(userId, string(hashedToken), ipAddress, userAgent, expireTime)
	if err != nil {
		return JWT_Token{}, err
	}

	return JWT_Token{
		AccessToken:     accessToken,
		RefreshTokenB64: base64.RawURLEncoding.EncodeToString(plainRefreshToken),
		ExpiresAt:       expireTime,
	}, nil
}
func RefreshAccessToken(userID string, refreshTokenB64 string, ipAddress string, userAgent string) (JWT_Token, error) {

	storedTokenInfo, err := models.GetRefreshTokenByUserID(userID)
	if err != nil {
		return JWT_Token{}, errors.New("invalid session")
	}

	if time.Now().After(storedTokenInfo.ExpiresAt) || storedTokenInfo.IPAddress != ipAddress || storedTokenInfo.UserAgent != userAgent {
		return JWT_Token{}, errors.New("refresh token expired, please log in again")
	}

	refreshToken, err := base64.RawURLEncoding.DecodeString(refreshTokenB64)
	if err != nil {
		return JWT_Token{}, errors.New("invalid refresh token")
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedTokenInfo.JWTTokenHash), refreshToken)
	if err != nil {
		return JWT_Token{}, errors.New("invalid refresh token")
	}

	return GenerateJWT_Token(userID, ipAddress, userAgent)
}
func VerifyJWT_Token(token string) (string, bool) {
	splited := strings.Split(token, ".")
	if len(splited) != 3 {
		return "", false
	}
	headerB64, payloadB64, signatureB64 := splited[0], splited[1], splited[2]
	signingInput := headerB64 + "." + payloadB64
	mac := hmac.New(sha256.New, jwtSecretKey)
	mac.Write([]byte(signingInput))
	expectedSignature := mac.Sum(nil)
	actualSignature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil || !hmac.Equal(actualSignature, expectedSignature) {
		return "", false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return "", false
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", false
	}
	userIDRaw, ok := payload["user_id"]
	if !ok {
		return "", false
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		return "", false
	}
	expireRaw, ok := payload["expire_at"]
	if !ok {
		return "", false
	}
	expireTime, ok := expireRaw.(float64)
	if !ok {
		return "", false
	}
	if time.Now().Unix() > int64(expireTime) {
		return "", false
	}
	return userID, true
}
