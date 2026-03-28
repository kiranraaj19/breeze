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
	authHandler := handlers.NewAuthHandler(database)
	dateHandler := handlers.NewDateHandler(database)

	// Set up routes
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", handler.HealthCheck).Methods("GET", "OPTIONS")

	// Auth routes (public)
	r.HandleFunc("/api/v1/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST", "OPTIONS")

	// Venue routes (public)
	r.HandleFunc("/api/v1/venues", handler.GetVenues).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/venues/{venue_id}", handler.GetVenue).Methods("GET", "OPTIONS")

	// Availability routes (public)
	r.HandleFunc("/api/v1/venues/{venue_id}/availability", handler.GetAvailability).Methods("GET", "OPTIONS")

	// Protected routes - require authentication
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(authHandler.AuthMiddleware)

	// User profile routes (protected)
	protected.HandleFunc("/users/me", authHandler.GetMe).Methods("GET", "OPTIONS")
	protected.HandleFunc("/users/me/dates", authHandler.GetMyDates).Methods("GET", "OPTIONS")

	// Date management routes (protected)
	protected.HandleFunc("/dates", handler.CreateDate).Methods("POST", "OPTIONS")
	
	// Date modification routes (cancel, reschedule, switch venue)
	protected.HandleFunc("/dates/{date_id}/cancel", dateHandler.CancelDate).Methods("POST", "OPTIONS")
	protected.HandleFunc("/dates/{date_id}/reschedule", dateHandler.RescheduleDate).Methods("POST", "OPTIONS")
	protected.HandleFunc("/dates/{date_id}/reschedule-options", dateHandler.GetRescheduleOptions).Methods("GET", "OPTIONS")
	protected.HandleFunc("/dates/{date_id}/switch-venue", dateHandler.SwitchVenue).Methods("POST", "OPTIONS")
	protected.HandleFunc("/dates/{date_id}/switch-options", dateHandler.GetVenuesForSwitch).Methods("GET", "OPTIONS")

	// Partner dashboard routes (public for now, could be protected with different auth)
	r.HandleFunc("/api/v1/venues/{venue_id}/dates", handler.GetDates).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/venues/{venue_id}/sync-status", handler.GetSyncStatus).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/venues/{venue_id}/sync", handler.TriggerSync).Methods("POST", "OPTIONS")

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
	fmt.Printf("🔐 Auth API: http://localhost:%s/api/v1/auth/login\n", port)
	fmt.Printf("🏢 Venues API: http://localhost:%s/api/v1/venues\n", port)
	fmt.Printf("👤 User API: http://localhost:%s/api/v1/users/me (protected)\n", port)
	fmt.Printf("📅 Date API: http://localhost:%s/api/v1/dates (protected)\n", port)
	log.Fatal(server.ListenAndServe())
}
