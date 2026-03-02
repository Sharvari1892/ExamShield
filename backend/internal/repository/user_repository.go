package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash, role string) (*User, error) {
	id := uuid.New().String()

	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, role)
		 VALUES ($1, $2, $3, $4)`,
		id, email, passwordHash, role,
	)

	if err != nil {
		return nil, err
	}

	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, role
		 FROM users
		 WHERE email=$1`,
		email,
	)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
