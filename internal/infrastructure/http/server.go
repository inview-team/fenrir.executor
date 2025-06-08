package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/inviewteam/fenrir.executor/internal/application"
	"github.com/inviewteam/fenrir.executor/internal/infrastructure/http/routes"
)

//	@title			Swagger Backend API
//	@version		1.0
//	@description	Backend Server for Competitions
//	@termsOfService	http://swagger.io/terms/

// @host		127.0.0.1:30000
// @BasePath	/api
type Server struct {
	srv http.Server
}

type Config struct {
	Timeout TimeoutConfig `yaml:"timeout,omitempty"`
}

type TimeoutConfig struct {
	Idle  time.Duration `yaml:"idle"`
	Read  time.Duration `yaml:"read"`
	Write time.Duration `yaml:"write"`
}

var (
	DefaultConfig = Config{
		Timeout: TimeoutConfig{
			Idle:  time.Second * 30,
			Read:  time.Second * 30,
			Write: time.Second * 30,
		},
	}
)

func NewServer(app *application.Application) *Server {
	return &Server{
		srv: http.Server{
			Handler:      routes.Make(app),
			Addr:         ":30000",
			IdleTimeout:  DefaultConfig.Timeout.Idle,
			ReadTimeout:  DefaultConfig.Timeout.Read,
			WriteTimeout: DefaultConfig.Timeout.Write,
		},
	}
}

func (s *Server) Start(ctx context.Context) {
	go func() {
		listener := make(chan os.Signal, 1)
		signal.Notify(listener, os.Interrupt, syscall.SIGTERM)
		fmt.Println("Received a shutdown signal:", <-listener)
		// Listen on application shutdown signals.

		// Shutdown HTTP server.
		if err := s.srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Failed to shutdown: %s", err)
		}
	}()

	fmt.Println("Listening on ", s.srv.Addr)
	// Start HTTP server.
	if err := s.srv.ListenAndServe(); err != nil {
		fmt.Printf("Failed to listen and serve: %s", err)
	}
}
