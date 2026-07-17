-- name: InsertTask :one
INSERT INTO tasks(task, scheduled_at)
    VALUES ($1, $2)
RETURNING
    id;

