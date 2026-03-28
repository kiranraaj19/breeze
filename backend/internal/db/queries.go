package db

import (
	"breeze-backend/internal/models"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GetVenue retrieves a venue by ID
func (db *DB) GetVenue(id uuid.UUID) (*models.Venue, error) {
	var venue models.Venue
	query := `SELECT id, name, address, city, timezone, status, partner_api_endpoint, created_at, updated_at 
	          FROM venues WHERE id = $1`
	err := db.Conn.QueryRow(query, id).Scan(
		&venue.ID, &venue.Name, &venue.Address, &venue.City,
		&venue.Timezone, &venue.Status, &venue.PartnerAPIEndpoint,
		&venue.CreatedAt, &venue.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &venue, nil
}

// ListVenues retrieves all venues
func (db *DB) ListVenues() ([]models.Venue, error) {
	query := `SELECT id, name, address, city, timezone, status, partner_api_endpoint, created_at, updated_at 
	          FROM venues ORDER BY name`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var venues []models.Venue
	for rows.Next() {
		var v models.Venue
		err := rows.Scan(
			&v.ID, &v.Name, &v.Address, &v.City,
			&v.Timezone, &v.Status, &v.PartnerAPIEndpoint,
			&v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		venues = append(venues, v)
	}
	return venues, nil
}

// GetAvailabilityForDateRange returns availability slots with booking counts for a venue
func (db *DB) GetAvailabilityForDateRange(venueID uuid.UUID, from, to time.Time) ([]models.SlotAvailability, error) {
	query := `
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
		ORDER BY a.date, a.start_time
	`

	rows, err := db.Conn.Query(query, venueID, from, to)
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

// GetDatesByVenue returns all dates for a venue
func (db *DB) GetDatesByVenue(venueID uuid.UUID) ([]models.DateWithReservation, error) {
	query := `
		SELECT 
			d.id, d.user_pair_id, d.date, d.start_time, d.status,
			r.external_reservation_id, COALESCE(r.status, 'none') as reservation_status,
			d.created_at
		FROM dates d
		LEFT JOIN reservations r ON r.date_id = d.id
		WHERE d.venue_id = $1
		ORDER BY d.date DESC, d.start_time DESC
	`

	rows, err := db.Conn.Query(query, venueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []models.DateWithReservation
	for rows.Next() {
		var d models.DateWithReservation
		err := rows.Scan(
			&d.ID, &d.UserPairID, &d.Date, &d.StartTime, &d.Status,
			&d.ExternalReservationID, &d.ReservationStatus, &d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		dates = append(dates, d)
	}
	return dates, nil
}

// GetLastSyncStatus returns the most recent sync status for a venue
func (db *DB) GetLastSyncStatus(venueID uuid.UUID) (*models.SyncLog, error) {
	var log models.SyncLog
	query := `
		SELECT id, venue_id, sync_type, status, started_at, completed_at, records_processed, error_message, created_at
		FROM sync_logs
		WHERE venue_id = $1
		ORDER BY started_at DESC
		LIMIT 1
	`
	err := db.Conn.QueryRow(query, venueID).Scan(
		&log.ID, &log.VenueID, &log.SyncType, &log.Status,
		&log.StartedAt, &log.CompletedAt, &log.RecordsProcessed,
		&log.ErrorMessage, &log.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// CreateSyncLog creates a new sync log entry
func (db *DB) CreateSyncLog(venueID *uuid.UUID, syncType string) (uuid.UUID, error) {
	var id uuid.UUID
	query := `
		INSERT INTO sync_logs (venue_id, sync_type, status, started_at)
		VALUES ($1, $2, 'running', NOW())
		RETURNING id
	`
	err := db.Conn.QueryRow(query, venueID, syncType).Scan(&id)
	return id, err
}

// UpdateSyncLog updates a sync log entry with completion status
func (db *DB) UpdateSyncLog(id uuid.UUID, status string, records int, errMsg *string) error {
	query := `
		UPDATE sync_logs 
		SET status = $2, completed_at = NOW(), records_processed = $3, error_message = $4
		WHERE id = $1
	`
	_, err := db.Conn.Exec(query, id, status, records, errMsg)
	return err
}

// CreateDate creates a new date booking
func (db *DB) CreateDate(venueID uuid.UUID, userPairID string, date time.Time, startTime string) (*models.Date, error) {
	// Parse start time
	parsedTime, err := time.Parse("15:04", startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}

	var newDate models.Date
	query := `
		INSERT INTO dates (venue_id, user_pair_id, date, start_time, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, venue_id, user_pair_id, date, start_time, status, created_at, updated_at
	`
	err = db.Conn.QueryRow(query, venueID, userPairID, date, parsedTime).Scan(
		&newDate.ID, &newDate.VenueID, &newDate.UserPairID, &newDate.Date,
		&newDate.StartTime, &newDate.Status, &newDate.CreatedAt, &newDate.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &newDate, nil
}

// CreateReservation creates a reservation record
func (db *DB) CreateReservation(dateID, venueID uuid.UUID, externalID *string) (*models.Reservation, error) {
	var res models.Reservation
	query := `
		INSERT INTO reservations (date_id, venue_id, external_reservation_id, status, last_synced_at)
		VALUES ($1, $2, $3, 'pending', NOW())
		RETURNING id, date_id, venue_id, external_reservation_id, status, created_at, updated_at, last_synced_at
	`
	err := db.Conn.QueryRow(query, dateID, venueID, externalID).Scan(
		&res.ID, &res.DateID, &res.VenueID, &res.ExternalReservationID,
		&res.Status, &res.CreatedAt, &res.UpdatedAt, &res.LastSyncedAt,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// UpdateReservationStatus updates a reservation status
func (db *DB) UpdateReservationStatus(id uuid.UUID, status string, externalID *string) error {
	query := `
		UPDATE reservations 
		SET status = $2, external_reservation_id = COALESCE($3, external_reservation_id), last_synced_at = NOW()
		WHERE id = $1
	`
	_, err := db.Conn.Exec(query, id, status, externalID)
	return err
}

// GetPendingReservations returns all pending reservations
func (db *DB) GetPendingReservations() ([]models.Reservation, error) {
	query := `
		SELECT id, date_id, venue_id, external_reservation_id, status, created_at, updated_at, last_synced_at
		FROM reservations
		WHERE status IN ('pending', 'failed')
		ORDER BY created_at ASC
	`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var r models.Reservation
		err := rows.Scan(
			&r.ID, &r.DateID, &r.VenueID, &r.ExternalReservationID,
			&r.Status, &r.CreatedAt, &r.UpdatedAt, &r.LastSyncedAt,
		)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, r)
	}
	return reservations, nil
}

// GetReservationByDateID retrieves a reservation by date ID
func (db *DB) GetReservationByDateID(dateID uuid.UUID) (*models.Reservation, error) {
	var r models.Reservation
	query := `
		SELECT id, date_id, venue_id, external_reservation_id, status, created_at, updated_at, last_synced_at
		FROM reservations
		WHERE date_id = $1
	`
	err := db.Conn.QueryRow(query, dateID).Scan(
		&r.ID, &r.DateID, &r.VenueID, &r.ExternalReservationID,
		&r.Status, &r.CreatedAt, &r.UpdatedAt, &r.LastSyncedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateDateStatus updates a date status
func (db *DB) UpdateDateStatus(id uuid.UUID, status string) error {
	query := `UPDATE dates SET status = $2 WHERE id = $1`
	_, err := db.Conn.Exec(query, id, status)
	return err
}

// GetDateByID retrieves a date by ID
func (db *DB) GetDateByID(id uuid.UUID) (*models.Date, error) {
	var d models.Date
	query := `
		SELECT id, venue_id, user_pair_id, date, start_time, status, created_at, updated_at
		FROM dates WHERE id = $1
	`
	err := db.Conn.QueryRow(query, id).Scan(
		&d.ID, &d.VenueID, &d.UserPairID, &d.Date,
		&d.StartTime, &d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}
