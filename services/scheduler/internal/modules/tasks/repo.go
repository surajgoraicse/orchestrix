package tasks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/surajgoraicse/orchestrix/libs/go-libs/utils"
	db_sqlc "github.com/surajgoraicse/orchestrix/services/scheduler/internal/db/sqlc"
)

type TaskRepo struct {
	queries *db_sqlc.Queries
}

func NewTaskRepo(queries *db_sqlc.Queries) *TaskRepo {
	return &TaskRepo{
		queries: queries,
	}
}

func (t *TaskRepo) CreateTask(ctx context.Context, task *Task) (string, error) {
	var scheduledAt pgtype.Timestamptz
	if task.ScheduledAt != nil {
		scheduledAt.Time = *task.ScheduledAt
		scheduledAt.Valid = true
	}

	id, err := t.queries.InsertTask(ctx, db_sqlc.InsertTaskParams{
		TaskType:    task.TaskType,
		Payload:     task.Payload,
		MaxRetries:  int32(task.MaxRetries),
		ScheduledAt: scheduledAt,
	})
	if err != nil {
		return "", err
	}
	return id.String(), nil

}

// FetchDueTasks : Fetches the due tasks from the database that are ready to be processed
func (t *TaskRepo) FetchDueTasks(ctx context.Context, limit int32) ([]Task, error) {
	dueTasks, err := t.queries.GetDueTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, task := range dueTasks {
		tasks = append(tasks, Task{
			ID:       task.ID.String(),
			TaskType: task.TaskType,
			Payload:  task.Payload,
		})
	}
	return tasks, nil
}

func (t *TaskRepo) MarkTaskAsDispatched(ctx context.Context, taskIDs []uuid.UUID) error {
	for _, taskID := range taskIDs {
		pgUUID, err := utils.GetPgUUIDFromUUID(taskID)
		if err != nil {
			return err
		}
		err = t.queries.MarkTaskAsDispatched(ctx, pgUUID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TaskRepo) MarkTaskAsCompleted(ctx context.Context, taskID uuid.UUID) error {
	pgUUID, err := utils.GetPgUUIDFromUUID(taskID)
	if err != nil {
		return err
	}
	err = t.queries.UpdateTaskCompleted(ctx, db_sqlc.UpdateTaskCompletedParams{
		CompletedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Error: pgtype.Text{
			String: "",
			Valid:  false,
		},
		ID: pgUUID,
	})
	if err != nil {
		return err
	}
	return nil
}

// mark the task as failed and also update the error message
func (t *TaskRepo) MarkTaskAsFailed(ctx context.Context, taskID uuid.UUID, err error) error {
	pgUUID, err := utils.GetPgUUIDFromUUID(taskID)
	if err != nil {
		return err
	}
	err = t.queries.UpdateTaskFailed(ctx, db_sqlc.UpdateTaskFailedParams{
		FailedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Error: pgtype.Text{
			String: err.Error(),
			Valid:  true,
		},
		ID: pgUUID,
	})
	if err != nil {
		return err
	}
	return nil
}

// IncrementAttempCount : Increments the attemp count of a task
func (t *TaskRepo) IncrementAttemptCount(ctx context.Context, taskID uuid.UUID) error {
	pgUUID, err := utils.GetPgUUIDFromUUID(taskID)
	if err != nil {
		return err
	}
	err = t.queries.IncrementAttemptCount(ctx, pgUUID)
	if err != nil {
		return err
	}
	return nil
}

// MarkTaskAsPicked marks the task as picked and also increments the attempt count
func (t *TaskRepo) MarkTaskAsPicked(ctx context.Context, taskID uuid.UUID) error {
	pgUUID, err := utils.GetPgUUIDFromUUID(taskID)
	if err != nil {
		return err
	}
	err = t.queries.MarkTaskAsPicked(ctx, pgUUID)
	if err != nil {
		return err
	}
	return nil
}
