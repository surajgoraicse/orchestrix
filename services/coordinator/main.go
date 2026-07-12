package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
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
		maxHeartbeatMisses: uint8(config.MaxHeartbeatMisses),
		heartbeatInterval:  config.HeartbeatInterval,
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

func (c *CoordinatorServer) Start() error {
	dbConfig := c.getDBConfig()
	db := database.NewDatabaseService(dbConfig)

	return nil
}

func (s *CoordinatorServer) SendHeartbeat(ctx context.Context, req *coordinatorv1.SendHeartbeatRequest) (*coordinatorv1.SendHeartbeatResponse, error) {
	return &coordinatorv1.SendHeartbeatResponse{}, nil
}

func (s *CoordinatorServer) UpdateTaskStatus(ctx context.Context, req *coordinatorv1.UpdateTaskStatusRequest) (*coordinatorv1.UpdateTaskStatusResponse, error) {
	return &coordinatorv1.UpdateTaskStatusResponse{}, nil
}
