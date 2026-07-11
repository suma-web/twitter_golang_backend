package post

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, userID int64, doc string, imageURL *string) (Post, error) {
	const query = `
		WITH created AS (
			INSERT INTO posts (user_id, doc, image_url)
			VALUES ($1, $2, $3)
			RETURNING id, user_id, doc, image_url, created_at
		)
		SELECT created.id, created.user_id, users.name, created.doc,
		       created.image_url, created.created_at
		FROM created JOIN users ON users.id = created.user_id`
	var created Post
	err := r.db.QueryRowContext(ctx, query, userID, doc, imageURL).Scan(
		&created.ID, &created.UserID, &created.Name, &created.Doc,
		&created.ImageURL, &created.CreatedAt,
	)
	if err != nil {
		return Post{}, fmt.Errorf("create post: %w", err)
	}
	return created, nil
}
