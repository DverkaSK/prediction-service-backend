package main

import (
	"context"
	"errors"
	"fmt"
	"miit-ai-backend/api/kafka"
	"miit-ai-backend/api/prediction"
	"miit-ai-backend/api/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	kafkaBroker       = "localhost:9092"
	kafkaTopic        = "prediction-requests"
	kafkaResultsTopic = "prediction-results"
)

func main() {
	producer := kafka.NewProducer(kafkaBroker, kafkaTopic)
	consumer := kafka.NewConsumer(kafkaBroker, kafkaResultsTopic, "prediction-service")

	defer func() {
		if err := producer.Close(); err != nil {
			fmt.Printf("Error closing producer: %v\n", err)
		}
	}()

	defer func() {
		if err := consumer.Close(); err != nil {
			fmt.Printf("Error closing consumer: %v\n", err)
		}
	}()

	predictHandler := prediction.NewPredictHandler(producer, consumer)

	srv := server.NewServer(predictHandler)

	go func() {
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nShutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down server: %v\n", err)
	}

	fmt.Println("Server stopped")
}
