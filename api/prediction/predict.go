package prediction

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"miit-ai-backend/api/kafka"
	"miit-ai-backend/api/prediction/structs"
	"net/http"
	"time"
)

type PredictHandler struct {
	producer *kafka.Producer
	consumer *kafka.Consumer
}

func NewPredictHandler(producer *kafka.Producer, consumer *kafka.Consumer) *PredictHandler {
	return &PredictHandler{
		producer: producer,
		consumer: consumer,
	}
}

func (h *PredictHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	requestID := uuid.New().String()

	var request structs.PredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	request.RequestID = requestID

	requestJSON, err := json.Marshal(request)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	err = h.producer.WriteMessage(context.Background(), []byte(requestID), requestJSON)
	if err != nil {
		http.Error(w, "Failed to send message to Kafka", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultChan := make(chan structs.PredictionResponse, 1)
	errChan := make(chan error, 1)

	go func() {
		for {
			message, err := h.consumer.ReadMessage(ctx)
			if err != nil {
				errChan <- err
				return
			}

			var response structs.PredictionResponse
			if err := json.Unmarshal(message.Value, &response); err != nil {
				continue
			}

			if response.RequestID == requestID {
				resultChan <- response
				return
			}
		}
	}()

	select {
	case response := <-resultChan:
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	case err := <-errChan:
		http.Error(w, "Error reading prediction result: "+err.Error(), http.StatusInternalServerError)
	case <-ctx.Done():
		http.Error(w, "Timeout waiting for prediction result", http.StatusGatewayTimeout)
	}
}
