package server

import (
	"context"
	"fmt"
	"miit-ai-backend/api/prediction"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	handler    *prediction.PredictHandler
}

func NewServer(handler *prediction.PredictHandler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":8080",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		handler: handler,
	}
}

func (s *Server) Run() error {
	http.HandleFunc("/predict", s.handler.Handle)

	fmt.Println("Server is running on :8080...")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
