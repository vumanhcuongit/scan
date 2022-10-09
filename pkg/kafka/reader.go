package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	kafka "github.com/segmentio/kafka-go"
)

type Reader struct{ *kafka.Reader }

type ReaderFunc func(ctx context.Context, message []byte) error

func NewReader(brokers string, topic string, groupId string) *Reader {
	return &Reader{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: groupId,
			Topic:   topic,
		}),
	}
}

func (w *Reader) Consume(ctx context.Context, fn ReaderFunc) error {
	log := ctxzap.Extract(ctx).Sugar()
	for {
		m, err := w.Reader.ReadMessage(ctx)
		if err != nil {
			log.Errorf("failed to consume message from Kafka, err: %+v", err)
			return fmt.Errorf("failed to consume message, err: %+v", err)
		}
		err = fn(ctx, m.Value)
		if err != nil {
			log.Infof("failed to process message: %+v", err)
		}
	}
}
