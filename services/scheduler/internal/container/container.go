package container

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surajgoraicse/orchestrix/libs/go-libs/database"
	"github.com/surajgoraicse/orchestrix/libs/go-libs/logger"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/apps/grpc"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/apps/rest"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/config"
	db_sqlc "github.com/surajgoraicse/orchestrix/services/scheduler/internal/db/sqlc"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/modules/tasks"
	"go.uber.org/zap"
)

// GrpcServers groups all our RPC endpoints.
type GrpcHandler struct {
	Tasks *grpc.TaskServer
}

// NewHandlers creates a new set of handlers for the scheduler service.
func NewHandlers(queries *db_sqlc.Queries, logger *zap.Logger) (*rest.RestHandlers, *GrpcHandler) {
	// initialize the core services
	taskService := tasks.NewTaskService(queries, logger)

	return &rest.RestHandlers{
			Tasks: rest.NewTaskHandler(taskService),
		}, &GrpcHandler{
			Tasks: grpc.NewTaskServer(taskService),
		}
}

type Container struct {
	Config  *config.Config
	DBPool  *pgxpool.Pool
	Queries *db_sqlc.Queries
	Logger  *zap.Logger

	// apps
	Rest *rest.RestHandlers
	Grpc *GrpcHandler
}

// NewContainer creates a new container for the scheduler service.
func NewContainer(ctx context.Context, config *config.Config) *Container {

	// initialize the logger
	logger, err := logger.InitLogger(config.ServiceName, config.AppMode)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v\n", err)
	}

	// initialize the DB
	dbConfig := getDbConfig(config)
	db := database.NewDatabaseService(dbConfig)
	dbPool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	queries := db_sqlc.New(dbPool)

	// initialize the handlers
	restHandlers, grpcHandlers := NewHandlers(queries, logger)

	return &Container{
		Config:  config,
		Logger:  logger,
		DBPool:  dbPool,
		Queries: queries,
		Rest:    restHandlers,
		Grpc:    grpcHandlers,
	}
}
func (c *Container) Close() {
	c.DBPool.Close()
	logger.Flush()
}

// getDbConfig returns the database configuration for the scheduler service.
func getDbConfig(config *config.Config) *database.DbConfig {
	return &database.DbConfig{
		DBHost:                config.DbHost,
		DBPort:                fmt.Sprintf("%d", config.DbPort),
		DBUser:                config.DbUser,
		DBPassword:            config.DbPassword,
		DBName:                config.DbName,
		DBSchema:              config.DbSchema,
		SSLMode:               config.SSLMode,
		DBMaxConn:             10,
		DBMinConn:             1,
		DBConnMaxLifetime:     10 * time.Minute,
		DBConnMaxIdleLifetime: 5 * time.Minute,
		DBHealthCheckPeriod:   1 * time.Minute,
		ConnectTimeout:        5 * time.Second,
	}
}
