package db

import (
	"breeze-backend/internal/models"
	"database/sql"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with hashed password
func (db *DB) CreateUser(email, password, fullName string) (*models.User, error) {
	// Generate password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Generate user_pair_id from email
	userPairID := "pair-" + email[:len(email)-len("@test.com")]

	var user models.User
	query := `
		INSERT INTO users (email, password_hash, user_pair_id, full_name, status)
		VALUES ($1, $2, $3, $4, 'active')
		RETURNING id, email, user_pair_id, full_name, status, created_at, updated_at
	`
	err = db.Conn.QueryRow(query, email, string(hash), userPairID, fullName).Scan(
		&user.ID, &user.Email, &user.UserPairID, &user.FullName, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email (including password hash)
func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password_hash, user_pair_id, full_name, status, created_at, updated_at
		FROM users WHERE email = $1
	`
	err := db.Conn.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.UserPairID, &user.FullName,
		&user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID (without password hash)
func (db *DB) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, user_pair_id, full_name, status, created_at, updated_at
		FROM users WHERE id = $1
	`
	err := db.Conn.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.UserPairID, &user.FullName,
		&user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// VerifyPassword checks if the provided password matches the stored hash
func VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetDatesByUserPairID retrieves all dates for a specific user pair
func (db *DB) GetDatesByUserPairID(userPairID string) ([]models.UserDate, error) {
	query := `
		SELECT 
			d.id, d.venue_id, v.name as venue_name, v.address as venue_address,
			d.user_pair_id, d.date, d.start_time, d.status,
			r.external_reservation_id, COALESCE(r.status, 'none') as reservation_status,
			d.created_at
		FROM dates d
		JOIN venues v ON v.id = d.venue_id
		LEFT JOIN reservations r ON r.date_id = d.id
		WHERE d.user_pair_id = $1
		ORDER BY d.date DESC, d.start_time DESC
	`

	rows, err := db.Conn.Query(query, userPairID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []models.UserDate
	for rows.Next() {
		var d models.UserDate
		err := rows.Scan(
			&d.ID, &d.VenueID, &d.VenueName, &d.VenueAddress,
			&d.UserPairID, &d.Date, &d.StartTime, &d.Status,
			&d.ExternalReservationID, &d.ReservationStatus, &d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		dates = append(dates, d)
	}
	return dates, nil
}

// GetUserByUserPairID retrieves a user by their user_pair_id
func (db *DB) GetUserByUserPairID(userPairID string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, user_pair_id, full_name, status, created_at, updated_at
		FROM users WHERE user_pair_id = $1
	`
	err := db.Conn.QueryRow(query, userPairID).Scan(
		&user.ID, &user.Email, &user.UserPairID, &user.FullName,
		&user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
