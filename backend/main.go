package main

import (
	"breeze-backend/internal/db"
	"breeze-backend/internal/handlers"
	"breeze-backend/internal/sync"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from parent directory
	if err := godotenv.Load("../.env"); err != nil {
		// Try loading from current directory
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	// Initialize database
	database, err := db.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Printf("Warning: failed to run migrations: %v", err)
		log.Println("Continuing anyway...")
	}

	// Initialize sync service
	syncService := sync.NewService(database)

	// Initialize handlers
	handler := handlers.NewHandler(database, syncService)

	// Set up routes
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", handler.HealthCheck).Methods("GET", "OPTIONS")

	// Venue routes
	r.HandleFunc("/api/v1/venues", handler.GetVenues).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/venues/{venue_id}", handler.GetVenue).Methods("GET", "OPTIONS")

	// Availability routes
	r.HandleFunc("/api/v1/venues/{venue_id}/availability", handler.GetAvailability).Methods("GET", "OPTIONS")

	// Dates routes
	r.HandleFunc("/api/v1/venues/{venue_id}/dates", handler.GetDates).Methods("GET", "OPTIONS")

	// Sync routes
	r.HandleFunc("/api/v1/venues/{venue_id}/sync-status", handler.GetSyncStatus).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/venues/{venue_id}/sync", handler.TriggerSync).Methods("POST", "OPTIONS")

	// Date management routes
	r.HandleFunc("/api/v1/dates", handler.CreateDate).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/dates/{date_id}/switch-venue", handler.SwitchVenue).Methods("POST", "OPTIONS")

	// Apply CORS middleware
	handlerWithCORS := handlers.EnableCORS(r)

	// Start background sync goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Running scheduled sync...")
				syncService.ProcessAllVenues()
			}
		}
	}()

	// Start server
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handlerWithCORS,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("🚀 Server starting on http://localhost:%s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🏢 Venues API: http://localhost:%s/api/v1/venues\n", port)
	log.Fatal(server.ListenAndServe())
}
