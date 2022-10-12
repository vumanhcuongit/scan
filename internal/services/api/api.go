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

//go:generate mockgen -source=api.go -destination=iscan_service.mock.go -package=api
type IScanService interface {
	Start(ctx context.Context) error

	// repositories
	CreateRepository(ctx context.Context, request *CreateRepositoryRequest) (*models.Repository, error)
	GetRepository(ctx context.Context, repositoryID int64) (*models.Repository, error)
	ListRepositories(ctx context.Context, request *ListRepositoriesRequest) ([]*models.Repository, error)
	UpdateRepository(ctx context.Context, repositoryID int64, request *UpdateRepositoryRequest) (*models.Repository, error)
	DeleteRepository(ctx context.Context, repositoryID int64) error

	// scan
	ListScans(ctx context.Context, request *ListScansRequest) ([]*models.Scan, error)
	TriggerScan(ctx context.Context, request *TriggerScanRequest) (*models.Scan, error)
	UpdateScan(ctx context.Context, scan *models.Scan, request *UpdateScanRequest) (*models.Scan, error)
	HandleResultMessage(ctx context.Context, result *models.ScanResultMessage) error
}

type ScanService struct {
	repo repos.IRepo
	base.Service
	scanChecker *ScanChecker
	kafkaReader *kafka.Reader
	kafkaWriter kafka.IWriter
}

func NewScanService(bs *base.Service, kafkaWriter kafka.IWriter, kafkaReader *kafka.Reader) IScanService {
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

		err = s.HandleResultMessage(ctx, &result)
		if err != nil {
			log.Errorf("failed to handle result's message, err: %+v", err)
			return err
		}
		return nil
	})
}
