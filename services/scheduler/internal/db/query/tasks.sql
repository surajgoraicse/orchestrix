-- name: InsertTask :one
INSERT INTO tasks(task_type, payload, max_retries, scheduled_at)
    VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetTask :one
SELECT id, task_type, payload, scheduled_at, picked_at, started_at, completed_at, failed_at, error FROM tasks WHERE id = $1;

-- name: GetDueTasks :many
SELECT id, task_type, payload FROM tasks WHERE scheduled_at < (NOW() + INTERVAL '30 seconds') AND attempt_count < max_retries AND picked_at IS NULL ORDER BY scheduled_at FOR UPDATE SKIP LOCKED LIMIT $1;

-- name: MarkTaskAsPicked :exec
UPDATE tasks SET picked_at = NOW(), attempt_count = attempt_count + 1 WHERE id = $1;

-- name: MarkTaskAsDispatched :exec
UPDATE tasks SET dispatched_at = NOW() WHERE id = $1;

-- -- name: UpdateTaskStarted :exec
-- UPDATE tasks SET started_at = $1, error = $2 WHERE id = $3;

-- name: UpdateTaskCompleted :exec
UPDATE tasks SET completed_at = $1, error = $2 WHERE id = $3;

-- name: UpdateTaskFailed :exec
UPDATE tasks SET failed_at = $1, error = $2 WHERE id = $3;

-- name: IncrementAttemptCount :exec
UPDATE tasks SET attempt_count = attempt_count + 1 WHERE id = $1;