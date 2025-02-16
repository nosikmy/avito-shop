package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const readHeaderTimeout = 1 * time.Minute

type Server struct {
	httpServer *http.Server
}

func NewServer(port string, handler http.Handler) (*Server, error) {
	if port == "" {
		return nil, fmt.Errorf("got empty port")
	}
	if handler == nil {
		return nil, fmt.Errorf("got empty handler")
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + port,
			Handler:           handler,
			ReadHeaderTimeout: readHeaderTimeout,
		},
	}, nil
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
