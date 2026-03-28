package sync

import (
	"breeze-backend/internal/db"
	"breeze-backend/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// Service handles synchronization with partner APIs
type Service struct {
	DB         *db.DB
	HTTPClient *http.Client
	BaseURL    string
}

// NewService creates a new sync service
func NewService(database *db.DB) *Service {
	baseURL := os.Getenv("PARTNER_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3001"
	}

	return &Service{
		DB: database,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL: baseURL,
	}
}

// SyncVenue performs a sync for a specific venue
func (s *Service) SyncVenue(venueID uuid.UUID) {
	// Create sync log entry
	logID, err := s.DB.CreateSyncLog(&venueID, "manual")
	if err != nil {
		fmt.Printf("Failed to create sync log: %v\n", err)
		return
	}

	recordsProcessed := 0

	// Get pending reservations for this venue
	pendingReservations, err := s.DB.GetPendingReservations()
	if err != nil {
		errMsg := err.Error()
		s.DB.UpdateSyncLog(logID, "failed", 0, &errMsg)
		return
	}

	// Process each pending reservation
	for _, res := range pendingReservations {
		// Only process reservations for this venue
		if res.VenueID != venueID {
			continue
		}

		if err := s.processReservation(&res); err != nil {
			fmt.Printf("Failed to process reservation %s: %v\n", res.ID, err)
			// Mark as failed but continue with others
			s.DB.UpdateReservationStatus(res.ID, "failed", nil)
		} else {
			recordsProcessed++
		}
	}

	// Fetch current reservations from partner API to check for cancellations
	if err := s.checkForCancellations(venueID); err != nil {
		fmt.Printf("Failed to check for cancellations: %v\n", err)
	}

	// Update sync log
	s.DB.UpdateSyncLog(logID, "success", recordsProcessed, nil)
}

// ProcessAllVenues syncs all venues (for background jobs)
func (s *Service) ProcessAllVenues() {
	venues, err := s.DB.ListVenues()
	if err != nil {
		fmt.Printf("Failed to list venues: %v\n", err)
		return
	}

	for _, venue := range venues {
		if venue.Status == "active" {
			s.SyncVenue(venue.ID)
		}
	}
}

// processReservation handles creating/updating a single reservation
func (s *Service) processReservation(res *models.Reservation) error {
	// Get the date details
	date, err := s.DB.GetDateByID(res.DateID)
	if err != nil {
		return fmt.Errorf("failed to get date: %w", err)
	}
	if date == nil {
		return fmt.Errorf("date not found: %s", res.DateID)
	}

	// Build the API request
	reqBody := map[string]interface{}{
		"date":  date.Date.Format("2006-01-02"),
		"start": date.StartTime.Format("15:04"),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/venues/%s/reservations", s.BaseURL, res.VenueID.String())

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request with retries
	var resp *http.Response
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = s.HTTPClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to call partner API after retries: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle different response codes
	switch resp.StatusCode {
	case 201:
		// Success - parse reservation ID
		var result struct {
			ReservationID string `json:"reservation_id"`
			Status        string `json:"status"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("failed to parse success response: %w", err)
		}

		// Update reservation with external ID and confirmed status
		if err := s.DB.UpdateReservationStatus(res.ID, "confirmed", &result.ReservationID); err != nil {
			return fmt.Errorf("failed to update reservation: %w", err)
		}

		// Update date status
		if err := s.DB.UpdateDateStatus(res.DateID, "confirmed"); err != nil {
			return fmt.Errorf("failed to update date status: %w", err)
		}

		return nil

	case 409:
		// Conflict - venue is at capacity
		return fmt.Errorf("venue at capacity")

	case 503:
		// Service unavailable - retry later
		return fmt.Errorf("partner API unavailable")

	default:
		return fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode, string(body))
	}
}

// checkForCancellations fetches reservations from partner and updates local state
func (s *Service) checkForCancellations(venueID uuid.UUID) error {
	// Calculate date range (past week to future)
	from := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	to := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/venues/%s/reservations?from=%s&to=%s", s.BaseURL, venueID.String(), from, to)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch reservations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		VenueID       string `json:"venue_id"`
		Reservations  []struct {
			ReservationID string `json:"reservation_id"`
			Date          string `json:"date"`
			Start         string `json:"start"`
			Status        string `json:"status"`
			UpdatedAt     string `json:"updated_at"`
		} `json:"reservations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Build map of external reservations
	externalReservations := make(map[string]string)
	for _, r := range result.Reservations {
		externalReservations[r.ReservationID] = r.Status
	}

	// Check our local reservations for any that are cancelled externally
	localReservations, err := s.DB.GetPendingReservations()
	if err != nil {
		return fmt.Errorf("failed to get local reservations: %w", err)
	}

	for _, localRes := range localReservations {
		if localRes.ExternalReservationID == nil {
			continue
		}

		externalStatus, exists := externalReservations[*localRes.ExternalReservationID]
		if exists && externalStatus == "cancelled" && localRes.Status != "cancelled" {
			// Venue cancelled our reservation - mark it and reschedule
			fmt.Printf("Reservation %s was cancelled by venue\n", *localRes.ExternalReservationID)
			
			if err := s.DB.UpdateReservationStatus(localRes.ID, "cancelled", nil); err != nil {
				fmt.Printf("Failed to update reservation status: %v\n", err)
				continue
			}

			if err := s.DB.UpdateDateStatus(localRes.DateID, "rescheduling"); err != nil {
				fmt.Printf("Failed to update date status: %v\n", err)
				continue
			}

			// TODO: Implement auto-reschedule to different venue
			// For now, we'll just mark it for manual intervention
		}
	}

	return nil
}

// CancelReservation cancels a reservation with the partner API
func (s *Service) CancelReservation(venueID uuid.UUID, reservationID string) error {
	url := fmt.Sprintf("%s/api/v1/venues/%s/reservations/%s", s.BaseURL, venueID.String(), reservationID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
