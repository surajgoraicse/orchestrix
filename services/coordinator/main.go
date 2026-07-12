package main

import (
	"context"
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
	"github.com/jackc/pgx/v5/pgxpool"
	coordinatorv1 "github.com/surajgoraicse/orchestrix/api/gen/go/coordinator/v1"
	workerv1 "github.com/surajgoraicse/orchestrix/api/gen/go/worker/v1"
	"github.com/surajgoraicse/orchestrix/libs/go-libs/database"
	"github.com/surajgoraicse/orchestrix/services/coordinator/internals/config"
	"google.golang.org/grpc"
)

func main() {

}

type WorkerNode struct {
	heartbeatMisses     uint8
	address             string
	grpcConnection      *grpc.ClientConn
	workerServiceClient workerv1.WorkerServiceClient
}

type CoordinatorServer struct {
	coordinatorv1.UnimplementedCoordinatorServiceServer
	listener           net.Listener
	grpcServer         *grpc.Server
	WorkerPool         map[uuid.UUID]*WorkerNode
	WorkerPoolMutex    sync.RWMutex
	dbScanInterval     time.Duration
	maxHeartbeatMisses uint8
	heartbeatInterval  time.Duration
	roundRobinIndex    atomic.Int32
	config             *config.Config
	dbPool             *pgxpool.Pool
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
}

func NewCoordinatorServer(config *config.Config) (*CoordinatorServer, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &CoordinatorServer{
		config:             config,
		WorkerPool:         make(map[uuid.UUID]*WorkerNode),
		WorkerPoolMutex:    sync.RWMutex{},
		dbScanInterval:     config.DbScanInterval,
		maxHeartbeatMisses: uint8(config.MaxHeartbeatMisses),
		heartbeatInterval:  config.HeartbeatInterval,
		roundRobinIndex:    atomic.Int32{},
		ctx:                ctx,
		cancel:             cancel,
		wg:                 sync.WaitGroup{},
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

	c.scanDatabase()

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

func (c *CoordinatorServer) scanDatabase() {
	ticker := time.NewTicker(c.dbScanInterval)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.executeAllScheduledTasks()
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

func (c *CoordinatorServer) executeAllScheduledTasks() {

}

func (s *CoordinatorServer) SendHeartbeat(ctx context.Context, req *coordinatorv1.SendHeartbeatRequest) (*coordinatorv1.SendHeartbeatResponse, error) {
	return &coordinatorv1.SendHeartbeatResponse{}, nil
}

func (s *CoordinatorServer) UpdateTaskStatus(ctx context.Context, req *coordinatorv1.UpdateTaskStatusRequest) (*coordinatorv1.UpdateTaskStatusResponse, error) {
	return &coordinatorv1.UpdateTaskStatusResponse{}, nil
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
