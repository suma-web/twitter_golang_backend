package post

import "time"

type Post struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Doc       string    `json:"doc"`
	ImageURL  *string   `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
}
