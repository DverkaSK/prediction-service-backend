package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"time"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(broker, topic, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          topic,
		GroupID:        groupID,
		StartOffset:    kafka.FirstOffset,
		MaxWait:        500 * time.Millisecond,
		CommitInterval: time.Second,
	})

	return &Consumer{reader: reader}
}

func (c *Consumer) ReadMessage(context context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(context)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
