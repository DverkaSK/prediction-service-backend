package prediction

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"miit-ai-backend/api/kafka"
	"miit-ai-backend/api/minio"
	"miit-ai-backend/api/prediction/structs"
	"net/http"
	"time"
)

type PredictHandler struct {
	producer    *kafka.Producer
	consumer    *kafka.Consumer
	minioClient *minio.Client
}

func NewPredictHandler(producer *kafka.Producer, consumer *kafka.Consumer, minioClient *minio.Client) *PredictHandler {
	return &PredictHandler{
		producer:    producer,
		consumer:    consumer,
		minioClient: minioClient,
	}
}

func (h *PredictHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только метод POST разрешен", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Ошибка при парсинге формы", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Ошибка при получении файла: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageID, err := h.minioClient.UploadImage(header)
	if err != nil {
		http.Error(w, "Ошибка при загрузке изображения в MinIO: "+err.Error(), http.StatusInternalServerError)
		return
	}

	requestID := uuid.New().String()

	predRequest := structs.PredictionRequest{
		ImageID:   imageID,
		RequestID: requestID,
	}

	requestJSON, err := json.Marshal(predRequest)
	if err != nil {
		http.Error(w, "Ошибка при сериализации запроса: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.producer.WriteMessage(context.Background(), []byte(requestID), requestJSON)
	if err != nil {
		http.Error(w, "Ошибка при отправке сообщения в Kafka: "+err.Error(), http.StatusInternalServerError)
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
			http.Error(w, "Ошибка при кодировании ответа: "+err.Error(), http.StatusInternalServerError)
			return
		}
	case err := <-errChan:
		http.Error(w, "Ошибка при чтении результата предсказания: "+err.Error(), http.StatusInternalServerError)
	case <-ctx.Done():
		http.Error(w, "Превышено время ожидания результата", http.StatusGatewayTimeout)
	}
}
