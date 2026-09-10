package domain

import (
	"time"
)

type Vacancy struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	RequiredExp string    `json:"required_exp"`
	Skills      []string  `json:"skills"`
	PayDay      float64   `json:"pay_day"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type JobDirection struct {
	Id            int32    `json:"id"`
	JdName        string   `json:"jd_name"`
	JdDescription string   `json:"jd_description"`
	JdTags        []string `json:"jd_tags"`
}

type JobDirections struct {
	JobDirections []JobDirection `json:"job_directions"`
}

type VacancyWithJd struct {
	Vacancy      `json:",inline"`
	JobDirection `json:",inline"`
}

// VacanciesWithJd - main struct
type VacanciesWithJd struct {
	VacanciesWithJd []VacancyWithJd `json:"vacancies"`
}

type CreateJobDirectionInput struct {
	Name        string   `json:"name" binding:"required"`
	Tags        []string `json:"tags"`
	Description string   `json:"description" binding:"required"`
}

type UpdateJobDirectionInput struct {
	Name        *string  `json:"name" binding:"omitempty"`
	Tags        []string `json:"tags" binding:"omitempty"`
	Description *string  `json:"description" binding:"omitempty"`
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
}

type ApplicantsFormInput struct {
	FullName    string `json:"fullName" binding:"omitempty"`
	PhoneNumber string `json:"phoneNumber" binding:"omitempty"`
	Email       string `json:"email" binding:"required,email"`
	City        string `json:"city" binding:"required,city"`
	Exp         string `json:"exp" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty"`
}

type FileUploadInput struct {
	FileData []byte
}

type UpdatedVacancyInput struct {
	Jd          *int32   `json:"jd" binding:"omitempty"`
	Name        *string  `json:"name" binding:"omitempty"`
	Description *string  `json:"description" binding:"omitempty"`
	RequiredExp *string  `json:"required_exp" binding:"omitempty"`
	PayDay      *float64 `json:"pay_day" binding:"omitempty"`
	Skills      []string `json:"skills" binding:"omitempty"`
}

type CreateVacancyInput struct {
	Jd          int32    `json:"jd" binding:"omitempty"`
	Name        string   `json:"name" binding:"omitempty"`
	Description string   `json:"description" binding:"omitempty"`
	RequiredExp *string  `json:"required_exp" binding:"omitempty"`
	PayDay      *float64 `json:"pay_day" binding:"omitempty"`
	Skills      []string `json:"skills" binding:"omitempty"`
}
