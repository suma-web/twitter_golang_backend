package post

import (
	"context"
	"database/sql"
	"errors"
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

func (r *Repository) List(ctx context.Context, limit, offset int) ([]Post, error) {
	const query = `
		SELECT posts.id, posts.user_id, users.name, posts.doc,
		       posts.image_url, posts.created_at
		FROM posts
		JOIN users ON users.id = posts.user_id
		ORDER BY posts.created_at DESC, posts.id DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	posts := make([]Post, 0)
	for rows.Next() {
		var item Post
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.Name, &item.Doc,
			&item.ImageURL, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}

	return posts, nil
}

func (r *Repository) ListByUserName(ctx context.Context, name string, limit, offset int) ([]Post, error) {
	const query = `
		SELECT posts.id, posts.user_id, users.name, posts.doc, posts.image_url, posts.created_at
		FROM posts JOIN users ON users.id = posts.user_id
		WHERE users.name = $1
		ORDER BY posts.created_at DESC, posts.id DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, name, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list user posts: %w", err)
	}
	defer rows.Close()
	posts := make([]Post, 0)
	for rows.Next() {
		var item Post
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.Doc, &item.ImageURL, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user post: %w", err)
		}
		posts = append(posts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user posts: %w", err)
	}
	return posts, nil
}

func (r *Repository) FindByID(ctx context.Context, postID int64) (Post, error) {
	const query = `
		SELECT posts.id, posts.user_id, users.name, posts.doc,
		       posts.image_url, posts.created_at
		FROM posts
		JOIN users ON users.id = posts.user_id
		WHERE posts.id = $1`

	var found Post
	err := r.db.QueryRowContext(ctx, query, postID).Scan(
		&found.ID, &found.UserID, &found.Name, &found.Doc,
		&found.ImageURL, &found.CreatedAt,
	)
	if err != nil {
		return Post{}, fmt.Errorf("find post by id: %w", err)
	}

	return found, nil
}

func (r *Repository) Delete(ctx context.Context, postID, userID int64) (*string, error) {
	const query = `
		DELETE FROM posts
		WHERE id = $1 AND user_id = $2
		RETURNING image_url`
	var imageURL *string
	if err := r.db.QueryRowContext(ctx, query, postID, userID).Scan(&imageURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("delete post: %w", err)
	}
	return imageURL, nil
}
