package Auth

import (
	"Village_combat/GO/Battle"
	"Village_combat/GO/Database"
	"Village_combat/GO/Models"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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
	// TODO : check if the password is strong enough
	// TODO : check is the user name valid that is is it already taken,their should be no weird symbols like \{(&%$#';" etc
	// TODO : may be Email Verification
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

	jwt, err := generateJWT_Token(registerUser.UserID, ipAddress, userAgent)
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
func LoginHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var userLoginCreds UserLoginCreds
	if err := json.NewDecoder(request.Body).Decode(&userLoginCreds); err != nil {
		http.Error(writer, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if (userLoginCreds.Username == "" && userLoginCreds.Email == "") || userLoginCreds.PasswordText == "" {
		http.Error(writer, "Missing required fields: username, email, or password", http.StatusBadRequest)
		return
	}

	var registerUser *Models.User
	var err error

	if userLoginCreds.Username != "" {
		registerUser, err = Database.GetUserByName(userLoginCreds.Username)
	} else {
		registerUser, err = Database.GetUserByEmail(userLoginCreds.Email)
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

	jwt, err := generateJWT_Token(registerUser.UserID, ipAddress, userAgent)
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
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var requestData struct {
		UserID       string `json:"user_id"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	ip := r.RemoteAddr
	ua := r.Header.Get("User-Agent")
	newToken, err := refreshAccessToken(requestData.UserID, requestData.RefreshToken, ip, ua)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
func generateJWT_Token(userId string, ipAddress string, userAgent string) (JWT_Token, error) {

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

	err = Database.AddRefreshToken(userId, string(hashedToken), ipAddress, userAgent, expireTime)
	if err != nil {
		return JWT_Token{}, err
	}

	return JWT_Token{
		AccessToken:     accessToken,
		RefreshTokenB64: base64.RawURLEncoding.EncodeToString(plainRefreshToken),
		ExpiresAt:       expireTime,
	}, nil
}
func refreshAccessToken(userID string, refreshTokenB64 string, ipAddress string, userAgent string) (JWT_Token, error) {

	storedTokenInfo, err := Database.GetRefreshTokenByUserID(userID)
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

	return generateJWT_Token(userID, ipAddress, userAgent)
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	var token = r.URL.Query().Get("token")
	userId, verified := VerifyJWT_Token(token)
	if !verified {
		http.Error(w, "Invalid Token.", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to establish WebSocket connection", http.StatusBadRequest)
		return
	}
	Mu := sync.Mutex{}
	var chanel chan []byte = make(chan []byte)
	Battle.Manager.Mu.Lock()

	Battle.Manager.Connections[userId] = Battle.Connection{Conn: conn, Mu: &Mu, Ch: chanel}
	Battle.Manager.Mu.Unlock()
	defer conn.Close()
	defer delete(Battle.Manager.Connections, userId)
	fmt.Printf("Client successfully connected to WebSocket server!, client : %s\n", userId)

	go func() {
		defer close(chanel)
		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				break
			}
			chanel <- p
		}
	}()
Loop:
	for {
		messageType := websocket.TextMessage
		Mu.Lock()
		p := <-chanel
		Mu.Unlock()
		if p == nil || len(p) == 0 {
			errPayload := []byte(`{"status": "error", "message": "Action payload cannot be empty"}`)
			err = conn.WriteMessage(messageType, errPayload)
			if err != nil {
				log.Println("Failed to send message to client:", err)
				break Loop
			}
			continue
		}
		fmt.Printf("Received from player: %s\n", string(p))
		var payload struct {
			Action      string `json:"action"`
			Message     string `json:"message"`
			AccessToken string `json:"access_token"`
		}
		err = json.Unmarshal(p, &payload)
		if err != nil {
			errPayload := []byte(`{"status": "error", "message": "Invalid JSON."}`)
			err = conn.WriteMessage(messageType, errPayload)
			if err != nil {
				log.Println("Failed to send message to client:", err)
				break Loop
			}
			continue
		}
		userid, verified := VerifyJWT_Token(payload.AccessToken)
		if !verified || userId != userid {
			errPayload := []byte(`{"status": "error", "message": "UnAuthorised."}`)
			err = conn.WriteMessage(messageType, errPayload)
			if err != nil {
				log.Println("Failed to send message to client:", err)
				break Loop
			}
			continue
		}
	Switch:
		switch payload.Action {
		case "INITIAL_LOAD":
			troops, err := Database.GetUserTrainedTroops(userId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
			}
			building, err := Database.GetPlacedBuildingJSON(userId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
			}
			constructionTasks, err := Database.GetConstructionTasks(userId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
			}
			UserData, err := Database.GetUserData(userId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
			}
			// TODO : check was i attacked when i was offline.
			data := struct {
				MsgType           string          `json:"msg_type"`
				Building          json.RawMessage `json:"building"`
				Troops            json.RawMessage `json:"troops"`
				ConstructionTasks json.RawMessage `json:"construction_tasks"`
				UserData          Models.UserData `json:"user_data"`
			}{
				MsgType:           "building_troop_of_user",
				Building:          building,
				Troops:            troops,
				ConstructionTasks: constructionTasks,
				UserData:          UserData,
			}
			err = conn.WriteJSON(data)
			if err != nil {
				log.Println("Failed to send message to client:", err)
				break Loop
			}
		case "ALL_BUILDING_TROOP_DATA":
			buildings, err := Database.GetAllBuildingConfigsJSON()
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
			}
			troops, err := Database.GetAllTroopsDataJSON()
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
			}
			defence, err := Database.GetAllDefenceBuildingConfigsJSON()
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
			}
			army, err := Database.GetAllArmyBuildingConfigsJSON()
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
			}
			resource, err := Database.GetAllResourceBuildingConfigsJSON()
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
			}
			id_level, err := Database.GetPlacedBuilding_ID_Level(userId)

			configMap := make(map[string]json.RawMessage)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			for _, il := range id_level {
				key := fmt.Sprintf("%s:%d", il.BuildingID, il.Level)
				if _, exists := configMap[key]; exists {
					continue
				}
				jsonData, err := Database.GetBuildingDataOfLevelJSON(il.BuildingID, il.Level)
				if err != nil {
					errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
					err = conn.WriteMessage(messageType, errPayload)
					if err != nil {
						log.Println("Failed to send message to client:", err)
						break Loop
					}
					break Switch
				}
				configMap[key] = jsonData
			}
			for id, _ := range Database.BuildingSize {
				key := fmt.Sprintf("%s:%d", id, 1)
				if _, exists := configMap[key]; exists {
					continue
				}
				jsonData, err := Database.GetBuildingDataOfLevelJSON(id, 1)
				if err != nil {
					errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
					err = conn.WriteMessage(messageType, errPayload)
					if err != nil {
						log.Println("Failed to send message to client:", err)
						break Loop
					}
					break Switch
				}
				configMap[key] = jsonData
			}
			data := struct {
				MsgType             string                     `json:"msg_type"`
				Building            json.RawMessage            `json:"building"`
				Troops              json.RawMessage            `json:"troops"`
				Defence             json.RawMessage            `json:"defence"`
				Army                json.RawMessage            `json:"army"`
				Resource            json.RawMessage            `json:"resource"`
				ParticularLevelData map[string]json.RawMessage `json:"particular_level_data"`
			}{
				MsgType:             "building_troop",
				Building:            buildings,
				Troops:              troops,
				Defence:             defence,
				Army:                army,
				Resource:            resource,
				ParticularLevelData: configMap,
			}
			err = conn.WriteJSON(data)
			if err != nil {
				log.Println("Failed to send message to client:", err)
				break Loop
			}
		case "BUILDING_LEVEL_DETAIL":
			var placedBuildingId = payload.Message
			level, err := Database.GetPlacedBuildingLevel(userId, placedBuildingId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			dataOfLevel, err := Database.GetBuildingDataOfLevelJSON(placedBuildingId, level)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			err = conn.WriteJSON(struct {
				MsgType     string          `json:"msg_type"`
				DataOfLevel json.RawMessage `json:"data_of_level"`
			}{
				MsgType:     "building_level_detail",
				DataOfLevel: dataOfLevel,
			})
			if err != nil {
				log.Println("Failed to send message to client:", err)
				break Loop
			}
		case "CREATE_BUILDING":
			var data struct {
				BuildingID string `json:"building_id"`
				X          int    `json:"x"`
				Y          int    `json:"y"`
				UseGems    bool   `json:"use_gems"`
			}
			err := json.Unmarshal([]byte(payload.Message), &data)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Invalid JSON."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			nearByBuildings, err := Database.GetNearByBuildings(userId, data.X, data.Y)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			newSize, exists := Database.BuildingSize[data.BuildingID]
			if !exists {
				errPayload := []byte(`{"status": "error", "message": "Unknown Building."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			hasCollision := false
			for _, b := range nearByBuildings {
				if data.X < b.At_x+b.Size_x && data.X+newSize.X > b.At_x &&
					data.Y < b.At_y+b.Size_y && data.Y+newSize.Y > b.At_y {
					hasCollision = true
					break
				}
			}

			if hasCollision {
				errPayload := []byte(`{"status": "error", "message": "Can't Place Here."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			// TODO : townhall level check
			cost, err := Database.GetConstructionCost(data.BuildingID, 1)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			userData, err := Database.GetUserData(userId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if userData.TownHallLevel < cost.TownHallLevelRequired {
				errPayload := []byte(`{"status": "error", "message": "Town Hall Level_from not sufficient."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			tx, err := Database.UserPurchase(userId, cost, data.UseGems)
			if errors.Is(err, errors.New("insufficient resources")) {
				errPayload := []byte(`{"status": "error", "message": "Insufficient Resources."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			} else if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			var time_req int = cost.TimeRequiredSeconds
			if data.UseGems {
				time_req = 1
			}
			placedBuilding, task, err := Database.ConstructBuilding(userId, data.BuildingID, data.X, data.Y, tx, time_req)
			if err != nil {
				tx.Rollback()
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			tx.Commit()
			err = conn.WriteJSON(map[string]interface{}{
				"msg_type":        "construction_started",
				"placed_building": placedBuilding,
				"task":            task,
			})
			if err != nil {
				log.Println("Failed to send message to client:", err)
			}
		case "CHECK_CONSTRUCTION_WORK":
			constructionTasks, buildings_updated, err := Database.CheckIsConstructionComplete(userId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if len(constructionTasks) > 0 {
				var gold_capacity_increment = 0
				var dark_elixir_capacity_increment = 0
				var elixir_capacity_increment = 0
				var levelDetails = make([]json.RawMessage, 0, len(buildings_updated))
				for _, building := range buildings_updated {
					if Database.BuildingID_Category[building.BuildingID] == Models.TownHall {
						if err = Database.IncrementUserTownHallLevel(userId); err != nil {
							errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
							err = conn.WriteMessage(messageType, errPayload)
							if err != nil {
								log.Println("Failed to send message to client:", err)
								break Loop
							}
							break Switch
						}
					} else if Database.BuildingID_Category[building.BuildingID] == Models.Resource {
						if building.BuildingID == Database.ElixirStorage_ID {
							difference, err := Database.GetCapacityDifference(building.BuildingID, building.Level-1, building.Level)
							if err != nil {
								errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
								err = conn.WriteMessage(messageType, errPayload)
								if err != nil {
									log.Println("Failed to send message to client:", err)
									break Loop
								}
								break Switch
							}
							elixir_capacity_increment += difference
						} else if building.BuildingID == Database.GoldStorage_ID {
							difference, err := Database.GetCapacityDifference(building.BuildingID, building.Level-1, building.Level)
							if err != nil {
								errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
								err = conn.WriteMessage(messageType, errPayload)
								if err != nil {
									log.Println("Failed to send message to client:", err)
									break Loop
								}
								break Switch
							}
							gold_capacity_increment += difference
						} else if building.BuildingID == Database.DarkElixirStorage_ID {
							difference, err := Database.GetCapacityDifference(building.BuildingID, building.Level-1, building.Level)
							if err != nil {
								errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
								err = conn.WriteMessage(messageType, errPayload)
								if err != nil {
									log.Println("Failed to send message to client:", err)
									break Loop
								}
								break Switch
							}
							dark_elixir_capacity_increment += difference
						}
					}
					levelJSON, err := Database.GetBuildingDataOfLevelJSON(building.BuildingID, building.Level)
					if err != nil {
						errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
						err = conn.WriteMessage(messageType, errPayload)
						if err != nil {
							log.Println("Failed to send message to client:", err)
							break Loop
						}
						break Switch
					}
					levelDetails = append(levelDetails, levelJSON)
				}
				userData, err := Database.AddUserCapacity(userId, gold_capacity_increment, elixir_capacity_increment, dark_elixir_capacity_increment)
				if err != nil {
					return
				}
				err = conn.WriteJSON(map[string]interface{}{
					"msg_type":                "construction_completed",
					"particular_level_detail": levelDetails,
					"construction_done":       constructionTasks,
					"user_data":               userData,
				})
				if err != nil {
					log.Println("Failed to send message to client:", err)
				}
			}
		case "UPGRADE_BUILDING":
			var data struct {
				PlacedBuildingID string `json:"placed_building_id"`
				UseGems          bool   `json:"use_gems"`
			}
			err := json.Unmarshal([]byte(payload.Message), &data)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Invalid JSON."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			constructionUnderProgress, err := Database.IsConstructionUnderProgress(userId, data.PlacedBuildingID)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if constructionUnderProgress {
				errPayload := []byte(`{"status": "error", "message": "Building already under construction."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			isBroken, err := Database.IsBuildingBroken(userId, data.PlacedBuildingID)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if isBroken {
				errPayload := []byte(`{"status": "error", "message": "Cannot upgrade broken building."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			building, err := Database.GetPlacedBuilding(userId, data.PlacedBuildingID)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			cost, err := Database.GetConstructionCost(building.BuildingID, building.Level+1)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			userData, err := Database.GetUserData(userId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if userData.TownHallLevel < cost.TownHallLevelRequired {
				errPayload := []byte(`{"status": "error", "message": "Town Hall Level_from not sufficient."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			tx, err := Database.UserPurchase(userId, cost, data.UseGems)
			if errors.Is(err, errors.New("insufficient resources")) {
				errPayload := []byte(`{"status": "error", "message": "Insufficient Resources."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			} else if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			var time_req int = cost.TimeRequiredSeconds
			if data.UseGems {
				time_req = 1
			}
			task, err := Database.UpgradeBuilding(userId, data.PlacedBuildingID, tx, time_req)
			if err != nil {
				tx.Rollback()
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			tx.Commit()
			err = conn.WriteJSON(map[string]interface{}{
				"msg_type": "construction_started",
				"task":     task,
			})
			if err != nil {
				log.Println("Failed to send message to client:", err)
			}
		case "TRAIN_TROOP":
			var data struct {
				BarrackPlacedBuildingID string `json:"barrack_placed_building_id"`
				TroopId                 string `json:"troop_id"`
				LevelFrom               int    `json:"level_from"`
				Count                   int    `json:"count"`
				UseGems                 bool   `json:"use_gems"`
			}
			err = json.Unmarshal([]byte(payload.Message), &data)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Invalid JSON."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			underProgress, err := Database.IsConstructionUnderProgress(userId, data.BarrackPlacedBuildingID)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if underProgress {
				errPayload := []byte(`{"status": "error", "message": "Already Building Construction work or Training going on here."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			building, err := Database.GetPlacedBuilding(userId, data.BarrackPlacedBuildingID)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if building.BuildingID != Database.Barracks_ID {
				errPayload := []byte(`{"status": "error", "message": "Can only Train in Barracks."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			isBroken, err := Database.IsBuildingBroken(userId, building.ID)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if isBroken {
				errPayload := []byte(`{"status": "error", "message": "Cannot train in broken barracks."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			upgradeCost, err := Database.GetTroopUpgradeCost(data.TroopId, data.LevelFrom+1)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			upgradeCost.TimeRequiredSeconds *= data.Count
			upgradeCost.DarkElixirRequired *= data.Count
			upgradeCost.ElixirRequired *= data.Count
			upgradeCost.GoldRequired *= data.Count
			upgradeCost.OrGemRequired *= data.Count
			tx, err := Database.UserPurchase(userId, upgradeCost, data.UseGems)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if data.LevelFrom != 0 {
				success, err := Database.SubtractTroopsOfUser(userId, data.TroopId, data.LevelFrom, data.Count, tx)
				if err != nil {
					tx.Rollback()
					errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
					err = conn.WriteMessage(messageType, errPayload)
					if err != nil {
						log.Println("Failed to send message to client:", err)
						break Loop
					}
					break Switch
				}
				if !success {
					tx.Rollback()
					errPayload := []byte(`{"status": "error", "message": "Not enough Troops."}`)
					err = conn.WriteMessage(messageType, errPayload)
					if err != nil {
						log.Println("Failed to send message to client:", err)
						break Loop
					}
					break Switch
				}
			}
			var time_req = upgradeCost.TimeRequiredSeconds
			if data.UseGems {
				time_req = 1
			}
			constructionTask, err := Database.StartTrainingTroops(userId, data.TroopId, data.Count, data.BarrackPlacedBuildingID, data.LevelFrom+1, time_req, tx)
			if err != nil {
				tx.Rollback()
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if tx.Commit().Error != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			err = conn.WriteJSON(map[string]interface{}{
				"msg_type": "troop_training_started",
				"task":     constructionTask,
			})
			if err != nil {
				log.Println("Failed to send message to client:", err)
				break Loop
			}
		case "COLLECT_RESOURCE":
			var data struct {
				PlacedBuildingId string `json:"placed_building_id"`
			}
			err = json.Unmarshal([]byte(payload.Message), &data)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Invalid JSON."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			isBroken, err := Database.IsBuildingBroken(userId, data.PlacedBuildingId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if isBroken {
				errPayload := []byte(`{"status": "error", "message": "Cannot collect from broken building."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			placedBuilding, err := Database.UpdatePlacedBuilding(userId, data.PlacedBuildingId)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			var dt = time.Now().Sub(placedBuilding.LastUpdatedAt).Hours()
			generationRate, err := Database.GetGenerationRate(placedBuilding.BuildingID, placedBuilding.Level)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			var amount = dt * generationRate
			var user Models.UserData
			if placedBuilding.BuildingID == Database.GoldMine_ID {
				user, err = Database.AddUserGold(userId, int(amount))
			} else if placedBuilding.BuildingID == Database.ElixirCollector_ID {
				user, err = Database.AddUserElixir(userId, int(amount))
			} else if placedBuilding.BuildingID == Database.DarkElixirDrill_ID {
				user, err = Database.AddUserDarkElixir(userId, int(amount))
			}
			err = conn.WriteJSON(map[string]interface{}{
				"msg_type":  "resource_collected",
				"user_data": user,
			})
			if err != nil {
				log.Println("Failed to send message to client:", err)
			}
		case "ATTACK":
			err := Database.SetUserBattleStatus(userId, true)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			opponent, err := Database.FindOpponent(userId, 10)
			if err != nil || opponent == nil {
				conn.WriteJSON(map[string]interface{}{
					"msg_type": "un_attack",
				})
				break
			}
			Battle.StartMatch(userId, opponent.UserID)
		case "DEFEND":
			time.Sleep(time.Second)
		case "REVENGE":
			var data struct {
				OpponentID string
			}
			err := json.Unmarshal([]byte(payload.Message), &data)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Invalid JSON."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if data.OpponentID == userId {
				// TODO : allow revenge this way :)
				errPayload := []byte(`{"status": "error", "message": "You want to take revenge with yourself WOW!!."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			_, err = Database.GetUserData(data.OpponentID) // just assuring that the user exist
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "It was a Ghost,that wasn't a user!!\nCOC ki aatma"}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			Battle.StartMatch(userId, data.OpponentID)
		case "REPAIR_BUILDING":
			var data struct {
				PlacedBuildingID string `json:"placed_building_id"`
				UseGems          bool   `json:"use_gems"`
			}
			err := json.Unmarshal([]byte(payload.Message), &data)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Invalid JSON."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			constructionUnderProgress, err := Database.IsConstructionUnderProgress(userId, data.PlacedBuildingID)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if constructionUnderProgress {
				errPayload := []byte(`{"status": "error", "message": "Building already under construction."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			isBroken, err := Database.IsBuildingBroken(userId, data.PlacedBuildingID)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			if !isBroken {
				errPayload := []byte(`{"status": "error", "message": "Cannot Repair unbroken building."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			building, err := Database.GetPlacedBuilding(userId, data.PlacedBuildingID)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			cost, err := Database.GetConstructionCost(building.BuildingID, building.Level+1)
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			cost.TimeRequiredSeconds /= 10
			cost.OrGemRequired /= 10
			cost.OrGemRequired += 1
			cost.GoldRequired /= 10
			cost.ElixirRequired /= 10
			cost.DarkElixirRequired /= 10
			tx, err := Database.UserPurchase(userId, cost, data.UseGems)
			if errors.Is(err, errors.New("insufficient resources")) {
				errPayload := []byte(`{"status": "error", "message": "Insufficient Resources."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			} else if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			var time_req int = cost.TimeRequiredSeconds
			if data.UseGems {
				time_req = 1
			}
			task, err := Database.StartConstruction_Building(userId, Models.BauildingRepair, data.PlacedBuildingID, time_req, tx)
			if err != nil {
				tx.Rollback()
				errPayload := []byte(`{"status": "error", "message": "Internal Server Error."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			tx.Commit()
			err = conn.WriteJSON(map[string]interface{}{
				"msg_type": "construction_started",
				"task":     task,
			})
			if err != nil {
				log.Println("Failed to send message to client:", err)
			}
		default:
			err = conn.WriteJSON(map[string]interface{}{
				"status":  "error",
				"message": payload,
			})
			if err != nil {
				log.Println("Failed to send message to client:", err)
				break Loop
			}
		}
	}
}
