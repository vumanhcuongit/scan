package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vumanhcuongit/scan/internal/repos"
	"github.com/vumanhcuongit/scan/internal/services/base"
	"github.com/vumanhcuongit/scan/pkg/kafka"
	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

type ScanService struct {
	repo repos.IRepo
	base.Service
	kafkaReader *kafka.Reader
	kafkaWriter *kafka.Writer
}

func NewScanService(bs *base.Service, kafkaWriter *kafka.Writer) *ScanService {
	return &ScanService{
		Service:     *bs,
		repo:        bs.Repo(),
		kafkaWriter: kafkaWriter,
	}
}

func (s *ScanService) Start(ctx context.Context) error {
	return s.startConsumer(ctx)
}

// startConsumers starts consumers consuming message from Scan Result topic
func (s *ScanService) startConsumer(ctx context.Context) error {
	log := zap.S()
	fmt.Println("---------------------")
	return s.kafkaReader.Consume(ctx, func(ctx context.Context, message []byte) error {
		log.Infof("starting to consume result message: %s", message)
		var result models.ScanResultMessage
		err := json.Unmarshal(message, &result)
		if err != nil {
			log.Errorf("failed to unmarshal message, err: %+v", err)
			return err
		}
		log.Infof("result is: %+v", result)

		return nil
	})
}
