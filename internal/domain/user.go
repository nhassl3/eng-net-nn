package domain

import (
	"encoding/json"
	"time"
)

type User struct {
	UUID      string    `json:"uuid"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	if u == nil {
		return ErrRedisNotFound
	}
	return json.Unmarshal(data, u)
}

type CreateUserInput struct {
	Username string `json:"username" validator:"required,min=3,max=50"`
	FullName string `json:"full_name" validator:"required,min=2,max=100"`
	Email    string `json:"email" validator:"required,email"`
	Password string `json:"password" validator:"required,min=8"`
}

type SignInInput struct {
	Username string `json:"username" validator:"required"`
	Password string `json:"password" validator:"required"`
}

type GetMeParams struct {
	UUID     *string `json:"uuid"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validator:"required"`
}
