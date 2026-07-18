package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

// RestHandlers groups all our REST handlers.
type RestHandlers struct {
	Tasks *TaskHandler
}

// RegisterRoutes registers all the REST routes for the scheduler service.
func (r *RestHandlers) RegisterRoutes(e *echo.Echo) {
	router := e.Group("/api")

	// health check route
	router.GET("/health", healthCheck)

	v1Router := router.Group("/v1")

	// tasks routes
	tasksRouterV1 := v1Router.Group("/tasks")
	r.Tasks.RegisterRoutes(tasksRouterV1)
}

type RestServer struct {
	httpServer *http.Server
}

// Start starts the REST server in a seperate goroutine.
func (s *RestServer) Start() error {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start REST server: %v\n", err)
		}
	}()
	return nil
}

// Shutdown shuts down the REST server.
func (s *RestServer) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// NewRestServer creates a new REST server.
func NewRestServer(port int, handlers *RestHandlers) *RestServer {
	e := echo.New()
	handlers.RegisterRoutes(e)
	return &RestServer{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: e,
		},
	}
}

// healthCheck is a health check endpoint for the REST server.
func healthCheck(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
