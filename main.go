package main

import (
	"Village_combat/controllers"
	"Village_combat/models"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	log.SetOutput(os.Stdout)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://admin_tentellam:i_wont_tell_you@localhost:5432/VillageGameDB?sslmode=disable"
	}

	log.Println("Checking for pending database migrations...")
	m, err := migrate.New("file://db/migrations", dbURL)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply database migrations: %v", err)
	}
	log.Println("controllers migrations applied successfully!")

	models.InitDB(dbURL)

	http.HandleFunc("/register", controllers.RegisterHandler)
	http.HandleFunc("/login", controllers.LoginHandler)
	http.HandleFunc("/ws", controllers.HandleWebSocket)
	http.HandleFunc("/refresh", controllers.RefreshHandler)

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Village Combat server deployed on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server collapsed under siege: %v", err)
	}
}
