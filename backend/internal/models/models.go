package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Venue represents a partner venue
type Venue struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	Name               string    `json:"name" db:"name"`
	Address            string    `json:"address" db:"address"`
	City               string    `json:"city" db:"city"`
	Timezone           string    `json:"timezone" db:"timezone"`
	Status             string    `json:"status" db:"status"`
	PartnerAPIEndpoint *string   `json:"partner_api_endpoint" db:"partner_api_endpoint"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// AvailabilitySlot represents a recurring weekly availability slot
type AvailabilitySlot struct {
	ID         uuid.UUID `json:"id" db:"id"`
	VenueID    uuid.UUID `json:"venue_id" db:"venue_id"`
	DayOfWeek  int       `json:"day_of_week" db:"day_of_week"`
	StartTime  TimeOnly  `json:"start_time" db:"start_time"`
	EndTime    TimeOnly  `json:"end_time" db:"end_time"`
	Capacity   int       `json:"capacity" db:"capacity"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Date represents a scheduled date between matched users
type Date struct {
	ID           uuid.UUID `json:"id" db:"id"`
	VenueID      uuid.UUID `json:"venue_id" db:"venue_id"`
	UserPairID   string    `json:"user_pair_id" db:"user_pair_id"`
	Date         DateOnly  `json:"date" db:"date"`
	StartTime    TimeOnly  `json:"start_time" db:"start_time"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Reservation tracks the external reservation state
type Reservation struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	DateID                 uuid.UUID  `json:"date_id" db:"date_id"`
	VenueID                uuid.UUID  `json:"venue_id" db:"venue_id"`
	ExternalReservationID  *string    `json:"external_reservation_id" db:"external_reservation_id"`
	Status                 string     `json:"status" db:"status"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
	LastSyncedAt           *time.Time `json:"last_synced_at" db:"last_synced_at"`
}

// SyncLog tracks sync operations
type SyncLog struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	VenueID           *uuid.UUID `json:"venue_id" db:"venue_id"`
	SyncType          string     `json:"sync_type" db:"sync_type"`
	Status            string     `json:"status" db:"status"`
	StartedAt         time.Time  `json:"started_at" db:"started_at"`
	CompletedAt       *time.Time `json:"completed_at" db:"completed_at"`
	RecordsProcessed  int        `json:"records_processed" db:"records_processed"`
	ErrorMessage      *string    `json:"error_message" db:"error_message"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// TimeOnly is a custom type for TIME columns
type TimeOnly struct {
	time.Time
}

// MarshalJSON implements json.Marshaler
func (t TimeOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format("15:04"))
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TimeOnly) UnmarshalJSON(data []byte) error {
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err != nil {
		return err
	}
	parsed, err := time.Parse("15:04", timeStr)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// Scan implements sql.Scanner
func (t *TimeOnly) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		t.Time = v
	case []byte:
		parsed, err := time.Parse("15:04:05", string(v))
		if err != nil {
			return err
		}
		t.Time = parsed
	case string:
		parsed, err := time.Parse("15:04:05", v)
		if err != nil {
			return err
		}
		t.Time = parsed
	default:
		return errors.New("cannot scan type into TimeOnly")
	}
	return nil
}

// Value implements driver.Valuer
func (t TimeOnly) Value() (driver.Value, error) {
	return t.Format("15:04:05"), nil
}

// DateOnly is a custom type for DATE columns
type DateOnly struct {
	time.Time
}

// MarshalJSON implements json.Marshaler
func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("2006-01-02"))
}

// UnmarshalJSON implements json.Unmarshaler
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var dateStr string
	if err := json.Unmarshal(data, &dateStr); err != nil {
		return err
	}
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}

// Scan implements sql.Scanner
func (d *DateOnly) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case []byte:
		parsed, err := time.Parse("2006-01-02", string(v))
		if err != nil {
			return err
		}
		d.Time = parsed
	case string:
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		d.Time = parsed
	default:
		return errors.New("cannot scan type into DateOnly")
	}
	return nil
}

// Value implements driver.Valuer
func (d DateOnly) Value() (driver.Value, error) {
	return d.Format("2006-01-02"), nil
}

// SlotAvailability represents a computed availability slot with current bookings
type SlotAvailability struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Capacity  int    `json:"capacity"`
	Booked    int    `json:"booked"`
	Available int    `json:"available"`
}

// DateWithReservation combines Date and Reservation info
type DateWithReservation struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	UserPairID            string     `json:"user_pair_id" db:"user_pair_id"`
	Date                  DateOnly   `json:"date" db:"date"`
	StartTime             TimeOnly   `json:"start_time" db:"start_time"`
	Status                string     `json:"status" db:"status"`
	ExternalReservationID *string    `json:"external_reservation_id" db:"external_reservation_id"`
	ReservationStatus     string     `json:"reservation_status" db:"reservation_status"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
}
