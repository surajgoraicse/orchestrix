package tasks

import (
	"context"
)

type ITaskService interface {
	ScheduleTask(ctx context.Context, req ScheduleTaskRequest) (ScheduleTaskResponse, error)
	GetTaskStatus(ctx context.Context, taskID string) (TaskStatus, error)
}
