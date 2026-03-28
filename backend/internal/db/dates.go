package db

import (
	"breeze-backend/internal/models"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CancelDate cancels a date and its reservation
func (db *DB) CancelDate(dateID uuid.UUID, userPairID string) error {
	// Verify the date belongs to the user
	var date models.Date
	query := `SELECT id, venue_id, user_pair_id, status FROM dates WHERE id = $1`
	err := db.Conn.QueryRow(query, dateID).Scan(&date.ID, &date.VenueID, &date.UserPairID, &date.Status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("date not found")
	}
	if err != nil {
		return err
	}

	if date.UserPairID != userPairID {
		return fmt.Errorf("unauthorized: date does not belong to user")
	}

	if date.Status == "cancelled" {
		return fmt.Errorf("date is already cancelled")
	}

	// Start transaction
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update date status
	_, err = tx.Exec(`UPDATE dates SET status = 'cancelled', updated_at = NOW() WHERE id = $1`, dateID)
	if err != nil {
		return err
	}

	// Update reservation status
	_, err = tx.Exec(`
		UPDATE reservations 
		SET status = 'cancelled', updated_at = NOW(), last_synced_at = NOW()
		WHERE date_id = $1
	`, dateID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RescheduleDate reschedules a date to a new date/time at the same venue
func (db *DB) RescheduleDate(dateID uuid.UUID, userPairID string, newDate time.Time, newStartTime string) (*models.Date, error) {
	// Verify the date belongs to the user
	var date models.Date
	query := `SELECT id, venue_id, user_pair_id, status FROM dates WHERE id = $1`
	err := db.Conn.QueryRow(query, dateID).Scan(&date.ID, &date.VenueID, &date.UserPairID, &date.Status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("date not found")
	}
	if err != nil {
		return nil, err
	}

	if date.UserPairID != userPairID {
		return nil, fmt.Errorf("unauthorized: date does not belong to user")
	}

	if date.Status == "cancelled" {
		return nil, fmt.Errorf("cannot reschedule cancelled date")
	}

	// Parse start time
	parsedTime, err := time.Parse("15:04", newStartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}

	// Check capacity at new slot
	dayOfWeek := int(newDate.Weekday())
	var capacity int
	err = db.Conn.QueryRow(`
		SELECT capacity FROM availability_slots 
		WHERE venue_id = $1 AND day_of_week = $2 AND start_time = $3 AND is_active = true
	`, date.VenueID, dayOfWeek, parsedTime).Scan(&capacity)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no availability slot found for this date/time")
	}
	if err != nil {
		return nil, err
	}

	// Count existing bookings at this slot
	var booked int
	err = db.Conn.QueryRow(`
		SELECT COUNT(*) FROM dates 
		WHERE venue_id = $1 AND date = $2 AND start_time = $3 AND status NOT IN ('cancelled')
	`, date.VenueID, newDate, parsedTime).Scan(&booked)
	if err != nil {
		return nil, err
	}

	if booked >= capacity {
		return nil, fmt.Errorf("no capacity available at this time slot")
	}

	// Start transaction
	tx, err := db.Conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update date
	_, err = tx.Exec(`
		UPDATE dates 
		SET date = $2, start_time = $3, status = 'rescheduling', updated_at = NOW()
		WHERE id = $1
	`, dateID, newDate, parsedTime)
	if err != nil {
		return nil, err
	}

	// Update reservation to pending (needs re-sync)
	_, err = tx.Exec(`
		UPDATE reservations 
		SET status = 'pending', updated_at = NOW()
		WHERE date_id = $1
	`, dateID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return updated date
	return db.GetDateByID(dateID)
}

// CheckSlotCapacity checks if a venue has capacity at a specific date/time
func (db *DB) CheckSlotCapacity(venueID uuid.UUID, date time.Time, startTime string) (capacity int, available int, err error) {
	parsedTime, err := time.Parse("15:04", startTime)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start time: %w", err)
	}

	dayOfWeek := int(date.Weekday())
	
	// Get slot capacity
	err = db.Conn.QueryRow(`
		SELECT capacity FROM availability_slots 
		WHERE venue_id = $1 AND day_of_week = $2 AND start_time = $3 AND is_active = true
	`, venueID, dayOfWeek, parsedTime).Scan(&capacity)
	if err == sql.ErrNoRows {
		return 0, 0, fmt.Errorf("no availability slot found")
	}
	if err != nil {
		return 0, 0, err
	}

	// Count existing bookings
	var booked int
	err = db.Conn.QueryRow(`
		SELECT COUNT(*) FROM dates 
		WHERE venue_id = $1 AND date = $2 AND start_time = $3 AND status NOT IN ('cancelled')
	`, venueID, date, parsedTime).Scan(&booked)
	if err != nil {
		return 0, 0, err
	}

	available = capacity - booked
	if available < 0 {
		available = 0
	}

	return capacity, available, nil
}

// SwitchVenue switches a date to a different venue with capacity check
func (db *DB) SwitchVenue(dateID uuid.UUID, userPairID string, newVenueID uuid.UUID) (*models.Date, error) {
	// Verify the date belongs to the user
	var date models.Date
	query := `SELECT id, venue_id, user_pair_id, date, start_time, status FROM dates WHERE id = $1`
	err := db.Conn.QueryRow(query, dateID).Scan(&date.ID, &date.VenueID, &date.UserPairID, &date.Date, &date.StartTime, &date.Status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("date not found")
	}
	if err != nil {
		return nil, err
	}

	if date.UserPairID != userPairID {
		return nil, fmt.Errorf("unauthorized: date does not belong to user")
	}

	if date.Status == "cancelled" {
		return nil, fmt.Errorf("cannot switch cancelled date")
	}

	// Check capacity at new venue
	capacity, available, err := db.CheckSlotCapacity(newVenueID, date.Date.Time, date.StartTime.Format("15:04"))
	if err != nil {
		return nil, err
	}
	if available <= 0 {
		return nil, fmt.Errorf("new venue has no capacity at this time slot (capacity: %d)", capacity)
	}

	// Start transaction
	tx, err := db.Conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get existing reservation
	var resID uuid.UUID
	var extResID *string
	err = tx.QueryRow(`SELECT id, external_reservation_id FROM reservations WHERE date_id = $1`, dateID).Scan(&resID, &extResID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Mark old reservation as cancelled
	if err != sql.ErrNoRows {
		_, err = tx.Exec(`
			UPDATE reservations 
			SET status = 'cancelled', updated_at = NOW(), last_synced_at = NOW()
			WHERE id = $1
		`, resID)
		if err != nil {
			return nil, err
		}
	}

	// Update date with new venue
	_, err = tx.Exec(`
		UPDATE dates 
		SET venue_id = $2, status = 'rescheduling', updated_at = NOW()
		WHERE id = $1
	`, dateID, newVenueID)
	if err != nil {
		return nil, err
	}

	// Create new reservation
	_, err = tx.Exec(`
		INSERT INTO reservations (date_id, venue_id, external_reservation_id, status, created_at, updated_at)
		VALUES ($1, $2, NULL, 'pending', NOW(), NOW())
		ON CONFLICT (date_id) DO UPDATE SET
			venue_id = $2,
			external_reservation_id = NULL,
			status = 'pending',
			updated_at = NOW()
	`, dateID, newVenueID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return updated date
	return db.GetDateByID(dateID)
}

// GetAvailableSlotsForReschedule gets available slots for rescheduling
func (db *DB) GetAvailableSlotsForReschedule(dateID uuid.UUID, userPairID string) ([]models.SlotAvailability, error) {
	// Get current date info
	var date models.Date
	query := `SELECT id, venue_id, user_pair_id, date, start_time FROM dates WHERE id = $1`
	err := db.Conn.QueryRow(query, dateID).Scan(&date.ID, &date.VenueID, &date.UserPairID, &date.Date, &date.StartTime)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("date not found")
	}
	if err != nil {
		return nil, err
	}

	if date.UserPairID != userPairID {
		return nil, fmt.Errorf("unauthorized: date does not belong to user")
	}

	// Get availability for next 30 days
	from := time.Now()
	to := from.AddDate(0, 0, 30)

	query = `
		WITH date_range AS (
			SELECT generate_series($2::date, $3::date, '1 day'::interval)::date AS date
		),
		available_slots AS (
			SELECT 
				d.date,
				s.start_time,
				s.end_time,
				s.capacity
			FROM date_range d
			JOIN availability_slots s ON s.day_of_week = EXTRACT(DOW FROM d.date)
			WHERE s.venue_id = $1 AND s.is_active = true
		),
		bookings AS (
			SELECT 
				d.date,
				d.start_time,
				COUNT(*) as booked
			FROM dates d
			WHERE d.venue_id = $1 
			  AND d.date BETWEEN $2 AND $3
			  AND d.status NOT IN ('cancelled')
			  AND d.id != $4
			GROUP BY d.date, d.start_time
		)
		SELECT 
			a.date::text,
			to_char(a.start_time, 'HH24:MI') as start_time,
			to_char(a.end_time, 'HH24:MI') as end_time,
			a.capacity,
			COALESCE(b.booked, 0) as booked,
			a.capacity - COALESCE(b.booked, 0) as available
		FROM available_slots a
		LEFT JOIN bookings b ON a.date = b.date AND a.start_time = b.start_time
		WHERE a.capacity - COALESCE(b.booked, 0) > 0
		ORDER BY a.date, a.start_time
		LIMIT 50
	`

	rows, err := db.Conn.Query(query, date.VenueID, from, to, dateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []models.SlotAvailability
	for rows.Next() {
		var s models.SlotAvailability
		err := rows.Scan(&s.Date, &s.StartTime, &s.EndTime, &s.Capacity, &s.Booked, &s.Available)
		if err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, nil
}
