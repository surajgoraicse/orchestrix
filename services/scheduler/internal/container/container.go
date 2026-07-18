package container

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/surajgoraicse/orchestrix/libs/go-libs/database"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/apps/grpc"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/apps/rest"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/config"
	db_sqlc "github.com/surajgoraicse/orchestrix/services/scheduler/internal/db/sqlc"
	"github.com/surajgoraicse/orchestrix/services/scheduler/internal/modules/tasks"
)

// RestHandlers groups all our HTTP endpoints.
type RestHandlers struct {
	Tasks *rest.TaskHandler
}

// GrpcServers groups all our RPC endpoints.
type GrpcHandler struct {
	Tasks *grpc.TaskServer
}

// NewHandlers creates a new set of handlers for the scheduler service.
func NewHandlers(queries *db_sqlc.Queries) (*RestHandlers, *GrpcHandler) {

	// initialize the core services
	taskService := tasks.NewTaskService(queries)

	return &RestHandlers{
			Tasks: rest.NewTaskHandler(taskService),
		}, &GrpcHandler{
			Tasks: grpc.NewTaskServer(taskService),
		}
}

type Container struct {
	Config  *config.Config
	DBPool  *pgxpool.Pool
	Queries *db_sqlc.Queries

	// apps
	Rest *RestHandlers
	Grpc *GrpcHandler
}

// NewContainer creates a new container for the scheduler service.
func NewContainer(ctx context.Context) *Container {
	config := config.NewConfig()

	// initialize the DB
	dbConfig := getDbConfig(config)
	db := database.NewDatabaseService(dbConfig)
	dbPool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	queries := db_sqlc.New(dbPool)

	// initialize the handlers
	restHandlers, grpcHandlers := NewHandlers(queries)

	return &Container{
		Config:  config,
		DBPool:  dbPool,
		Queries: queries,
		Rest:    restHandlers,
		Grpc:    grpcHandlers,
	}
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

// RegisterRoutes registers all the REST routes for the scheduler service.
func (r *RestHandlers) RegisterRoutes(e *echo.Echo) {
	router := e.Group("/api")

	v1Router := router.Group("/v1")

	// tasks routes
	tasksRouterV1 := v1Router.Group("/tasks")
	r.Tasks.RegisterRoutes(tasksRouterV1)
}
