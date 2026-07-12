package main

import (
	"context"
	"net"

	"github.com/google/uuid"
	coordinatorv1 "github.com/surajgoraicse/orchestrix/api/gen/go/coordinator/v1"
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
	coordinatorv1.UnimplementedCoordinatorServer
	serverPort int
	listener   net.Listener
	grpcServer *grpc.Server
	WorkerPool map[uuid.UUID]*WorkerNode
}

func NewCoordinatorServer() (*CoordinatorServer, error) {
	return &CoordinatorServer{}, nil
}

func (s *CoordinatorServer) SendHeartbeat(ctx context.Context, req *coordinatorv1.SendHeartbeatRequest) (*coordinatorv1.SendHeartbeatResponse, error) {
	return &coordinatorv1.SendHeartbeatResponse{}, nil
}

func (s *CoordinatorServer) UpdateTaskStatus(ctx context.Context, req *coordinatorv1.UpdateTaskStatusRequest) (*coordinatorv1.UpdateTaskStatusResponse, error) {
	return &pb.UpdateTaskStatusResponse{}, nil
}
