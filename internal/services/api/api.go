package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vumanhcuongit/scan/internal/repos"
	"github.com/vumanhcuongit/scan/internal/services/base"
	"github.com/vumanhcuongit/scan/pkg/kafka"
	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

type ScanService struct {
	repo repos.IRepo
	base.Service
	scanChecker *ScanChecker
	kafkaReader *kafka.Reader
	kafkaWriter *kafka.Writer
}

func NewScanService(bs *base.Service, kafkaWriter *kafka.Writer, kafkaReader *kafka.Reader) *ScanService {
	scanChecker := NewScanChecker(bs.Repo(), &bs.Config().ScanChecker)
	return &ScanService{
		Service:     *bs,
		repo:        bs.Repo(),
		scanChecker: scanChecker,
		kafkaWriter: kafkaWriter,
		kafkaReader: kafkaReader,
	}
}

func (s *ScanService) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(s.Config().ScanChecker.IntervalInMinutes) * time.Minute)
	go func() {
		for range ticker.C {
			_ = s.scanChecker.Check(ctx)
		}
	}()
	return s.startConsumer(ctx)
}

// startConsumers starts consumers consuming message from Scan Result topic
func (s *ScanService) startConsumer(ctx context.Context) error {
	log := zap.S()
	return s.kafkaReader.Consume(ctx, func(ctx context.Context, message []byte) error {
		log.Infof("starting to consume result message: %s", message)
		var result models.ScanResultMessage
		err := json.Unmarshal(message, &result)
		if err != nil {
			log.Errorf("failed to unmarshal message, err: %+v", err)
			return err
		}
		log.Infof("scan's result is: %+v", result)

		err = s.handleResultMessage(ctx, &result)
		if err != nil {
			log.Errorf("failed to handle result's message, err: %+v", err)
			return err
		}
		return nil
	})
}
