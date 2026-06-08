package main

import (
	"Village_combat/GO/Auth"
	"Village_combat/GO/Database"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	Database.InitDB("postgres://admin_tentellam:i_wont_tell_you@localhost:5432/VillageGameDB?sslmode=disable")

	http.HandleFunc("/register", Auth.RegisterHandler)
	http.HandleFunc("/login", Auth.LoginHandler)

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	log.Println("Village Combat server deployed on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server collapsed under siege: %v", err)
	}
}
