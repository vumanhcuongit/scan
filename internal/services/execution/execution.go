package execution

import (
	"context"
	"encoding/json"

	"github.com/vumanhcuongit/scan/pkg/kafka"
	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

type Execution struct {
	kafkaReader *kafka.Reader
	kafkaWriter *kafka.Writer
}

func New(kafkaReader *kafka.Reader, kafkaWriter *kafka.Writer) *Execution {
	return &Execution{
		kafkaReader: kafkaReader,
		kafkaWriter: kafkaWriter,
	}
}

func (e *Execution) Run(ctx context.Context) error {
	zapLogger, _ := zap.NewProduction()
	defer func() {
		_ = zapLogger.Sync()
	}()
	log := zapLogger.Sugar()
	return e.kafkaReader.Consume(ctx, func(ctx context.Context, message []byte) error {
		log.Infof("starting to scan for request: %s", message)
		var req models.ScanRequestMessage
		err := json.Unmarshal(message, &req)
		if err != nil {
			log.Errorf("failed to unmarshal message, err: %+v", err)
			return err
		}
		log.Infof("req is: %+v", req)

		return nil
	})
}
