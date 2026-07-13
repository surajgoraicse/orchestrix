package main

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	coordinatorv1 "github.com/surajgoraicse/orchestrix/api/gen/go/coordinator/v1"
	workerv1 "github.com/surajgoraicse/orchestrix/api/gen/go/worker/v1"
	"github.com/surajgoraicse/orchestrix/services/worker/internals/config"
	"google.golang.org/grpc"
)

type Task struct {
	TaskID      string
	TaskPayload string
}

type WorkerServer struct {
	workerv1.UnimplementedWorkerServiceServer
	id                      uint32
	serverPort              string
	coordinatorAddress      string
	listener                net.Listener
	grpcServer              *grpc.Server
	coordinatorConnection   *grpc.ClientConn
	coordinatorServerClient coordinatorv1.CoordinatorServiceClient
	heartbeatInterval       time.Duration
	taskQueue               chan *Task
	ctx                     context.Context
	cancel                  context.CancelFunc
	wg                      sync.WaitGroup
}

func NewWorkerServer(config *config.Config) *WorkerServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerServer{
		id:                 uuid.New().ID(),
		taskQueue:          make(chan *Task, 100),
		coordinatorAddress: config.CoordinatorAddress,
		heartbeatInterval:  config.HeartbeatInterval,
		ctx:                ctx,
		cancel:             cancel,
		wg:                 sync.WaitGroup{},
	}
}


func main() {

}
