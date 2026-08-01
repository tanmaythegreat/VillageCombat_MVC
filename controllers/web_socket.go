package controllers

import (
	"Village_combat/auth"
	"Village_combat/battle"
	"Village_combat/models"
	"Village_combat/services/buildings"
	"Village_combat/services/resources"
	"Village_combat/services/troops"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to establish WebSocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()
	_, p, err := conn.ReadMessage()
	var payload struct {
		Action      string `json:"action"`
		Message     string `json:"message"`
		AccessToken string `json:"access_token"`
	}
	err = json.Unmarshal(p, &payload)
	if err != nil {
		errPayload := []byte(`{"status": "error", "message": "Invalid JSON."}`)
		err = conn.WriteMessage(websocket.TextMessage, errPayload)
		return
	}

	var token = payload.AccessToken
	userId, verified := auth.VerifyJWT_Token(token)
	if !verified {
		http.Error(w, "Invalid Token.", http.StatusUnauthorized)
		return
	}
	err = conn.WriteJSON(map[string]interface{}{
		"msg_type": "authorised",
	})
	if err != nil {
		log.Println("Failed to send message to client:", err)
		return
	}
	Mu := sync.Mutex{}
	var chanel chan []byte = make(chan []byte)
	battle.Manager.Mu.Lock()
	battle.Manager.Connections[userId] = battle.Connection{Conn: conn, Mu: &Mu, Ch: chanel}
	battle.Manager.Mu.Unlock()
	log.Printf("user joined Manager state:%v\n", battle.Manager)

	defer func() {
		battle.Manager.Mu.Lock()
		delete(battle.Manager.Connections, userId)
		battle.Manager.Mu.Unlock()
	}()
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
		userid, verified := auth.VerifyJWT_Token(payload.AccessToken)
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
			if buildings.InitialLoad(userId, conn) != nil {
				break Loop
			}
		case "ALL_BUILDING_TROOP_DATA":
			if buildings.AllBuildingDataLoad(userId, conn) != nil {
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
			if buildings.CreateBuilding(userId, data, conn) != nil {
				break Loop
			}
		case "CHECK_CONSTRUCTION_WORK":
			if buildings.CheckConstructionWork(userid, conn) != nil {
				break Loop
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
			if buildings.UpgradeBuilding(userId, data, conn) != nil {
				break Loop
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
			if troops.TrainTroop(userId, data, conn) != nil {
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
			if resources.CollectResource(userId, data, conn) != nil {
				break Loop
			}
		case "ATTACK":
			opponent, err := models.FindOpponent(userId, 10)
			if err != nil || opponent == nil {
				conn.WriteJSON(map[string]interface{}{
					"msg_type": "un_attack",
				})
				conn.WriteJSON(map[string]interface{}{
					"status":  "error",
					"message": "could not find opponent",
				})
			} else {
				battle.StartMatch(userId, opponent.UserID)
			}
		case "DEFEND":
			time.Sleep(time.Second)
		case "REVENGE":
			var OpponentName string = payload.Message

			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "Invalid JSON."}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}

			opponent, err := models.GetUserByName(OpponentName) // just assuring that the user exist
			if err != nil {
				errPayload := []byte(`{"status": "error", "message": "It was a Ghost,that wasn't a user!!\nCOC ki aatma"}`)
				err = conn.WriteMessage(messageType, errPayload)
				if err != nil {
					log.Println("Failed to send message to client:", err)
					break Loop
				}
				break Switch
			}
			//if opponent.Username==username{
			//	errPayload := []byte(`{"status": "error", "message": "I absolutely sympathise with you but pls dont destroy your own village.Every thing will be okay."}`)
			//	err = conn.WriteMessage(messageType, errPayload)
			//	if err != nil {
			//		log.Println("Failed to send message to client:", err)
			//		break Loop
			//	}
			//	break Switch
			//}
			battle.StartMatch(userId, opponent.UserID)
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
			if buildings.RepairBuilding(userId, data, conn) != nil {
				break Loop
			}
		case "REPLAY":
			if battle.Replay(payload.Message, conn) != nil {
				break Loop
			}
		case "BATTLE_HISTORY":
			var data struct {
				LastFoughtAt time.Time `json:"fought_at"`
				ToLoad       int       `json:"to_load"`
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
			err = conn.WriteJSON(map[string]interface{}{
				"msg_type": "battle_history",
				"history":  models.GetBattleHistories(userId, data.LastFoughtAt, data.ToLoad),
			})
			if err != nil {
				break Loop
			}
		case "MOVE":
			var data struct {
				PlacedBuildingID string `json:"placed_building_id"`
				GridX            int    `json:"grid_x"`
				GridY            int    `json:"grid_y"`
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
			if buildings.MoveBuilding(userId, data, conn) != nil {
				break Loop
			}
		case "LOGOUT":
			_ = models.RemoveRefreshToken(userId)
			break Loop
		case "COLLECT_ALL":
			if resources.CollectAllResource(userId, conn) != nil {
				break Loop
			}
		default:
			err = conn.WriteJSON(map[string]interface{}{
				"status":  "error",
				"message": payload,
			})
			if err != nil {
				break Loop
			}
		}
	}
}
