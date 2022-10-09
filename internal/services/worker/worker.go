package worker

import (
	"context"
	"encoding/json"

	"github.com/vumanhcuongit/scan/pkg/kafka"
	"go.uber.org/zap"
)

type Worker struct {
	kafkaReader *kafka.Reader
}

type ScanRequest struct {
	Owner      string `json:"owner"`
	Repository string `json:"repository"`
}

func New(kafkaReader *kafka.Reader) *Worker {
	return &Worker{
		kafkaReader: kafkaReader,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	zapLogger, _ := zap.NewProduction()
	defer func() {
		_ = zapLogger.Sync()
	}()
	log := zapLogger.Sugar()
	return w.kafkaReader.Consume(ctx, func(ctx context.Context, message []byte) error {
		log.Infof("startting to scan for request: %s", message)
		var req ScanRequest
		err := json.Unmarshal(message, &req)
		if err != nil {
			log.Errorf("failed to unmarshal message, err: %+v", err)
			return err
		}
		log.Infof("req is: %+v", req)

		return nil
	})
}
