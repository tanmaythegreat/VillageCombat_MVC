package main

import (
	"Village_combat/controllers"
	"Village_combat/middleware"
	"Village_combat/models"
	"Village_combat/services/authentication"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	log.SetOutput(os.Stdout)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://admin_tentellam:i_wont_tell_you@localhost:5432/VillageGameDB?sslmode=disable"
		log.Println("DATABASE_URL env variable not set")
	}
	log.Println("Using database URL:", dbURL)

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
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			middleware.Cleanup()
		}
	}()
	http.HandleFunc("/register", middleware.RateLimit(authentication.RegisterHandler))
	http.HandleFunc("/login", middleware.RateLimit(authentication.LoginHandler))
	http.HandleFunc("/ws", middleware.RateLimit(controllers.HandleWebSocket))
	http.HandleFunc("/refresh", middleware.RateLimit(authentication.RefreshHandler))

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
