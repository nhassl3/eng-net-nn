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

type Plans struct {
	Plans []Plan `json:"plans"`
	Total int    `json:"total"`
}

type UserPlan struct {
	*User `json:"user"`
	*Plan `json:"plan"`
}

type CreatePlanInput struct {
	FullName        string `json:"full_name" binding:"required"`
	Direction       int32  `json:"direction" binding:"required,min=1"`
	TaskDescription string `json:"task_description" binding:"required"`
	EmailToFeedback string `json:"email_to_feedback" binding:"required,email"`
}

type CreatePlanInputEmail struct {
	FullName        string `json:"full_name" binding:"required"`
	Direction       string `json:"direction" binding:"required,min=1"`
	TaskDescription string `json:"task_description" binding:"required"`
	EmailToFeedback string `json:"email_to_feedback" binding:"required,email"`
}
