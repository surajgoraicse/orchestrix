**Scheduler**: The scheduler is the front-end server of the system. It receives tasks from the clients and schedules them for execution. It exposes both REST and gRPC for future proofing and switch their behavior based on environment variables. For immediately executable tasks, it writes it in the db and routes it to the coordinator in the same transaction, ie, the coordinator does not have to pool for it.

- REST will be used by user facing application.
- gRPC can be used when this application is used as a microservice.

---

## responsibilities of the scheduler :

1. **Ingestion:** Exposing REST and gRPC endpoints to receive new tasks from external clients.
2. **Persistence:** Saving tasks and their states to PostgreSQL.
    - _Example:_ If a client submits a task for immediate execution, the Scheduler writes it to the `tasks` table and the `outbox_events` table in a single ACID transaction.
3. **Time Evaluation:** Periodically querying the database (using `SKIP LOCKED`) to find tasks whose scheduled time has arrived.
4. **Dispatching:** Passing due tasks to our `TaskPublisher` port, which routes them to either RabbitMQ or our Coordinator depending on our environment variable.
5. **State Management:** Updating the database when a task moves from `PENDING` to `DISPATCHED`, and later (via callbacks or message consumption) to `COMPLETED` or `FAILED`.

---

## core domain :

Our core domain holds our business models and the main service logic. It knows nothing about REST, gRPC, PostgreSQL, or RabbitMQ.

```go
// Core Model
type Task struct {
	ID          string     `json:"id"`
	Task        []byte     `json:"task"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	PickedAt    *time.Time `json:"picked_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`
	Error       *string    `json:"error,omitempty"`
}


// Outbound Port for Database Operations
type TaskRepository interface {
    CreateTask(ctx context.Context, task *Task) (*Task, error)
    FetchDueTasks(ctx context.Context, limit int) ([]Task, error)
    MarkAsDispatched(ctx context.Context, taskIDs []uuid.UUID) error
}

// Outbound Port for Dispatching (Defined in our previous discussion)
type TaskPublisher interface {
    Publish(ctx context.Context, task Task) error
}
```

## the adapters :

We then plug our infrastructure into these ports:

- **Inbound Adapters (Driving):** `RestHandler` and `GrpcHandler` that parse requests and call methods on `SchedulerService`.
- **Outbound Adapters (Driven):**
    - `PostgresTaskRepository` (implements `TaskRepository` using `pgx` and our `SKIP LOCKED` SQL query).
    - `RabbitMQPublisher` or `CoordinatorPublisher` (implements `TaskPublisher`).
