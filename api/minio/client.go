package minio

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"mime/multipart"
	"path/filepath"
)

type Client struct {
	client     *minio.Client
	bucketName string
}

func NewMinioClient(endpoint, accessKey, secretKey, bucketName string) (*Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента minio: %v", err)
	}

	exists, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существования бакета: %v", err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("ошибка создания бакета: %v", err)
		}
	}

	return &Client{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (c *Client) UploadImage(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	imageID := uuid.New().String()

	ext := filepath.Ext(fileHeader.Filename)

	objectName := imageID + ext

	_, err = c.client.PutObject(
		context.Background(),
		"images",
		objectName,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{ContentType: fileHeader.Header.Get("Content-Type")},
	)
	if err != nil {
		return "", err
	}

	return objectName, nil
}
