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
			Addr:           ":8080",
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		handler: handler,
	}
}

func (s *Server) Run() error {
	http.HandleFunc("/predict", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		s.handler.Handle(w, r)
	})

	fmt.Println("Server is running on :8080...")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
