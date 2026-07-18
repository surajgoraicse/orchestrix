package tasks

import "time"

// ScheduleTaskRequest represents the request structure for scheduling a task
type ScheduleTaskRequest struct {
	Task        string `json:"task"`
	ScheduledAt string `json:"scheduled_at"`
}


type ScheduleTaskResponse struct {
	ID          string `json:"id"`
	Task        string `json:"task"`
	ScheduledAt string `json:"scheduled_at"`
}

type Task struct {
	ID          string     `json:"id"`
	Task        string     `json:"task"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	PickedAt    *time.Time `json:"picked_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`
	Error       *string    `json:"error,omitempty"`
}
