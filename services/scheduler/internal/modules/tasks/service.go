package tasks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func (t *TaskService) ScheduleTask(ctx context.Context, req *ScheduleTaskRequest) (*ScheduleTaskResponse, error) {
	log.Println("Received task request: ", req)

	// Parse the scheduled_at time
	scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("Invalid scheduled_at time")
	}

	// convert the scheduled time to Unix timestamp
	unixTimestamp := time.Unix(scheduledTime.Unix(), 0)

	// insert it into db
	taskId, err := t.insertTaskIntoDb(ctx, Task{
		Task:        req.Task,
		ScheduledAt: &unixTimestamp,
	})
	if err != nil {
		return nil, fmt.Errorf("%s", "Failed to insert task into database : "+err.Error())
	}

	return &ScheduleTaskResponse{
		ID:          taskId,
		Task:        req.Task,
		ScheduledAt: req.ScheduledAt,
	}, nil

}

func (t *TaskService) GetTaskStatus(ctx context.Context, taskID string) (*Task, error) {
	return t.getTaskFromDB(ctx, taskID)
}

func (t *TaskService) insertTaskIntoDb(ctx context.Context, task Task) (string, error) {
	var scheduledAt pgtype.Timestamp
	if task.ScheduledAt != nil {
		scheduledAt.Time = *task.ScheduledAt
		scheduledAt.Valid = true
	}

	uuidVal, err := t.queries.InsertTask(ctx, db_sqlc.InsertTaskParams{
		Task:        task.Task,
		ScheduledAt: scheduledAt,
	})
	if err != nil {
		return "", err
	}

	var u uuid.UUID
	copy(u[:], uuidVal.Bytes[:])
	return u.String(), nil
}

func (t *TaskService) getTaskFromDB(ctx context.Context, taskID string) (*Task, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(taskID); err != nil {
		return &Task{}, err
	}

	dbTask, err := t.queries.GetTask(ctx, pgUUID)
	if err != nil {
		return &Task{}, err
	}

	var u uuid.UUID
	copy(u[:], dbTask.ID.Bytes[:])

	return &Task{
		ID:          u.String(),
		Task:        dbTask.Task,
		ScheduledAt: toTimePtr(dbTask.ScheduledAt),
		PickedAt:    toTimePtr(dbTask.PickedAt),
		StartedAt:   toTimePtr(dbTask.StartedAt),
		CompletedAt: toTimePtr(dbTask.CompletedAt),
		FailedAt:    toTimePtr(dbTask.FailedAt),
		Error:       toStringPtr(dbTask.Error),
	}, nil

}

func toTimePtr(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func toStringPtr(txt pgtype.Text) *string {
	if !txt.Valid {
		return nil
	}
	s := txt.String
	return &s
}
