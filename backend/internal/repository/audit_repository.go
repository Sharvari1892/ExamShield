package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Sharvari1892/examshield/internal/domain"
)

type AuditRepository struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{db: db}
}

// Fetch last hash for session
func (r *AuditRepository) getLastHash(ctx context.Context, sessionID string) (string, error) {
	var lastHash string

	err := r.db.QueryRow(ctx,
		`SELECT current_hash
		 FROM audit_events
		 WHERE session_id=$1
		 ORDER BY timestamp DESC
		 LIMIT 1`,
		sessionID,
	).Scan(&lastHash)

	if err != nil {
		// No previous event → first event in chain
		return "", nil
	}

	return lastHash, nil
}

// Create new audit event with hash chain
func (r *AuditRepository) CreateEvent(
	ctx context.Context,
	sessionID string,
	eventType string,
	payload interface{},
) error {

	// 1️⃣ Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 2️⃣ Get previous hash
	prevHash, err := r.getLastHash(ctx, sessionID)
	if err != nil {
		return err
	}

	// 3️⃣ Timestamp
	timestamp := time.Now().UTC()

	// 4️⃣ Create hash input
	hashInput := prevHash + string(payloadBytes) + timestamp.Format(time.RFC3339Nano)

	// 5️⃣ SHA256
	hash := sha256.Sum256([]byte(hashInput))
	currentHash := hex.EncodeToString(hash[:])

	// 6️⃣ Insert into DB
	_, err = r.db.Exec(ctx,
		`INSERT INTO audit_events
		 (id, session_id, event_type, payload, timestamp, prev_hash, current_hash)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		uuid.New().String(),
		sessionID,
		eventType,
		payloadBytes,
		timestamp,
		prevHash,
		currentHash,
	)

	return err
}

func (r *AuditRepository) GetEventsBySession(
	ctx context.Context,
	sessionID string,
) ([]domain.AuditEvent, error) {

	rows, err := r.db.Query(ctx, `
		SELECT prev_hash, payload::text, timestamp, current_hash
		FROM audit_events
		WHERE session_id = $1
		ORDER BY timestamp ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.AuditEvent

	for rows.Next() {
		var e domain.AuditEvent
		var timestamp time.Time

		if err := rows.Scan(
			&e.PrevHash,
			&e.Payload,
			&timestamp,
			&e.CurrentHash,
		); err != nil {
			return nil, err
		}

		// IMPORTANT: format timestamp exactly as stored in hash
		e.Timestamp = timestamp.UTC().Format(time.RFC3339Nano)

		events = append(events, e)
	}

	return events, nil
}
