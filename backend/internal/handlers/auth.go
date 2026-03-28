package handlers

import (
	"breeze-backend/internal/db"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	DB *db.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(database *db.DB) *AuthHandler {
	return &AuthHandler{DB: database}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the response after successful auth
type AuthResponse struct {
	Token      string `json:"token"`
	ID         string `json:"id"`
	Email      string `json:"email"`
	UserPairID string `json:"user_pair_id"`
	FullName   string `json:"full_name,omitempty"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if len(req.Password) < 4 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 4 characters")
		return
	}

	// Check if user already exists
	existing, err := h.DB.GetUserByEmail(req.Email)
	if err != nil {
		fmt.Printf("Error checking existing user: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existing != nil {
		respondWithError(w, http.StatusConflict, "User with this email already exists")
		return
	}

	// Create user
	user, err := h.DB.CreateUser(req.Email, req.Password, req.FullName)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate simple token (in production, use JWT)
	token := generateToken(user.ID)

	respondWithJSON(w, http.StatusCreated, AuthResponse{
		Token:      token,
		ID:         user.ID.String(),
		Email:      user.Email,
		UserPairID: user.UserPairID,
		FullName:   "",
	})
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Get user by email
	user, err := h.DB.GetUserByEmail(req.Email)
	if err != nil {
		fmt.Printf("Error getting user: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if user == nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Verify password
	if !db.VerifyPassword(user.PasswordHash, req.Password) {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Check if user is active
	if user.Status != "active" {
		respondWithError(w, http.StatusForbidden, "Account is inactive")
		return
	}

	// Generate simple token (in production, use JWT)
	token := generateToken(user.ID)

	fullName := ""
	if user.FullName != nil {
		fullName = *user.FullName
	}

	respondWithJSON(w, http.StatusOK, AuthResponse{
		Token:      token,
		ID:         user.ID.String(),
		Email:      user.Email,
		UserPairID: user.UserPairID,
		FullName:   fullName,
	})
}

// GetMe returns the current user's profile
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := uuid.Parse(userID.(string))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	user, err := h.DB.GetUserByID(id)
	if err != nil {
		fmt.Printf("Error getting user: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if user == nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// GetMyDates returns all dates for the logged-in user
func (h *AuthHandler) GetMyDates(w http.ResponseWriter, r *http.Request) {
	userPairID := r.Context().Value("user_pair_id")
	if userPairID == nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dates, err := h.DB.GetDatesByUserPairID(userPairID.(string))
	if err != nil {
		fmt.Printf("Error getting user dates: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch dates")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"dates": dates,
	})
}

// generateToken creates a simple token (in production, use JWT)
func generateToken(userID uuid.UUID) string {
	// Simple token format: userID:timestamp
	return fmt.Sprintf("%s:%d", userID.String(), time.Now().Unix())
}

// AuthMiddleware validates the auth token and adds user info to context
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]
		// Parse token (userID:timestamp)
		tokenParts := strings.Split(token, ":")
		if len(tokenParts) != 2 {
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		userID := tokenParts[0]

		// Validate user exists
		id, err := uuid.Parse(userID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		user, err := h.DB.GetUserByID(id)
		if err != nil || user == nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Add user info to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "user_pair_id", user.UserPairID)
		ctx = context.WithValue(ctx, "email", user.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


