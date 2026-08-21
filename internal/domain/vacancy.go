package domain

import (
	"time"
)

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

type RespondVacancies struct {
	RespondVacancies []RespondVacancy `json:"respond_vacancies"`
	Total            int              `json:"total"`
}

type RespondVacancy struct {
	UUID        string    `json:"uuid"`
	FullName    string    `json:"fullName"`
	PhoneNumber string    `json:"phoneNumber"`
	Email       string    `json:"email"`
	City        string    `json:"city"`
	Exp         string    `json:"exp"`
	Description string    `json:"description"`
	ResumeUrl   string    `json:"resumeUrl"`
	VacancyId   string    `json:"vacancyId"`
	CreatedAt   time.Time `json:"created_at"`
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
}

type FileUploadInput struct {
	FileData []byte
}

type UpdatedVacancyInput struct {
	Jd          *int32   `json:"jd,omitempty"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	RequiredExp *float64 `json:"required_exp,omitempty"`
	PayDay      *float64 `json:"pay_day,omitempty"`
	Skills      []string `json:"skills,omitempty"`
}

type CreateVacancyInput struct {
	Jd          int32    `json:"jd,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	RequiredExp *float64 `json:"required_exp,omitempty"`
	PayDay      *float64 `json:"pay_day,omitempty"`
	Skills      []string `json:"skills,omitempty"`
}
