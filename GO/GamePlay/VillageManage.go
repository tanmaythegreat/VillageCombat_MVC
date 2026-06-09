package GamePlay

import (
	"encoding/json"
	"net/http"
)

func GetAllPlacedBuildings(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var userLoginCreds struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(request.Body).Decode(&userLoginCreds); err != nil {
		http.Error(writer, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
}
