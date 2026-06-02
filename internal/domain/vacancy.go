package domain

import "time"

type Vacancy struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	RequiredExp float64   `json:"required_exp"`
	Skills      []string  `json:"skills"`
	PayDay      float64   `json:"pay_day"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Vacancies struct {
	Vacancies []Vacancy `json:"vacancies"`
	Total     int       `json:"total"`
}

type ApplicantsFormInput struct {
	FullName    string `json:"fullName,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Email       string `json:"email,omitempty" validator:"required,email"`
	City        string `json:"city,omitempty" validator:"required,city"`
	Exp         string `json:"exp,omitempty"`
	Description string `json:"description,omitempty"`
	Resume      string `json:"resume,omitempty"` // like a link to s3 storage (minio)
}

type UpdatedVacancyInput struct {
	Jd          *int32
	Name        *string
	Description *string
	RequiredExp *float64
	PayDay      *float64
	Skills      []string
}

type CreateVacancyInput struct {
	Jd          int32    `json:"jd,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	RequiredExp *float64 `json:"required_exp,omitempty"`
	PayDay      *float64 `json:"pay_day,omitempty"`
	Skills      []string `json:"skills,omitempty"`
}
