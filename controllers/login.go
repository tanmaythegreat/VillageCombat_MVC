package controllers

import (
	"Village_combat/auth"
	"Village_combat/models"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var userLoginCreds auth.UserLoginCreds
	if err := json.NewDecoder(request.Body).Decode(&userLoginCreds); err != nil {
		http.Error(writer, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if (userLoginCreds.Username == "" && userLoginCreds.Email == "") || userLoginCreds.PasswordText == "" {
		http.Error(writer, "Missing required fields: username, email, or password", http.StatusBadRequest)
		return
	}

	var registerUser *models.User
	var err error

	if userLoginCreds.Username != "" {
		registerUser, err = models.GetUserByName(userLoginCreds.Username)
	} else {
		registerUser, err = models.GetUserByEmail(userLoginCreds.Email)
	}

	if err != nil {
		_ = bcrypt.CompareHashAndPassword([]byte("this is here to prevent time based attacks."), []byte(userLoginCreds.PasswordText))
		http.Error(writer, "Invalid username/email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(registerUser.PasswordHash), []byte(userLoginCreds.PasswordText))
	if err != nil {
		http.Error(writer, "Invalid username/email or password", http.StatusUnauthorized)
		return
	}

	ipAddress := request.RemoteAddr
	userAgent := request.UserAgent()

	jwt, err := auth.GenerateJWT_Token(registerUser.UserID, ipAddress, userAgent)
	if err != nil {
		http.Error(writer, "Failed to generate token session", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(writer).Encode(jwt); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
