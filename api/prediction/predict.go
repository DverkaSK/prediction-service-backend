package prediction

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"log"
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

	files := r.MultipartForm.File["image"]

	if len(files) == 0 {
		http.Error(w, "Не загружено ни одного файла", http.StatusBadRequest)
		return
	}
	if len(files) > 1 {
		http.Error(w, "Разрешена загрузка только одного файла за раз", http.StatusBadRequest)
		return
	}

	file, err := files[0].Open()
	if err != nil {
		http.Error(w, "Ошибка при открытии файла: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Ошибка при закрытии файла: %v", err)
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		http.Error(w, "Ошибка при чтении файла: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		http.Error(w, "Ошибка при обработке файла: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fileType := http.DetectContentType(buffer)
	if !isAllowedImageType(fileType) {
		http.Error(w, "Неподдерживаемый тип файла. Разрешены только изображения (JPEG, PNG, GIF)", http.StatusBadRequest)
		return
	}

	if files[0].Size > 20<<20 {
		http.Error(w, "Размер файла превышает допустимый предел в 10MB", http.StatusBadRequest)
		return
	}

	imageID, err := h.minioClient.UploadImage(files[0])
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

func isAllowedImageType(fileType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
	}
	return allowedTypes[fileType]
}
