package rest

import (
	"github.com/labstack/echo/v5"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/modules/tasks"
)

type TaskHandler struct {
	taskService tasks.ITaskService
}

func NewTaskHandler(taskService tasks.ITaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (t *TaskHandler) RegisterRoutes(router *echo.Group) {
	router.POST("/", t.ScheduleTask)
	router.GET("/:taskID", t.GetTaskStatus)
}

func (t *TaskHandler) ScheduleTask(c *echo.Context) error {
	return nil
}

func (t *TaskHandler) GetTaskStatus(c *echo.Context) error {
	return nil
}
