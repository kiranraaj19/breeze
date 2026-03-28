package handlers

import (
	"breeze-backend/internal/db"
	"breeze-backend/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// DateHandler handles date-related requests
type DateHandler struct {
	DB *db.DB
}

// NewDateHandler creates a new date handler
func NewDateHandler(database *db.DB) *DateHandler {
	return &DateHandler{DB: database}
}

// CancelDateRequest represents a cancel request
type CancelDateRequest struct {
	DateID string `json:"date_id"`
}

// CancelDate cancels a user's date
func (h *DateHandler) CancelDate(w http.ResponseWriter, r *http.Request) {
	userPairID := r.Context().Value("user_pair_id")
	if userPairID == nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	dateID, err := uuid.Parse(vars["date_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date ID")
		return
	}

	if err := h.DB.CancelDate(dateID, userPairID.(string)); err != nil {
		fmt.Printf("Error cancelling date: %v\n", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Date cancelled successfully",
		"date_id": dateID.String(),
		"status":  "cancelled",
	})
}

// RescheduleDateRequest represents a reschedule request
type RescheduleDateRequest struct {
	NewDate     string `json:"new_date"`
	NewStartTime string `json:"new_start_time"`
}

// RescheduleDate reschedules a date to a new time
func (h *DateHandler) RescheduleDate(w http.ResponseWriter, r *http.Request) {
	userPairID := r.Context().Value("user_pair_id")
	if userPairID == nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	dateID, err := uuid.Parse(vars["date_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date ID")
		return
	}

	var req RescheduleDateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	newDate, err := time.Parse("2006-01-02", req.NewDate)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format")
		return
	}

	updatedDate, err := h.DB.RescheduleDate(dateID, userPairID.(string), newDate, req.NewStartTime)
	if err != nil {
		fmt.Printf("Error rescheduling date: %v\n", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Date rescheduled successfully",
		"date":    updatedDate,
	})
}

// GetRescheduleOptions gets available slots for rescheduling
func (h *DateHandler) GetRescheduleOptions(w http.ResponseWriter, r *http.Request) {
	userPairID := r.Context().Value("user_pair_id")
	if userPairID == nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	dateID, err := uuid.Parse(vars["date_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date ID")
		return
	}

	slots, err := h.DB.GetAvailableSlotsForReschedule(dateID, userPairID.(string))
	if err != nil {
		fmt.Printf("Error getting reschedule options: %v\n", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"slots": slots,
	})
}

// SwitchVenueRequest represents a venue switch request
type SwitchVenueRequest struct {
	NewVenueID string `json:"new_venue_id"`
}

// SwitchVenue switches a date to a different venue
func (h *DateHandler) SwitchVenue(w http.ResponseWriter, r *http.Request) {
	userPairID := r.Context().Value("user_pair_id")
	if userPairID == nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	dateID, err := uuid.Parse(vars["date_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date ID")
		return
	}

	var req SwitchVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	newVenueID, err := uuid.Parse(req.NewVenueID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	updatedDate, err := h.DB.SwitchVenue(dateID, userPairID.(string), newVenueID)
	if err != nil {
		fmt.Printf("Error switching venue: %v\n", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get venue info for response
	venue, _ := h.DB.GetVenue(newVenueID)
	venueName := ""
	if venue != nil {
		venueName = venue.Name
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Venue switched successfully",
		"date":       updatedDate,
		"venue_id":   newVenueID.String(),
		"venue_name": venueName,
	})
}

// GetVenuesForSwitch gets venues available for switching with capacity check
func (h *DateHandler) GetVenuesForSwitch(w http.ResponseWriter, r *http.Request) {
	userPairID := r.Context().Value("user_pair_id")
	if userPairID == nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	dateID, err := uuid.Parse(vars["date_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date ID")
		return
	}

	// Get the date info
	date, err := h.DB.GetDateByID(dateID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get date")
		return
	}
	if date == nil {
		respondWithError(w, http.StatusNotFound, "Date not found")
		return
	}

	if date.UserPairID != userPairID.(string) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get all active venues
	allVenues, err := h.DB.ListVenues()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get venues")
		return
	}

	// Check capacity at each venue for the same date/time
	type VenueWithCapacity struct {
		models.Venue
		Capacity  int  `json:"capacity"`
		Available int  `json:"available"`
		CanSwitch bool `json:"can_switch"`
	}

	var availableVenues []VenueWithCapacity
	for _, venue := range allVenues {
		if venue.ID == date.VenueID {
			continue // Skip current venue
		}
		if venue.Status != "active" {
			continue
		}

		capacity, available, err := h.DB.CheckSlotCapacity(venue.ID, date.Date.Time, date.StartTime.Format("15:04"))
		if err != nil {
			continue // Skip venues without this slot
		}

		availableVenues = append(availableVenues, VenueWithCapacity{
			Venue:     venue,
			Capacity:  capacity,
			Available: available,
			CanSwitch: available > 0,
		})
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"current_venue_id": date.VenueID,
		"date":             date.Date,
		"start_time":       date.StartTime,
		"venues":           availableVenues,
	})
}
