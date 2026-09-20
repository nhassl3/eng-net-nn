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
	Username string `json:"username" binding:"required,min=3,max=30"`
	FullName string `json:"full_name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
	
type SignInInput struct {
	Username string `json:"username"`
	Email    string `json:"email" binding:"omitempty,email"`
	ID       string `json:"id"`
	Password string `json:"password" binding:"required"`
}

type GetMeParams struct {
	UUID     *string `json:"uuid" binding:"omitempty"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Username *string `json:"username" binding:"omitempty,min=3,max=30"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
