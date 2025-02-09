package main

import (
	"context"
	"errors"
	"fmt"
	"miit-ai-backend/api/kafka"
	"miit-ai-backend/api/minio"
	"miit-ai-backend/api/prediction"
	"miit-ai-backend/api/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

const (
	kafkaBroker       = "KAFKA_BROKER"
	kafkaTopic        = "prediction-requests"
	kafkaResultsTopic = "prediction-results"
	predictionService = "prediction-service"
	minioEndpoint     = "MINIO_ENDPOINT"
	minioAccessKey    = "MINIO_ACCESS_KEY"
	minioSecretKey    = "MINIO_SECRET_KEY"
	minioBucketName   = "MINIO_BUCKET_NAME"
)

func main() {
	kafkaBrokerAddr := getEnvOrDefault(kafkaBroker, "localhost:9092")
	producer := kafka.NewProducer(kafkaBrokerAddr, kafkaTopic)
	consumer := kafka.NewConsumer(kafkaBrokerAddr, kafkaResultsTopic, predictionService)

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

	minioClient, err := minio.NewMinioClient(
		getEnvOrDefault(minioEndpoint, "localhost:9000"),
		getEnvOrDefault(minioAccessKey, "minio"),
		getEnvOrDefault(minioSecretKey, "minio123"),
		getEnvOrDefault(minioBucketName, "images"),
	)
	if err != nil {
		fmt.Printf("Failed to create MinIO client: %v\n", err)
		return
	}

	predictHandler := prediction.NewPredictHandler(producer, consumer, minioClient)
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
