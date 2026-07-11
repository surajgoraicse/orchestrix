package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/surajgoraicse/orchestrix/libs/go-libs/database"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internals/config"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load environment variables: %v\n", err)
	}
	config := config.NewConfig()
	server := NewSchedulerServer(config)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start scheduler server: %v\n", err)
	}
}

// ScheduleTaskRequest represents the request structure for scheduling a task
type ScheduleTaskRequest struct {
	Task        string `json:"task"`
	ScheduledAt string `json:"scheduled_at"`
}

type ScheduleTaskResponse struct {
	ID          string `json:"id"`
	Task        string `json:"task"`
	ScheduledAt string `json:"scheduled_at"`
}

type Task struct {
	ID          string     `json:"id"`
	Task        string     `json:"task"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	PickedAt    *time.Time `json:"picked_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`
	Error       *string    `json:"error,omitempty"`
}

// TaskStatus represents the status of a task
type TaskStatus struct {
	Task
	Error string `json:"error,omitempty"`
}

type SchedulerServer struct {
	httpServer *http.Server
	dbPool     *pgxpool.Pool
	ctx        context.Context
	cancel     context.CancelFunc
	config     *config.Config
}

func NewSchedulerServer(config *config.Config) *SchedulerServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &SchedulerServer{
		ctx:    ctx,
		config: config,
		cancel: cancel,
	}
}

func (s *SchedulerServer) getDbConfig() *database.DbConfig {
	return &database.DbConfig{
		DBHost:                s.config.DbHost,
		DBPort:                fmt.Sprintf("%d", s.config.DbPort),
		DBUser:                s.config.DbUser,
		DBPassword:            s.config.DbPassword,
		DBName:                s.config.DbName,
		SSLMode:               s.config.SSLMode,
		DBMaxConn:             10,
		DBMinConn:             1,
		DBConnMaxLifetime:     10 * time.Minute,
		DBConnMaxIdleLifetime: 5 * time.Minute,
		DBHealthCheckPeriod:   1 * time.Minute,
		ConnectTimeout:        5 * time.Second,
	}
}

func (s *SchedulerServer) Start() error {
	dbConfig := s.getDbConfig()
	database.NewDatabaseService(dbConfig)

	http.HandleFunc("/schedule", s.handleScheduleTask)
	http.HandleFunc("/status", s.handleGetTaskStatus)
	s.httpServer = &http.Server{
		Addr: fmt.Sprintf(":%d", s.config.ServerPort),
	}

	log.Printf("Scheduler server started on port %d", s.config.ServerPort)

	// start the server in a seperate goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start scheduler server: %v\n", err)
		}
	}()

	// graceful shutdown the scheduler server
	return s.gracefulShutdown()
}

// gracefulShutdown handles the graceful shutdown of the server
func (s *SchedulerServer) gracefulShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("OS signal received, shutting down gracefully...")

	if s.httpServer != nil {
		// create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// shutdown the server
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	if s.dbPool != nil {
		s.dbPool.Close()
	}

	return nil
}

func (s *SchedulerServer) handleScheduleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var taskRequest ScheduleTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&taskRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Println("Received task request: ", taskRequest)

	// Parse the scheduled_at time
	scheduledTime, err := time.Parse(time.RFC3339, taskRequest.ScheduledAt)
	if err != nil {
		http.Error(w, "Invalid scheduled_at time", http.StatusBadRequest)
		return
	}

	// convert the scheduled time to Unix timestamp
	unixTimestamp := time.Unix(scheduledTime.Unix(), 0)

	// insert it into db
	taskId, err := s.insertTaskIntoDb(r.Context(), Task{
		Task:        taskRequest.Task,
		ScheduledAt: &unixTimestamp,
	})
	if err != nil {
		http.Error(w, "Failed to insert task into database", http.StatusInternalServerError)
		return
	}

	// Respond with the scheduled task
	taskResponse := ScheduleTaskResponse{
		ID:          taskId,
		Task:        taskRequest.Task,
		ScheduledAt: taskRequest.ScheduledAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskResponse)
}

func (s *SchedulerServer) handleGetTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	taskStatus, err := s.getTaskFromDB(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taskStatus)

}

func (s *SchedulerServer) insertTaskIntoDb(ctx context.Context, task Task) (string, error) {
	sqlStatement := `
		INSERT INTO tasks (task, scheduled_at) VALUES ($1, $2) RETURNING id
	`
	var taskID string
	err := s.dbPool.QueryRow(ctx, sqlStatement, task.Task, task.ScheduledAt).Scan(&taskID)
	if err != nil {
		return "", err
	}
	return taskID, nil
}

func (s *SchedulerServer) getTaskFromDB(ctx context.Context, taskID string) (TaskStatus, error) {
	sqlStatement := `
		SELECT id, task, scheduled_at, picked_at, started_at, completed_at, failed_at, error FROM tasks WHERE id = $1
	`
	var taskStatus TaskStatus
	err := s.dbPool.QueryRow(ctx, sqlStatement, taskID).Scan(&taskStatus.ID, &taskStatus.Task, &taskStatus.ScheduledAt, &taskStatus.PickedAt, &taskStatus.StartedAt, &taskStatus.CompletedAt, &taskStatus.FailedAt, &taskStatus.Error)
	if err != nil {
		return TaskStatus{}, err
	}
	return taskStatus, nil
}
