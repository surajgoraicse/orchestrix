package grpc

import (
	"context"
	"net"

	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/modules/tasks"
	"google.golang.org/grpc"
)

type TaskServer struct {
	taskService tasks.ITaskService
}

func NewTaskServer(taskService tasks.ITaskService) *TaskServer {
	return &TaskServer{
		taskService: taskService,
	}
}

type GrpcServer struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func (s *GrpcServer) Start() error {
	return nil
}

func (s *GrpcServer) Shutdown(ctx context.Context) error {
	return nil
}

func NewGrpcServer() *GrpcServer {
	return nil
}
