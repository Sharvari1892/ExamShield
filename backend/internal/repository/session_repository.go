package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	ID        string
	UserID    string
	ExamID    string
	StartTime time.Time
	EndTime   time.Time
	Status    string
}

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(
	ctx context.Context,
	userID string,
	examID string,
	durationSeconds int,
) (*Session, error) {

	id := uuid.New().String()
	start := time.Now().UTC()
	end := start.Add(time.Duration(durationSeconds) * time.Second)

	_, err := r.db.Exec(ctx,
		`INSERT INTO exam_sessions
		 (id, user_id, exam_id, start_time, end_time, status)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		id, userID, examID, start, end, "active",
	)

	if err != nil {
		return nil, err
	}

	return &Session{
		ID:        id,
		UserID:    userID,
		ExamID:    examID,
		StartTime: start,
		EndTime:   end,
		Status:    "active",
	}, nil
}
