package Auth

import (
	"Village_combat/GO/Database"
	"Village_combat/GO/Models"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserLoginCreds struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordText string `json:"password_text"`
}

type JWT_Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func RegisterHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload UserLoginCreds
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		http.Error(writer, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.Username == "" || payload.Email == "" || payload.PasswordText == "" {
		http.Error(writer, "Missing required fields: username, email, password", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.PasswordText), bcrypt.DefaultCost)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	registerUser, _, err := Database.RegisterUser(payload.Username, string(hashedPassword), payload.Email)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusConflict)
		return
	}

	ipAddress := request.RemoteAddr
	userAgent := request.UserAgent()

	jwt, err := GenerateJWT_Token(registerUser, ipAddress, userAgent)
	if err != nil {
		http.Error(writer, "Failed to generate token session", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(writer).Encode(jwt); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func GenerateJWT_Token(user *Models.User, ipAddress string, userAgent string) (JWT_Token, error) {
	mockAccessToken := "signed_jwt_access_string_here"
	mockRefreshToken := "random_cryptographic_refresh_string"
	mockTokenHash := "hashed_version_of_refresh_token_for_db"

	err := Database.AddRefreshToken(user.UserID, mockTokenHash, ipAddress, userAgent)
	if err != nil {
		return JWT_Token{}, err
	}

	tokenResponse := JWT_Token{
		AccessToken:  mockAccessToken,
		RefreshToken: mockRefreshToken,
		ExpiresAt:    time.Now().Add(24 * 7 * time.Hour),
	}

	return tokenResponse, nil
}
