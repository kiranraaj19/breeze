package handlers

import (
	"breeze-backend/internal/db"
	"breeze-backend/internal/models"
	"breeze-backend/internal/sync"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	DB   *db.DB
	Sync *sync.Service
}

// NewHandler creates a new handler instance
func NewHandler(database *db.DB, syncService *sync.Service) *Handler {
	return &Handler{
		DB:   database,
		Sync: syncService,
	}
}

// GetVenues returns all venues
func (h *Handler) GetVenues(w http.ResponseWriter, r *http.Request) {
	venues, err := h.DB.ListVenues()
	if err != nil {
		fmt.Printf("Error listing venues: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch venues: "+err.Error())
		return
	}
	if venues == nil {
		venues = []models.Venue{}
	}
	respondWithJSON(w, http.StatusOK, venues)
}

// GetVenue returns a single venue by ID
func (h *Handler) GetVenue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := uuid.Parse(vars["venue_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid venue ID")
		return
	}

	venue, err := h.DB.GetVenue(venueID)
	if err != nil {
		fmt.Printf("Error getting venue %s: %v\n", venueID, err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if venue == nil {
		respondWithError(w, http.StatusNotFound, "venue not found")
		return
	}

	respondWithJSON(w, http.StatusOK, venue)
}

// GetAvailability returns availability slots for a venue
func (h *Handler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := uuid.Parse(vars["venue_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid venue ID")
		return
	}

	// Parse date range from query params
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" {
		fromStr = time.Now().Format("2006-01-02")
	}
	if toStr == "" {
		toStr = time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid from date")
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid to date")
		return
	}

	slots, err := h.DB.GetAvailabilityForDateRange(venueID, from, to)
	if err != nil {
		fmt.Printf("Error getting availability for venue %s: %v\n", venueID, err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"venue_id": venueID.String(),
		"from":     fromStr,
		"to":       toStr,
		"slots":    slots,
	})
}

// GetDates returns dates for a venue
func (h *Handler) GetDates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := uuid.Parse(vars["venue_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid venue ID")
		return
	}

	dates, err := h.DB.GetDatesByVenue(venueID)
	if err != nil {
		fmt.Printf("Error getting dates for venue %s: %v\n", venueID, err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"venue_id": venueID.String(),
		"dates":    dates,
	})
}

// GetSyncStatus returns the last sync status for a venue
func (h *Handler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := uuid.Parse(vars["venue_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid venue ID")
		return
	}

	syncLog, err := h.DB.GetLastSyncStatus(venueID)
	if err != nil {
		fmt.Printf("Error getting sync status for venue %s: %v\n", venueID, err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"venue_id":  venueID.String(),
		"last_sync": syncLog,
	})
}

// TriggerSync triggers a manual sync for a venue
func (h *Handler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueID, err := uuid.Parse(vars["venue_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid venue ID")
		return
	}

	// Start sync in background
	go h.Sync.SyncVenue(venueID)

	respondWithJSON(w, http.StatusAccepted, map[string]interface{}{
		"venue_id": venueID.String(),
		"status":   "sync_started",
	})
}

// CreateDateRequest represents a request to create a new date
type CreateDateRequest struct {
	VenueID    string `json:"venue_id"`
	UserPairID string `json:"user_pair_id"`
	Date       string `json:"date"`
	StartTime  string `json:"start_time"`
}

// CreateDate creates a new date booking
func (h *Handler) CreateDate(w http.ResponseWriter, r *http.Request) {
	var req CreateDateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	venueID, err := uuid.Parse(req.VenueID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid venue ID")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid date format")
		return
	}

	newDate, err := h.DB.CreateDate(venueID, req.UserPairID, date, req.StartTime)
	if err != nil {
		fmt.Printf("Error creating date: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create reservation record
	_, err = h.DB.CreateReservation(newDate.ID, venueID, nil)
	if err != nil {
		fmt.Printf("Error creating reservation: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, newDate)
}

// HealthCheck returns health status
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	fmt.Printf("API Error %d: %s\n", code, message)
	respondWithJSON(w, code, map[string]interface{}{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// CORS middleware
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
