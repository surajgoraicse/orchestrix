package rest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
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
	router.GET("/:taskID", t.GetTaskByID)
	router.PUT("/:taskID", t.EditTask)
	router.DELETE("/:taskID", t.DeleteTask)
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

func (t *TaskHandler) GetTaskByID(c *echo.Context) error {
	taskID := c.Param("taskID")
	if taskID == "" {
		return response.NewResponse(c, http.StatusBadRequest, "Invalid request", nil, errors.New("taskID is required"))
	}
	parsedUUID, err := uuid.Parse(taskID)
	if err != nil {
		return response.NewResponse(c, http.StatusBadRequest, "Invalid request", nil, err)
	}
	task, err := t.taskService.GetTaskStatus(c.Request().Context(), parsedUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.NewResponse(c, http.StatusNotFound, "Task not found", nil, err)
		}
		return response.NewResponse(c, http.StatusInternalServerError, "Failed to get task status", nil, err)
	}

	return response.NewResponse(c, http.StatusOK, "Task status fetched successfully", task, nil)
}

func (t *TaskHandler) EditTask(c *echo.Context) error {
	taskID := c.Param("taskID")
	if taskID == "" {
		return response.NewResponse(c, http.StatusBadRequest, "Invalid request", nil, errors.New("taskID is required"))
	}
	uuid, err := uuid.Parse(taskID)
	if err != nil {
		return response.NewResponse(c, http.StatusBadRequest, "Invalid request", nil, err)
	}

	var taskRequest tasks.ScheduleTaskRequest
	if err := c.Bind(&taskRequest); err != nil {
		return response.NewResponse(c, http.StatusBadRequest, "Invalid request", nil, err)
	}

	task, err := t.taskService.EditTask(c.Request().Context(), uuid, taskRequest)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.NewResponse(c, http.StatusNotFound, "Task not found", nil, err)
		}
		return response.NewResponse(c, http.StatusInternalServerError, "Failed to edit task", nil, err)
	}

	return response.NewResponse(c, http.StatusOK, "Task edited successfully", task, nil)
}

func (t *TaskHandler) DeleteTask(c *echo.Context) error {
	taskID := c.Param("taskID")
	if taskID == "" {
		return response.NewResponse(c, http.StatusBadRequest, "Invalid request", nil, errors.New("taskID is required"))
	}
	parsedUUID, err := uuid.Parse(taskID)
	if err != nil {
		return response.NewResponse(c, http.StatusBadRequest, "Invalid request", nil, err)
	}
	err = t.taskService.DeleteTask(c.Request().Context(), parsedUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.NewResponse(c, http.StatusNotFound, "Task not found", nil, err)
		}
		return response.NewResponse(c, http.StatusInternalServerError, "Failed to delete task", nil, err)
	}

	return response.NewResponse(c, http.StatusOK, "Task deleted successfully", nil, nil)
}
