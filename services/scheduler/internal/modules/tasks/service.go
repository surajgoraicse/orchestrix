package tasks

import (
	"context"

	db_sqlc "github.com/surajgoraicse/orchestrix/services/scheduler/internal/db/sqlc"
)

type TaskService struct {
	queries *db_sqlc.Queries
}

func NewTaskService(queries *db_sqlc.Queries) *TaskService {
	return &TaskService{
		queries: queries,
	}
}

func (t *TaskService) ScheduleTask(ctx context.Context, req ScheduleTaskRequest) (ScheduleTaskResponse, error) {
	return ScheduleTaskResponse{}, nil
}

func (t *TaskService) GetTaskStatus(ctx context.Context, taskID string) (TaskStatus, error) {
	return TaskStatus{}, nil
}
