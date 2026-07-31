package authentication

import (
	"Village_combat/auth"
	"Village_combat/models"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload auth.UserLoginCreds
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		http.Error(writer, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.Username == "" || payload.Email == "" || payload.PasswordText == "" {
		http.Error(writer, "Missing required fields: username, email, password", http.StatusBadRequest)
		return
	}
	if err := auth.ValidatePasswordStrength(payload.PasswordText); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO : check is the user name valid that is is it already taken,their should be no weird symbols like \{(&%$#';" etc
	// TODO : may be Email Verification
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.PasswordText), bcrypt.DefaultCost)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	registerUser, _, err := models.RegisterUser(payload.Username, string(hashedPassword), payload.Email)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusConflict)
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
	writer.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(writer).Encode(jwt); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
