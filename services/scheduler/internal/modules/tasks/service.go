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

func (t *TaskService) GetTaskStatus(ctx context.Context, taskID string) (*Task, error) {
	parsedUUID, err := uuid.Parse(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %w", err)
	}
	return t.repo.FetchTaskByID(ctx, parsedUUID)
}
