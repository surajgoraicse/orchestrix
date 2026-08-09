package tasks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/surajgoraicse/orchestrix/libs/go-libs/database/pg"
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

// CreateTask : Creates a new task
func (t *TaskRepo) CreateTask(ctx context.Context, task *Task) (uuid.UUID, error) {
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
		return uuid.Nil, err
	}
	return uuid.UUID(id.Bytes), nil
}

// FetchTaskByID : Fetches the task by ID from the database
func (t *TaskRepo) FetchTaskByID(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	pgUUID, err := pg.GetPgUUIDFromUUID(taskID)
	if err != nil {
		return nil, err
	}
	dbtask, err := t.queries.GetTask(ctx, pgUUID)
	if err != nil {
		return nil, err
	}
	return &Task{
		ID:           uuid.UUID(dbtask.ID.Bytes),
		TaskType:     dbtask.TaskType,
		Payload:      dbtask.Payload,
		Status:       string(dbtask.Status),
		MaxRetries:   int(dbtask.MaxRetries),
		AttemptCount: int(dbtask.AttemptCount),
		ScheduledAt:  pg.ToTimePtr(dbtask.ScheduledAt),
		PickedAt:     pg.ToTimePtr(dbtask.PickedAt),
		DispatchedAt: pg.ToTimePtr(dbtask.DispatchedAt),
		CompletedAt:  pg.ToTimePtr(dbtask.CompletedAt),
		FailedAt:     pg.ToTimePtr(dbtask.FailedAt),
		Error:        pg.ToStringPtr(dbtask.Error),
	}, nil
}

// FetchDueTasks : Fetches the due tasks from the database that are ready to be processed
func (t *TaskRepo) FetchDueTasks(ctx context.Context, limit int) ([]Task, error) {
	dueTasks, err := t.queries.GetDueTasks(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, task := range dueTasks {
		tasks = append(tasks, Task{
			ID:       uuid.UUID(task.ID.Bytes),
			TaskType: task.TaskType,
			Payload:  task.Payload,
		})
	}
	return tasks, nil
}

// EditTask : Edits the task
func (t *TaskRepo) EditTask(ctx context.Context, task *Task) error {
	pgUUID, err := pg.GetPgUUIDFromUUID(task.ID)
	if err != nil {
		return err
	}
	var scheduledAt pgtype.Timestamptz
	if task.ScheduledAt != nil {
		scheduledAt.Time = *task.ScheduledAt
		scheduledAt.Valid = true
	}
	err = t.queries.EditTask(ctx, db_sqlc.EditTaskParams{
		TaskType:    task.TaskType,
		Payload:     task.Payload,
		MaxRetries:  int32(task.MaxRetries),
		ScheduledAt: scheduledAt,
		ID:          pgUUID,
	})
	if err != nil {
		return err
	}
	return nil
}

// SoftDeleteTask : Soft deletes the task
func (t *TaskRepo) SoftDeleteTask(ctx context.Context, taskID uuid.UUID) error {
	pgUUID, err := pg.GetPgUUIDFromUUID(taskID)
	if err != nil {
		return err
	}
	err = t.queries.SoftDeleteTask(ctx, pgUUID)
	if err != nil {
		return err
	}
	return nil
}

// MarkTaskAsDispatched : Marks the task as dispatched
func (t *TaskRepo) MarkTaskAsDispatched(ctx context.Context, taskID uuid.UUID) error {
	pgUUID, err := pg.GetPgUUIDFromUUID(taskID)
	if err != nil {
		return err
	}
	err = t.queries.MarkTaskAsDispatched(ctx, pgUUID)
	if err != nil {
		return err
	}
	return nil
}

// MarkTaskAsCompleted : Marks the task as completed
func (t *TaskRepo) MarkTaskAsCompleted(ctx context.Context, taskID uuid.UUID) error {
	pgUUID, err := pg.GetPgUUIDFromUUID(taskID)
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
func (t *TaskRepo) MarkTaskAsFailed(ctx context.Context, taskID uuid.UUID, taskErr error) error {
	pgUUID, err := pg.GetPgUUIDFromUUID(taskID)
	if err != nil {
		return err
	}
	err = t.queries.UpdateTaskFailed(ctx, db_sqlc.UpdateTaskFailedParams{
		FailedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Error: pgtype.Text{
			String: taskErr.Error(),
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
	pgUUID, err := pg.GetPgUUIDFromUUID(taskID)
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
	pgUUID, err := pg.GetPgUUIDFromUUID(taskID)
	if err != nil {
		return err
	}
	err = t.queries.MarkTaskAsPicked(ctx, pgUUID)
	if err != nil {
		return err
	}
	return nil
}
