package kafka

import (
	"context"
	"strings"

	kafka "github.com/segmentio/kafka-go"
)

type Writer struct{ *kafka.Writer }

func NewWriter(brokers string, topic string) *Writer {
	return &Writer{
		&kafka.Writer{
			Addr:                   kafka.TCP(strings.Split(brokers, ",")...),
			Topic:                  topic,
			AllowAutoTopicCreation: true,
		},
	}
}

func (w *Writer) WriteMessage(ctx context.Context, message []byte) error {
	return w.Writer.WriteMessages(ctx, kafka.Message{
		Value: message,
	})
}
