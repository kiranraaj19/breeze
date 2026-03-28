// Mock Partner Reservation API
// This simulates the external partner API for testing

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Reservation represents a reservation in the mock API
type Reservation struct {
	ID        string    `json:"reservation_id"`
	VenueID   string    `json:"venue_id"`
	Date      string    `json:"date"`
	Start     string    `json:"start"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	reservations = make(map[string]*Reservation)
	mu           sync.RWMutex
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/venues/{venue_id}/reservations", createReservation).Methods("POST")
	r.HandleFunc("/api/v1/venues/{venue_id}/reservations", listReservations).Methods("GET")
	r.HandleFunc("/api/v1/venues/{venue_id}/reservations/{reservation_id}", cancelReservation).Methods("DELETE")

	// Add some delay to simulate real API latency
	handler := addLatency(r)

	port := os.Getenv("MOCK_API_PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("🎭 Mock Partner API starting on http://localhost:%s\n", port)
	fmt.Println("Available endpoints:")
	fmt.Printf("  POST   /api/v1/venues/{venue_id}/reservations\n")
	fmt.Printf("  GET    /api/v1/venues/{venue_id}/reservations\n")
	fmt.Printf("  DELETE /api/v1/venues/{venue_id}/reservations/{reservation_id}\n")

	http.ListenAndServe(":"+port, handler)
}

func addLatency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add random delay (100-500ms)
		time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)

		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func createReservation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID := vars["venue_id"]

	var req struct {
		Date  string `json:"date"`
		Start string `json:"start"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Simulate occasional failures (5% chance of 503, 5% chance of 409)
	randNum := rand.Intn(100)
	if randNum < 5 {
		respondWithError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
		return
	}
	if randNum < 10 {
		respondWithError(w, http.StatusConflict, "venue at capacity")
		return
	}

	resID := "res-" + uuid.New().String()[:8]

	res := &Reservation{
		ID:        resID,
		VenueID:   venueID,
		Date:      req.Date,
		Start:     req.Start,
		Status:    "confirmed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mu.Lock()
	reservations[resID] = res
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func listReservations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID := vars["venue_id"]

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	mu.RLock()
	var venueReservations []Reservation
	for _, res := range reservations {
		if res.VenueID == venueID {
			// Filter by date range if provided
			if from != "" && res.Date < from {
				continue
			}
			if to != "" && res.Date > to {
				continue
			}
			venueReservations = append(venueReservations, *res)
		}
	}
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"venue_id":      venueID,
		"reservations":  venueReservations,
	})
}

func cancelReservation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resID := vars["reservation_id"]

	mu.Lock()
	res, exists := reservations[resID]
	if !exists {
		mu.Unlock()
		respondWithError(w, http.StatusNotFound, "reservation not found")
		return
	}

	res.Status = "cancelled"
	res.UpdatedAt = time.Now()
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reservation_id": resID,
		"status":         "cancelled",
	})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   message,
		"code":    code,
		"message": message,
	})
}
