package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	name string,
	email string,
	birthday time.Time,
	passwordHash string,
) (User, error) {
	const query = `
		INSERT INTO users (name, email, birthday, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, birthday, created_at
	`

	var createdUser User

	err := r.db.QueryRowContext(
		ctx,
		query,
		name,
		email,
		birthday,
		passwordHash,
	).Scan(
		&createdUser.ID,
		&createdUser.Name,
		&createdUser.Email,
		&createdUser.Birthday,
		&createdUser.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return createdUser, nil
}
