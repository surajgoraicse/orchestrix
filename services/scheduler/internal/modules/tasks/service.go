package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TaskService struct {
	repo   TaskRepository
	logger *zap.Logger
}

func NewTaskService(repo TaskRepository, logger *zap.Logger) *TaskService {
	return &TaskService{
		repo:   repo,
		logger: logger,
	}
}

// ScheduleTask : Schedule a new task in the database
func (t *TaskService) ScheduleTask(ctx context.Context, req *ScheduleTaskRequest) (*ScheduleTaskResponse, error) {
	t.logger.Info("Received task request: ", zap.Any("req", req))

	// Parse the scheduled_at time
	scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("Invalid scheduled_at time")
	}

	// convert the scheduled time to Unix timestamp
	unixTimestamp := time.Unix(scheduledTime.Unix(), 0)

	// insert it into db
	taskId, err := t.repo.CreateTask(ctx, &Task{
		TaskType:    req.TaskType,
		Payload:     req.Payload,
		MaxRetries:  req.MaxRetries,
		ScheduledAt: &unixTimestamp,
	})
	if err != nil {
		return nil, fmt.Errorf("%s", "Failed to insert task into database : "+err.Error())
	}

	return &ScheduleTaskResponse{
		ID:          taskId,
		TaskType:    req.TaskType,
		Payload:     req.Payload,
		ScheduledAt: req.ScheduledAt,
	}, nil
}

// EditTask : Edits the task
func (t *TaskService) EditTask(ctx context.Context, task *Task) error {
	return t.repo.EditTask(ctx, task)
}

// GetTaskStatus : fetch the task status from database based on the task id
func (t *TaskService) GetTaskStatus(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	return t.repo.FetchTaskByID(ctx, taskID)
}

// FetchDueTasks : fetch the due tasks from database that are ready to be processed
func (t *TaskService) FetchDueTasks(ctx context.Context, limit int) ([]Task, error) {
	return t.repo.FetchDueTasks(ctx, limit)
}

// MarkTaskAsCompleted : Mark the task as completed
func (t *TaskService) MarkTaskAsCompleted(ctx context.Context, taskID uuid.UUID) error {
	return t.repo.MarkTaskAsCompleted(ctx, taskID)
}

// MarkTaskAsFailed : Mark the task as failed
func (t *TaskService) MarkTaskAsFailed(ctx context.Context, taskID uuid.UUID, taskErr error) error {
	return t.repo.MarkTaskAsFailed(ctx, taskID, taskErr)
}

// MarkTaskAsDispatched : Mark the task as dispatch
func (t *TaskService) MarkTaskAsDispatched(ctx context.Context, taskID uuid.UUID) error {
	return t.repo.MarkTaskAsDispatched(ctx, taskID)
}

// MarkTaskAsPicked : Mark the task as picked
func (t *TaskService) MarkTaskAsPicked(ctx context.Context, taskID uuid.UUID) error {

	return t.repo.MarkTaskAsPicked(ctx, taskID)
}

// IncrementAttemptCount : Increment the attempt count of a task
func (t *TaskService) IncrementAttemptCount(ctx context.Context, taskID uuid.UUID) error {
	return t.repo.IncrementAttemptCount(ctx, taskID)
}
