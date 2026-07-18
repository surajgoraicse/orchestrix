package rest

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
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
	taskID := c.Param("taskID")
	if taskID == "" {
		return response.NewResponse(c, http.StatusBadRequest, "Invalid request", nil, errors.New("taskID is required"))
	}

	task, err := t.taskService.GetTaskStatus(c.Request().Context(), taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.NewResponse(c, http.StatusNotFound, "Task not found", nil, err)
		}
		return response.NewResponse(c, http.StatusInternalServerError, "Failed to get task status", nil, err)
	}

	return response.NewResponse(c, http.StatusOK, "Task status fetched successfully", task, nil)
}
