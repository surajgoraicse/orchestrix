package rest

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/surajgoraicse/orchestrix/libs/go-libs/response"
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
	var taskRequest tasks.ScheduleTaskRequest
	if err := c.Bind(&taskRequest); err != nil {
		return response.NewResponse(c, http.StatusBadRequest, "Invalid request", nil, err)
	}

	task, err := t.taskService.ScheduleTask(c.Request().Context(), &taskRequest)
	if err != nil {
		return response.NewResponse(c, http.StatusInternalServerError, "Failed to schedule task", nil, err)
	}

	return response.NewResponse(c, http.StatusOK, "Task scheduled successfully", task, nil)
}

func (t *TaskHandler) GetTaskStatus(c *echo.Context) error {
	return nil
}
