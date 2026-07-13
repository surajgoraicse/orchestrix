package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	coordinatorv1 "github.com/surajgoraicse/orchestrix/api/gen/go/coordinator/v1"
	workerv1 "github.com/surajgoraicse/orchestrix/api/gen/go/worker/v1"
	"github.com/surajgoraicse/orchestrix/services/worker/internals/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (w *WorkerServer) Start() error {
	w.startWorkerPool(3)
	if err := w.connectToCoordinator(); err != nil {
		return err
	}
	if err := w.startGRPCServer(); err != nil {
		return err
	}
	go w.sendHeartbeat()

	return w.gracefulShutdown()
}

func (w *WorkerServer) gracefulShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down gracefully...")
	w.cancel()
	w.wg.Wait()
	if w.grpcServer != nil {
		w.grpcServer.GracefulStop()
	}
	if w.coordinatorConnection != nil {
		w.coordinatorConnection.Close()
	}
	return nil
}

func (w *WorkerServer) startGRPCServer() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	w.serverPort = fmt.Sprintf(":%d", w.listener.Addr().(*net.TCPAddr).Port)
	w.listener = listener
	w.grpcServer = grpc.NewServer()
	workerv1.RegisterWorkerServiceServer(w.grpcServer, w)
	go func() {
		if err := w.grpcServer.Serve(w.listener); err != nil {
			log.Fatalln("Error starting gRPC server:", err)
		}
	}()
	return nil
}

func (w *WorkerServer) sendHeartbeat() {
	w.wg.Add(1)
	defer w.wg.Done()

	ticker := time.NewTicker(w.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.sendHeartbeatToCoordinator(); err != nil {
				log.Println("Error sending heartbeat:", err)
			}
		}
	}
}
func (w *WorkerServer) sendHeartbeatToCoordinator() error {
	_, err := w.coordinatorServerClient.SendHeartbeat(w.ctx, &coordinatorv1.SendHeartbeatRequest{
		WorkerId:      fmt.Sprintf("%d", w.id),
		WorkerAddress: w.serverPort,
	})
	if err != nil {
		return err
	}
	return nil
}

func (w *WorkerServer) connectToCoordinator() error {
	conn, err := grpc.NewClient(w.coordinatorAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	w.coordinatorConnection = conn
	w.coordinatorServerClient = coordinatorv1.NewCoordinatorServiceClient(conn)
	log.Println("connected to coordinator")
	return nil
}

func (w *WorkerServer) startWorkerPool(workerPoolSize int) {
	for i := 0; i < workerPoolSize; i++ {
		w.wg.Add(1)
		go w.worker()
	}
}

func (w *WorkerServer) worker() {
	defer w.wg.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		case task := <-w.taskQueue:
			err := w.UpdateTaskStatus(task, coordinatorv1.TaskStatus_TASK_STARTED, "")
			if err != nil {
				log.Println("Error updating task status:", err)
			}
			w.processTask(task)
			err = w.UpdateTaskStatus(task, coordinatorv1.TaskStatus_TASK_COMPLETED, "")
			if err != nil {
				log.Println("Error updating task status:", err)
			}
		}
	}
}

func (w *WorkerServer) UpdateTaskStatus(task *Task, status coordinatorv1.TaskStatus, taskError string) error {
	_, err := w.coordinatorServerClient.UpdateTaskStatus(w.ctx, &coordinatorv1.UpdateTaskStatusRequest{
		TaskId:      task.TaskID,
		Status:      status,
		StartedAt:   time.Now().Unix(),
		CompletedAt: time.Now().Unix(),
		FailedAt:    time.Now().Unix(),
		Error:       taskError,
	})
	if err != nil {
		return err
	}
	return nil
}

func (w *WorkerServer) processTask(task *Task) {
	log.Println("Processing task:", task.TaskID, task.TaskPayload)
	time.Sleep(2 * time.Second)
	log.Println("Task completed:", task.TaskID, task.TaskPayload)
}

func (w *WorkerServer) SubmitTask(ctx context.Context, req *workerv1.SubmitTaskRequest) (*workerv1.SubmitTaskResponse, error) {
	log.Println("Received task:", req.TaskId, req.TaskPayload)
	w.taskQueue <- &Task{
		TaskID:      req.TaskId,
		TaskPayload: req.TaskPayload,
	}
	return &workerv1.SubmitTaskResponse{
		TaskId:  req.TaskId,
		Success: true,
	}, nil
}

func main() {

}
