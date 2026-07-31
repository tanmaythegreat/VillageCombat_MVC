package services

import (
	"log"

	"github.com/gorilla/websocket"
)

func SendError(conn *websocket.Conn, err error) error {
	errPayload := map[string]string{
		"status":  "error",
		"message": "Internal Server Error",
	}
	log.Print(err.Error())
	return conn.WriteJSON(errPayload)
}
