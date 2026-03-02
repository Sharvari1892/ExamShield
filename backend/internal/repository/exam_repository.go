package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Exam struct {
	ID              string
	Title           string
	DurationSeconds int
	Published       bool
}

type Question struct {
	ID              string
	ExamID          string
	DifficultyLevel int
	Content         string
}

type ExamRepository struct {
	db *pgxpool.Pool
}

func NewExamRepository(db *pgxpool.Pool) *ExamRepository {
	return &ExamRepository{db: db}
}

func (r *ExamRepository) CreateExam(ctx context.Context, title string, duration int) (*Exam, error) {
	id := uuid.New().String()

	_, err := r.db.Exec(ctx,
		`INSERT INTO exams (id, title, duration_seconds)
		 VALUES ($1, $2, $3)`,
		id, title, duration,
	)

	if err != nil {
		return nil, err
	}

	return &Exam{
		ID:              id,
		Title:           title,
		DurationSeconds: duration,
	}, nil
}

func (r *ExamRepository) CreateQuestion(ctx context.Context, examID string, difficulty int, content string) (*Question, error) {
	id := uuid.New().String()

	_, err := r.db.Exec(ctx,
		`INSERT INTO questions (id, exam_id, difficulty_level, content)
		 VALUES ($1, $2, $3, $4)`,
		id, examID, difficulty, content,
	)

	if err != nil {
		return nil, err
	}

	return &Question{
		ID:              id,
		ExamID:          examID,
		DifficultyLevel: difficulty,
		Content:         content,
	}, nil
}

func (r *ExamRepository) GetExamWithQuestions(ctx context.Context, examID string) (*Exam, []Question, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, title, duration_seconds, published FROM exams WHERE id=$1`,
		examID,
	)

	var exam Exam
	err := row.Scan(&exam.ID, &exam.Title, &exam.DurationSeconds, &exam.Published)
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, exam_id, difficulty_level, content
		 FROM questions WHERE exam_id=$1`,
		examID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		rows.Scan(&q.ID, &q.ExamID, &q.DifficultyLevel, &q.Content)
		questions = append(questions, q)
	}

	return &exam, questions, nil
}
