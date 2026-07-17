package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	coordinatorv1 "github.com/surajgoraicse/orchestrix/api/gen/go/coordinator/v1"
	workerv1 "github.com/surajgoraicse/orchestrix/api/gen/go/worker/v1"
	"github.com/surajgoraicse/orchestrix/libs/go-libs/database"
	"github.com/surajgoraicse/orchestrix/services/coordinator/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrNoAvailableWorkers = errors.New("no available workers")
)

func main() {
	config := config.NewConfig()
	coordinatorServer, err := NewCoordinatorServer(config)
	if err != nil {
		log.Fatalf("failed to create coordinator server: %v", err)
	}
	if err := coordinatorServer.Start(); err != nil {
		log.Fatalf("failed to start coordinator server: %v", err)
	}
}

type WorkerNode struct {
	id                  uuid.UUID
	lastHeartbeatAt     time.Time
	address             string
	grpcConnection      *grpc.ClientConn
	workerServiceClient workerv1.WorkerServiceClient
}

type CoordinatorServer struct {
	coordinatorv1.UnimplementedCoordinatorServiceServer
	listener          net.Listener
	grpcServer        *grpc.Server
	WorkerPool        []*WorkerNode
	WorkerPoolMutex   sync.RWMutex
	dbScanInterval    time.Duration
	heartbeatInterval time.Duration
	roundRobinIndex   atomic.Uint32
	config            *config.Config
	dbPool            *pgxpool.Pool
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

func NewCoordinatorServer(config *config.Config) (*CoordinatorServer, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &CoordinatorServer{
		config:            config,
		WorkerPool:        make([]*WorkerNode, 0),
		WorkerPoolMutex:   sync.RWMutex{},
		dbScanInterval:    config.DbScanInterval,
		heartbeatInterval: config.HeartbeatInterval,
		roundRobinIndex:   atomic.Uint32{},
		ctx:               ctx,
		cancel:            cancel,
		wg:                sync.WaitGroup{},
	}, nil
}

func (c *CoordinatorServer) getDBConfig() *database.DbConfig {
	return &database.DbConfig{
		DBHost:                c.config.DbHost,
		DBPort:                fmt.Sprintf("%d", c.config.DbPort),
		DBUser:                c.config.DbUser,
		DBPassword:            c.config.DbPassword,
		DBName:                c.config.DbName,
		SSLMode:               c.config.SSLMode,
		DBMaxConn:             10,
		DBMinConn:             1,
		DBConnMaxLifetime:     10 * time.Minute,
		DBConnMaxIdleLifetime: 5 * time.Minute,
		DBHealthCheckPeriod:   1 * time.Minute,
		ConnectTimeout:        5 * time.Second,
	}
}

// Start starts the coordinator server
func (c *CoordinatorServer) Start() error {
	dbConfig := c.getDBConfig()
	db := database.NewDatabaseService(dbConfig)
	var err error
	c.dbPool, err = db.Connect(c.ctx)
	if err != nil {
		return err
	}

	log.Println("starting the gRPC server on port ", c.config.ServerPort)
	err = c.startGrpcServer()
	if err != nil {
		return fmt.Errorf("failed to start gRPC server: %v", err)
	}
	c.wg.Go(c.manageWorkerPool)
	c.wg.Go(c.scanDatabase)

	return c.gracefulShutdown()
}

// startGrpcServer starts the gRPC server on the configured port
func (c *CoordinatorServer) startGrpcServer() error {
	var err error
	c.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", c.config.ServerPort))
	if err != nil {
		return err
	}
	c.grpcServer = grpc.NewServer()
	coordinatorv1.RegisterCoordinatorServiceServer(c.grpcServer, c)

	// start the gRPC server in a goroutine
	go func() {
		if err := c.grpcServer.Serve(c.listener); err != nil {
			log.Fatalf("error starting the gRPC server %v", err)
		}
	}()
	return nil
}

// scanDatabase scans the database for scheduled tasks and executes them
func (c *CoordinatorServer) scanDatabase() {
	ticker := time.NewTicker(c.dbScanInterval)
	c.wg.Go(
		func() {
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					func() {
						// we could run this asynchronously in a seperate goroutine
						// that would require locking the db rows to prevent duplicate execution
						// for now we keep it synchronously with 10 seconds timeout
						log.Println("scanning database for scheduled tasks")
						ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
						defer cancel()
						c.executeAllScheduledTasks(ctx)
					}()
				case <-c.ctx.Done():
					return
				}
			}
		},
	)
}

// executeAllScheduledTasks executes all scheduled tasks from the database
func (c *CoordinatorServer) executeAllScheduledTasks(ctx context.Context) {
	done := make(chan struct{})

	go func() {
		// perform the actual work that might get hung up
		defer close(done)

		tx, err := c.dbPool.Begin(ctx)
		if err != nil {
			log.Printf("failed to begin transaction: %v", err)
			return
		}
		// rollback the transaction if it is not committed
		defer func() {
			err := tx.Rollback(ctx)
			if errors.Is(err, pgx.ErrTxClosed) {
				return
			}
			if err != nil {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}()

		rows, err := tx.Query(ctx, `SELECT id, task FROM scheduler.tasks WHERE scheduled_at < (NOW() + INTERVAL '30 seconds') AND picked_at IS NULL ORDER BY scheduled_at FOR UPDATE SKIP LOCKED`)
		if err != nil {
			log.Printf("Error executing query: %v\n", err)
			return
		}
		defer rows.Close()

		var tasks []*workerv1.SubmitTaskRequest
		for rows.Next() {
			var task workerv1.SubmitTaskRequest
			if err := rows.Scan(&task.TaskId, &task.TaskPayload); err != nil {
				log.Printf("Error scanning row: %v\n", err)
				continue
			}
			tasks = append(tasks, &task)
		}
		if err = rows.Err(); err != nil {
			log.Printf("Error scanning rows: %v\n", err)
			return
		}
		for _, task := range tasks {
			if _, err := c.submitTaskToWorker(ctx, task); err != nil {
				log.Printf("Error submitting task: %v\n", err)
				continue
			}
			if _, err := tx.Exec(ctx, `UPDATE scheduler.tasks SET picked_at = NOW() WHERE id = $1`, task.TaskId); err != nil {
				log.Printf("Error updating task status to picked_at: %v\n", err)
				continue
			}
		}
		if err := tx.Commit(ctx); err != nil {
			log.Printf("Error committing transaction: %v\n", err)
			return
		}

	}()

	select {
	case <-done:
		// Completed successfully
	case <-ctx.Done():
		// Timed out or Parent context cancelled.
		// Note: The goroutine above might still be leaked/running in the background
		// until the process exit, but this function exits immediately.
		log.Printf("executeAllScheduledTasks: task execution timed out or was cancelled")
	}
}

// removeInactiveWorkers removes inactive workers from the pool
func (c *CoordinatorServer) removeInactiveWorkers() {
	c.WorkerPoolMutex.Lock()
	defer c.WorkerPoolMutex.Unlock()
	for i, worker := range c.WorkerPool {
		if time.Since(worker.lastHeartbeatAt) > c.config.DeadWorkerTTL {
			log.Printf("Removing inactive worker: %s", worker.address)
			c.WorkerPool = append(c.WorkerPool[:i], c.WorkerPool[i+1:]...)
		}
	}
}

// manageWorkerPool manages the worker pool
// it removes inactive workers from the pool in a separate goroutine
func (c *CoordinatorServer) manageWorkerPool() {
	ticker := time.NewTicker(c.heartbeatInterval * 2)
	defer ticker.Stop()
	select {
	case <-ticker.C:
		c.removeInactiveWorkers()
	case <-c.ctx.Done():
		return
	}

}

// getNextWorker returns the next available worker
// it uses round robin algorithm to select the next worker
func (c *CoordinatorServer) getNextWorker() (*WorkerNode, error) {
	c.WorkerPoolMutex.RLock()
	defer c.WorkerPoolMutex.RUnlock()
	nosOfWorkers := len(c.WorkerPool)
	if nosOfWorkers == 0 {
		return nil, ErrNoAvailableWorkers
	}
	i := c.roundRobinIndex.Load() % uint32(nosOfWorkers)
	c.roundRobinIndex.Add(1)
	return c.WorkerPool[i], nil
}

// submitTaskToWorker submits a task to a worker
func (c *CoordinatorServer) submitTaskToWorker(ctx context.Context, task *workerv1.SubmitTaskRequest) (*workerv1.SubmitTaskResponse, error) {
	worker, err := c.getNextWorker()
	if err != nil {
		return nil, err
	}
	res, err := worker.workerServiceClient.SubmitTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to submit task to worker %s: %v", worker.address, err)
	}
	// log.Printf("Task %s submitted to worker %s with response %v", task.TaskId, worker.address, res)
	return res, nil
}

func (c *CoordinatorServer) SendHeartbeat(ctx context.Context, req *coordinatorv1.SendHeartbeatRequest) (*coordinatorv1.SendHeartbeatResponse, error) {
	c.WorkerPoolMutex.Lock()
	defer c.WorkerPoolMutex.Unlock()

	reqWorkerId := req.GetWorkerId()
	reqWorkerUUID, err := uuid.Parse(reqWorkerId)
	if err != nil {
		return nil, fmt.Errorf("invalid worker id: %v", err)
	}

	for _, worker := range c.WorkerPool {
		if worker.id == reqWorkerUUID {
			worker.lastHeartbeatAt = time.Now()
			return &coordinatorv1.SendHeartbeatResponse{
				Ack:     true,
				Message: "successfully received heartbeat",
			}, nil
		}
	}

	log.Println("Registering a new worker ", req.GetWorkerAddress())
	conn, err := grpc.NewClient(req.GetWorkerAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial worker %s: %v", req.GetWorkerAddress(), err)
	}

	c.WorkerPool = append(c.WorkerPool, &WorkerNode{
		id:                  reqWorkerUUID,
		address:             req.GetWorkerAddress(),
		lastHeartbeatAt:     time.Now(),
		workerServiceClient: workerv1.NewWorkerServiceClient(conn),
	})

	return &coordinatorv1.SendHeartbeatResponse{
		Ack:     true,
		Message: "successfully registered worker",
	}, nil
}

func (c *CoordinatorServer) UpdateTaskStatus(ctx context.Context, req *coordinatorv1.UpdateTaskStatusRequest) (*coordinatorv1.UpdateTaskStatusResponse, error) {
	status := req.Status
	taskID := req.TaskId
	taskError := req.Error
	var timeStamp pgtype.Timestamp

	var column = ""

	switch status {
	case coordinatorv1.TaskStatus_TASK_STARTED:
		column = "started_at"
		timeStamp.Time = time.Unix(req.StartedAt, 0)
		timeStamp.Valid = true

	case coordinatorv1.TaskStatus_TASK_COMPLETED:
		column = "completed_at"
		timeStamp.Time = time.Unix(req.CompletedAt, 0)
		timeStamp.Valid = true

	case coordinatorv1.TaskStatus_TASK_FAILED:
		column = "failed_at"
		timeStamp.Time = time.Unix(req.FailedAt, 0)
		timeStamp.Valid = true

	default:
		return &coordinatorv1.UpdateTaskStatusResponse{
			Success: false,
		}, fmt.Errorf("invalid task status: %v", status)
	}
	sqlStatement := fmt.Sprintf("UPDATE scheduler.tasks SET %s = $1, error = $2 WHERE id = $3", column)
	_, err := c.dbPool.Exec(ctx, sqlStatement, timeStamp, taskError, taskID)
	if err != nil {
		return &coordinatorv1.UpdateTaskStatusResponse{
			Success: false,
		}, err
	}
	return &coordinatorv1.UpdateTaskStatusResponse{
		Success: true,
	}, nil
}

// gracefulShutdown handles the graceful shutdown of the server
func (c *CoordinatorServer) gracefulShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("OS signal received, shutting down gracefully...")

	// 1. Shutdown the gRPC server (stops incoming traffic)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var shutdownWg sync.WaitGroup
	shutdownWg.Add(1)

	go func() {
		defer shutdownWg.Done()
		log.Println("Shutting down gRPC server")

		done := make(chan struct{})
		go func() {
			c.grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			log.Println("gRPC server closed successfully")
		case <-shutdownCtx.Done():
			log.Println("gRPC graceful stop timed out, forcing stop")
			c.grpcServer.Stop()
		}
	}()

	shutdownWg.Wait()

	// 2. Cancel the main context (signals background loop to stop)
	log.Println("Canceling main context")
	c.cancel()

	// 3. Wait for all background goroutines (like scanDatabase) to exit
	log.Println("Waiting for background workers to finish")
	c.wg.Wait()

	// 4. Safely close database connection pool
	log.Println("Closing database connection")
	c.dbPool.Close()

	log.Println("Server shut down successfully")
	return nil
}
