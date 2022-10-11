package api

import (
	"context"

	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/internal/repos"
	"go.uber.org/zap"
)

type ScanChecker struct {
	repo repos.IRepo
	cfg  *config.ScanCheckerConfig
}

func NewScanChecker(repo repos.IRepo, cfg *config.ScanCheckerConfig) *ScanChecker {
	return &ScanChecker{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *ScanChecker) Check(ctx context.Context) error {
	log := zap.S()
	log.Info("starting to check if there are stale scans")
	err := s.repo.Scan().MarkStaleScansAsFailure(ctx, s.cfg.MaxStaleTimeInMinutes)
	if err != nil {
		log.Warnf("failed to mark stale scans as failure, err: %+v", err)
		return err
	}
	log.Info("finish the check")
	return nil
}
