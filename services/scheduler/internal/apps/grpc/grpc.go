package grpc

import "github.com/surajgoraicse/orchestrix/services/scheduler/internal/modules/tasks"

type TaskServer struct {
	taskService tasks.ITaskService
}

func NewTaskServer(taskService tasks.ITaskService) *TaskServer {
	return &TaskServer{
		taskService: taskService,
	}
}
