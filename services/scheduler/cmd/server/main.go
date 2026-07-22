package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/apps/grpc"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/apps/rest"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/config"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/container"
	"go.uber.org/zap"
)

type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Failed to load environment variables: %v\n", err)
	}
	// global context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config
	config := config.NewConfig()

	// Load the Dependencies
	di := container.NewContainer(ctx, config)
	defer di.Close()

	// start the server based on the app mode
	var server Server
	switch di.Config.AppMode {
	case "grpc":
		server = grpc.NewGrpcServer()
	case "rest":
		server = rest.NewRestServer(di.Config.RestServerPort, di.Rest)
	default:
		panic("Invalid app mode: " + di.Config.AppMode)
	}

	// start the server in a seperate goroutine
	if err := server.Start(); err != nil {
		di.Logger.Fatal("Failed to start server: ", zap.Error(err))
	}

	// graceful shutdown
	gracefulShutdown(server, cancel)

}

// gracefulShutdown handles the graceful shutdown of the server.
func gracefulShutdown(server Server, globalCancel context.CancelFunc) error {
	// wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// create a context for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	// cancel the global context
	globalCancel()

	return nil
}
