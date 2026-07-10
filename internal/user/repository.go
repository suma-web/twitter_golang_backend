package user

import (
	"context"
	"database/sql"
	"errors"
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

func (r *Repository) FindByLoginIdentifier(
	ctx context.Context,
	identifier string,
) (User, error) {
	const query = `
		SELECT id, name, email, birthday, password_hash, created_at
		FROM users
		WHERE LOWER(email) = LOWER($1) OR name = $1
		ORDER BY id
		LIMIT 1
	`

	var foundUser User

	err := r.db.QueryRowContext(ctx, query, identifier).Scan(
		&foundUser.ID,
		&foundUser.Name,
		&foundUser.Email,
		&foundUser.Birthday,
		&foundUser.PasswordHash,
		&foundUser.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, sql.ErrNoRows
		}

		return User{}, fmt.Errorf("find user by login identifier: %w", err)
	}

	return foundUser, nil
}
