package tasks

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID           uuid.UUID  `json:"id"`
	TaskType     string     `json:"task_type"`
	Payload      []byte     `json:"payload"`
	Status       string     `json:"status"`
	MaxRetries   int        `json:"max_retries"`
	AttemptCount int        `json:"attempt_count"`
	ScheduledAt  *time.Time `json:"scheduled_at"`
	PickedAt     *time.Time `json:"picked_at,omitempty"`
	DispatchedAt *time.Time `json:"dispatched_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	FailedAt     *time.Time `json:"failed_at,omitempty"`
	Error        *string    `json:"error,omitempty"`
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task *Task) (string, error)
	EditTask(ctx context.Context, task *Task) error
	FetchDueTasks(ctx context.Context, limit int) ([]Task, error)
	FetchTaskByID(ctx context.Context, taskID uuid.UUID) (*Task, error)
	SoftDeleteTask(ctx context.Context, taskID uuid.UUID) error
	MarkTaskAsPicked(ctx context.Context, taskID uuid.UUID) error
	MarkTaskAsDispatched(ctx context.Context, taskIDs uuid.UUID) error
	MarkTaskAsCompleted(ctx context.Context, taskID uuid.UUID) error
	MarkTaskAsFailed(ctx context.Context, taskID uuid.UUID, err error) error
	IncrementAttemptCount(ctx context.Context, taskID uuid.UUID) error
}

type ITaskService interface {
	ScheduleTask(ctx context.Context, req *ScheduleTaskRequest) (*ScheduleTaskResponse, error)
	GetTaskStatus(ctx context.Context, taskID uuid.UUID) (*Task, error)
	EditTask(ctx context.Context, taskID uuid.UUID, req ScheduleTaskRequest) (ScheduleTaskResponse, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	IncrementAttemptCount(ctx context.Context, taskID uuid.UUID) error
	FetchDueTasks(ctx context.Context, limit int) ([]Task, error)
	MarkTaskAsCompleted(ctx context.Context, taskID uuid.UUID) error
	MarkTaskAsFailed(ctx context.Context, taskID uuid.UUID, err error) error
	MarkTaskAsPicked(ctx context.Context, taskID uuid.UUID) error
	MarkTaskAsDispatched(ctx context.Context, taskID uuid.UUID) error
}

type TaskPublisher interface {
	Publish(ctx context.Context, task *Task) error
}

// ScheduleTaskRequest represents the request structure for scheduling a task
type ScheduleTaskRequest struct {
	TaskType    string `json:"task_type"`
	Payload     []byte `json:"payload"`
	MaxRetries  int    `json:"max_retries"`
	ScheduledAt string `json:"scheduled_at"`
}

type ScheduleTaskResponse struct {
	ID          string `json:"id"`
	TaskType    string `json:"task_type"`
	Payload     []byte `json:"payload"`
	ScheduledAt string `json:"scheduled_at"`
}
