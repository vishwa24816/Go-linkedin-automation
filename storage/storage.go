package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import for its side effects (driver registration)
)

// RequestStatus defines the status of a connection request.
type RequestStatus string

const (
	StatusPending   RequestStatus = "pending"
	StatusAccepted  RequestStatus = "accepted"
	StatusRejected  RequestStatus = "rejected"
	StatusSent      RequestStatus = "sent"
)

// SentRequest represents a sent connection request.
type SentRequest struct {
	ID         int64
	ProfileURL string
	Note       string
	SentAt     time.Time
	Status     RequestStatus
}

// MessageRecord represents a sent follow-up message.
type MessageRecord struct {
	ID           int64
	ProfileURL   string
	Message      string
	SentAt       time.Time
	TemplateUsed string
}

// Storage provides methods for interacting with the database.
type Storage struct {
	db *sql.DB
}

// NewStorage initializes and returns a new Storage instance.
func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.InitDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return storage, nil
}

// InitDB creates necessary tables if they don't exist.
func (s *Storage) InitDB() error {
	createRequestsTableSQL := `
	CREATE TABLE IF NOT EXISTS sent_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_url TEXT NOT NULL UNIQUE,
		note TEXT,
		sent_at DATETIME NOT NULL,
		status TEXT NOT NULL
	);`

	createMessagesTableSQL := `
	CREATE TABLE IF NOT EXISTS message_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_url TEXT NOT NULL,
		message TEXT NOT NULL,
		sent_at DATETIME NOT NULL,
		template_used TEXT,
		UNIQUE(profile_url, message, sent_at) ON CONFLICT IGNORE
	);`

	_, err := s.db.Exec(createRequestsTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create sent_requests table: %w", err)
	}

	_, err = s.db.Exec(createMessagesTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create message_records table: %w", err)
	}

	log.Println("Database tables initialized successfully.")
	return nil
}

// Close closes the database connection.
func (s *Storage) Close() error {
	return s.db.Close()
}

// SaveSentRequest saves a new sent connection request to the database.
func (s *Storage) SaveSentRequest(req *SentRequest) error {
	query := `INSERT INTO sent_requests (profile_url, note, sent_at, status) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, req.ProfileURL, req.Note, req.SentAt, req.Status)
	if err != nil {
		return fmt.Errorf("failed to save sent request: %w", err)
	}
	return nil
}

// GetSentRequestByProfileURL retrieves a sent request by its profile URL.
func (s *Storage) GetSentRequestByProfileURL(profileURL string) (*SentRequest, error) {
	query := `SELECT id, profile_url, note, sent_at, status FROM sent_requests WHERE profile_url = ?`
	row := s.db.QueryRow(query, profileURL)

	req := &SentRequest{}
	err := row.Scan(&req.ID, &req.ProfileURL, &req.Note, &req.SentAt, &req.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get sent request: %w", err)
	}
	return req, nil
}

// UpdateRequestStatus updates the status of a sent connection request.
func (s *Storage) UpdateRequestStatus(profileURL string, status RequestStatus) error {
	query := `UPDATE sent_requests SET status = ? WHERE profile_url = ?`
	_, err := s.db.Exec(query, status, profileURL)
	if err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}
	return nil
}

// GetCountOfSentRequestsToday returns the number of requests sent today.
func (s *Storage) GetCountOfSentRequestsToday() (int, error) {
	today := time.Now().Format("2006-01-02") + " 00:00:00"
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02") + " 00:00:00"

	query := `SELECT COUNT(*) FROM sent_requests WHERE sent_at >= ? AND sent_at < ?`
	var count int
	err := s.db.QueryRow(query, today, tomorrow).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get count of sent requests today: %w", err)
	}
	return count, nil
}

// SaveMessageRecord saves a new message record to the database.
func (s *Storage) SaveMessageRecord(msg *MessageRecord) error {
	query := `INSERT INTO message_records (profile_url, message, sent_at, template_used) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, msg.ProfileURL, msg.Message, msg.SentAt, msg.TemplateUsed)
	if err != nil {
		return fmt.Errorf("failed to save message record: %w", err)
	}
	return nil
}

// GetMessageRecord retrieves a message record for a profile.
func (s *Storage) GetMessageRecord(profileURL string) (*MessageRecord, error) {
	query := `SELECT id, profile_url, message, sent_at, template_used FROM message_records WHERE profile_url = ? ORDER BY sent_at DESC LIMIT 1`
	row := s.db.QueryRow(query, profileURL)

	msg := &MessageRecord{}
	err := row.Scan(&msg.ID, &msg.ProfileURL, &msg.Message, &msg.SentAt, &msg.TemplateUsed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get message record: %w", err)
	}
	return msg, nil
}

// GetProfileURLsWithPendingRequests retrieves all profile URLs that have a pending connection request.
func (s *Storage) GetProfileURLsWithPendingRequests() ([]string, error) {
	query := `SELECT profile_url FROM sent_requests WHERE status = ?`
	rows, err := s.db.Query(query, StatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile URLs with pending requests: %w", err)
	}
	defer rows.Close()

	var profileURLs []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, fmt.Errorf("failed to scan profile URL: %w", err)
		}
		profileURLs = append(profileURLs, url)
	}
	return profileURLs, nil
}

// GetProfilesWithAcceptedRequestsWithoutMessage retrieves profiles with accepted requests that haven't received a message.
func (s *Storage) GetProfilesWithAcceptedRequestsWithoutMessage() ([]string, error) {
	query := `
	SELECT sr.profile_url
	FROM sent_requests sr
	LEFT JOIN message_records mr ON sr.profile_url = mr.profile_url
	WHERE sr.status = ? AND mr.id IS NULL;`

	rows, err := s.db.Query(query, StatusAccepted)
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles with accepted requests without message: %w", err)
	}
	defer rows.Close()

	var profileURLs []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, fmt.Errorf("failed to scan profile URL: %w", err)
		}
		profileURLs = append(profileURLs, url)
	}
	return profileURLs, nil
}
