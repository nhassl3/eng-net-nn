package domain

import (
	"time"
)

type Plan struct {
	UUID            string    `json:"uuid"`
	FullName        string    `json:"full_name"`
	Direction       int32     `json:"direction"`
	TaskDescription string    `json:"task_description"`
	EmailToFeedback string    `json:"email_to_feedback"`
	CreatedAt       time.Time `json:"created_at"`
}

type UserPlan struct {
	*User `json:"user"`
	*Plan `json:"plan"`
}

type CreatePlanInput struct {
	FullName        string `json:"full_name"`
	Direction       int32  `json:"direction"`
	TaskDescription string `json:"task_description"`
	EmailToFeedback string `json:"email_to_feedback"`
}
