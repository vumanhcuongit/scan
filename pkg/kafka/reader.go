package kafka

import (
	"context"
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

func (r *Reader) Consume(ctx context.Context, fn ReaderFunc) error {
	log := ctxzap.Extract(ctx).Sugar()
	for {
		m, err := r.Reader.ReadMessage(ctx)
		if err != nil {
			log.Errorf("failed to consume message from Kafka, err: %+v", err)
			return err
		}
		err = fn(ctx, m.Value)
		if err != nil {
			log.Errorf("failed to process message: %+v", err)
		}
	}
}
