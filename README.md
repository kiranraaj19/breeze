# Breeze - Partner Availability & Date Management System

A full-stack application for managing partner venue availability and reservations for the Breeze dating app.

## Project Structure

```
breeze/
├── .env                          # Common environment configuration
├── README.md                     # This file
├── backend/                      # Go backend API
│   ├── main.go                   # Entry point
│   ├── go.mod                    # Go dependencies
│   ├── internal/
│   │   ├── db/                   # Database connection & queries
│   │   ├── handlers/             # HTTP handlers
│   │   ├── models/               # Data models
│   │   └── sync/                 # Sync service
│   └── migrations/               # SQL migrations
│       ├── 001_initial_schema.sql
│       └── 002_seed_data.sql
├── frontend/                     # Next.js Partner Dashboard (port 3000)
│   ├── app/                      # App router pages
│   ├── components/               # React components
│   └── lib/                      # API client
├── user-frontend/                # Next.js User Portal (port 3002) - NEW!
│   ├── app/                      # Login, Book, My Dates pages
│   ├── context/                  # Auth context
│   ├── components/               # React components
│   └── lib/                      # API client
└── mock-partner-api/             # Mock partner API for testing
    └── main.go
```

## Prerequisites

- **Go** 1.21+
- **Node.js** 18+
- **PostgreSQL** 14+ (running database named `breeze`)

## Quick Start

### User Portal (NEW)

A separate frontend for users to book dates:

```bash
cd user-frontend
npm install
npm run dev
# Runs on http://localhost:3002
```

Features:
- **Login**: Simple email/password authentication
- **Book Date**: Browse venues, select date/time, book instantly
- **My Dates**: View all your booked dates with status

Demo accounts:
- `user1@test.com` / `password`
- `user2@test.com` / `password`

---

### 1. Configure Environment

The `.env` file in the root directory contains all configuration:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/breeze?sslmode=disable
BACKEND_PORT=8080
FRONTEND_PORT=3000
PARTNER_API_URL=http://localhost:3001
```

Update the `DATABASE_URL` to match your PostgreSQL credentials.

### 2. Start PostgreSQL

Ensure PostgreSQL is running and the `breeze` database exists:

```bash
# Create database if it doesn't exist
createdb breeze
```

### 3. Start the Mock Partner API for Sync Logic between User and Partner Portal

For local testing, start the mock partner API:

```bash
cd mock-partner-api
go run main.go
# Runs on http://localhost:3001
```

### 4. Start the Backend - For data fetching and update

```bash
cd backend
go mod tidy
go run main.go
# Runs on http://localhost:8080
```

The backend will:
- Run database migrations automatically
- Seed the database with 4 sample venues
- Start the HTTP server
- Begin background sync jobs (every 5 minutes)

### 5. Start the Frontend

```bash
cd frontend
npm install
npm run dev
# Runs on http://localhost:3000
```

### 6. Access the Application

Open http://localhost:3000 in your browser to see the partner dashboard.

## API Endpoints

### Dashboard APIs

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/venues` | List all venues |
| `GET` | `/api/v1/venues/{id}` | Get venue details |
| `GET` | `/api/v1/venues/{id}/availability?from=&to=` | Get availability slots |
| `GET` | `/api/v1/venues/{id}/dates` | Get upcoming dates |
| `GET` | `/api/v1/venues/{id}/sync-status` | Get last sync status |
| `POST` | `/api/v1/venues/{id}/sync` | Trigger manual sync |

### Date Management APIs

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/dates` | Create a new date booking |
| `POST` | `/api/v1/dates/{id}/switch-venue` | Switch venue for a date |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Service health check |

## Database Schema

### Tables

- **venues**: Partner venue information
- **availability_slots**: Recurring weekly availability
- **dates**: Scheduled dates between user pairs
- **reservations**: External reservation tracking
- **sync_logs**: Audit trail for sync operations

### Key Indexes

- `idx_venues_city`: For filtering venues by city
- `idx_dates_venue_date`: For date range queries
- `idx_reservations_status`: For sync operations

## Frontend Pages

| Page | URL | Description |
|------|-----|-------------|
| Venues List | `/` | All partner venues |
| Availability | `/venues/{id}` | Weekly calendar view |
| Dates List | `/venues/{id}/dates` | Upcoming dates table |

## Sync Service

The sync service handles:

1. **Creating reservations**: When a date is booked, creates reservation in partner API
2. **Checking cancellations**: Detects external cancellations and marks dates for rescheduling
3. **Retry logic**: Handles transient failures with exponential backoff

Sync runs:
- Automatically every 5 minutes (background job)
- Manually via "Sync Now" button in dashboard

## Design Decisions

### Concurrency & Overbooking Prevention

- Unique constraints on `(venue_id, date, start_time)` in dates table
- Row-level locking when checking capacity
- Reservation status tracking separate from date status

### Error Handling

- Sync logs capture all operations for debugging
- Failed reservations are marked for retry
- Graceful degradation when partner API is unavailable

### Scalability Considerations

- Horizontal scaling: stateless backend, database handles concurrency
- Connection pooling configured for PostgreSQL
- Background jobs don't block HTTP requests

## Future Enhancements

If I had more time, I would add:

1. **Authentication**: JWT-based auth for partner access
2. **Real-time updates**: WebSockets for live availability changes
3. **Auto-rescheduling**: When venue cancels, automatically find alternative venue
4. **Metrics**: Prometheus metrics for sync success rates, latency
5. **Rate limiting**: Protect partner APIs from excessive calls
6. **Tests**: Unit tests for sync logic, integration tests for API

## AI Assistant Usage

This project was built with AI assistance. Key prompts and interactions:

### Prompt Examples

1. "Design a PostgreSQL schema for a partner availability system with recurring weekly slots and reservation tracking"
2. "Create a Go HTTP handler structure with CORS support for a REST API"
3. "Build a React calendar component that shows weekly availability with capacity indicators"

### Key Decisions

- **Accepted**: Using separate `reservations` table to track external state
- **Accepted**: Custom TimeOnly/DateOnly types for proper JSON serialization
- **Modified**: AI suggested using an ORM, but I used raw SQL for better control
- **Modified**: Simplified the auto-rescheduling logic due to time constraints

---

Built for the Breeze Product Engineer Technical Case.
