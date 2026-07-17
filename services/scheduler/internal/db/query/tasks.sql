-- name: InsertTask :one
INSERT INTO tasks(task, scheduled_at)
    VALUES ($1, $2)
RETURNING
    id;

-- name: GetPendingTasks :many
SELECT id, task FROM tasks WHERE scheduled_at < (NOW() + INTERVAL '30 seconds') AND picked_at IS NULL ORDER BY scheduled_at FOR UPDATE SKIP LOCKED;

-- name: MarkTaskAsPicked :exec
UPDATE tasks SET picked_at = NOW() WHERE id = $1;

-- name: UpdateTaskStarted :exec
UPDATE tasks SET started_at = $1, error = $2 WHERE id = $3;

-- name: UpdateTaskCompleted :exec
UPDATE tasks SET completed_at = $1, error = $2 WHERE id = $3;

-- name: UpdateTaskFailed :exec
UPDATE tasks SET failed_at = $1, error = $2 WHERE id = $3;