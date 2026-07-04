package user

import "time"

type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Birthday string `json:"birthday"`
	Password string `json:"password"`
}

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Birthday  time.Time `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type SignupResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Birthday  string `json:"birthday"`
	CreatedAt string `json:"created_at"`
}
