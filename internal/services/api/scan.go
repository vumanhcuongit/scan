package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

type TriggerScanRequest struct {
	RepositoryID int64 `json:"repository_id" binding:"required"`
}

type UpdateScanRequest struct {
	Status     string     `json:"status"`
	Findings   []byte     `json:"findings"`
	QueuedAt   *time.Time `json:"queued_at"`
	ScanningAt *time.Time `json:"scanning_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

type ListScansRequest struct {
	RepositoryID *int64 `json:"repository_id" form:"repository_id"`
	Size         int    `json:"size" form:"size"`
	Page         int    `json:"page" form:"page"`
}

func (s *ScanService) ListScans(ctx context.Context, request *ListScansRequest) ([]*models.Scan, error) {
	log := zap.S()
	log.Infof("starting to list scans with request %+v", request)

	filter := &models.ScanFilter{}
	if request.RepositoryID != nil {
		filter.RepositoryID = request.RepositoryID
	}
	scans, err := s.repo.Scan().List(ctx, request.Size, request.Page, filter)
	if err != nil {
		log.Warnf("failed to list scans, err: %+v", err)
		return nil, err
	}

	return scans, nil
}

func (s *ScanService) TriggerScan(ctx context.Context, request *TriggerScanRequest) (*models.Scan, error) {
	log := zap.S()
	log.Infof("starting to trigger a scan with request %+v", request)

	// first check if this repository exists or not
	repository, err := s.GetRepository(ctx, request.RepositoryID)
	if err != nil {
		log.Warnf("failed to get repository, err: %+v", err)
		return nil, err
	}

	scan, err := s.createScan(ctx, repository)
	if err != nil {
		log.Warnf("failed to create scan, err: %+v", err)
		return nil, err
	}

	err = s.produceTriggerScanMessage(ctx, scan, repository)
	if err != nil {
		log.Warnf("failed to write message to queue, err: %+v", err)
		return nil, err
	}

	updatedScan, err := s.updateQueuedScan(ctx, scan)
	if err != nil {
		log.Warnf("failed to update scan, err: %+v", err)
		return nil, err
	}

	return updatedScan, nil
}

func (s *ScanService) UpdateScan(ctx context.Context, scan *models.Scan, request *UpdateScanRequest) (*models.Scan, error) {
	log := zap.S()
	log.Infof("starting to update repository with request %+v", request)

	changesets := map[string]interface{}{}
	if request.Status != "" {
		changesets["status"] = request.Status
		scan.Status = request.Status
	}
	if request.Findings != nil {
		changesets["findings"] = request.Findings
		scan.Findings = request.Findings
	}
	if request.QueuedAt != nil {
		changesets["queued_at"] = request.QueuedAt
		scan.QueuedAt = request.QueuedAt
	}
	if request.ScanningAt != nil {
		changesets["scanning_at"] = request.ScanningAt
		scan.ScanningAt = request.ScanningAt
	}
	if request.FinishedAt != nil {
		changesets["finished_at"] = request.FinishedAt
		scan.FinishedAt = request.FinishedAt
	}

	err := s.repo.Scan().UpdateWithMap(ctx, scan, changesets)
	if err != nil {
		log.Warnf("failed to update scan, err: +%v", err)
		return nil, err
	}

	return scan, nil
}

func (s *ScanService) createScan(ctx context.Context, repository *models.Repository) (*models.Scan, error) {
	log := zap.S()
	record, err := models.NewScan(repository)
	if err != nil {
		log.Warnf("failed to init scan, err: %+v", err)
		return nil, err
	}

	scan, err := s.repo.Scan().Create(ctx, record)
	if err != nil {
		log.Warnf("failed to create scan, err: %+v", err)
		return nil, err
	}

	return scan, nil
}

func (s *ScanService) produceTriggerScanMessage(ctx context.Context, scan *models.Scan, repository *models.Repository) error {
	log := zap.S()

	message, err := json.Marshal(models.ScanRequestMessage{
		ScanID:     scan.ID,
		Owner:      repository.Owner,
		Repository: repository.Name,
	})
	if err != nil {
		log.Warnf("failed to marshal message, err: %+v", err)
		return err
	}

	err = s.kafkaWriter.WriteMessage(ctx, message)
	if err != nil {
		log.Warnf("failed to write message to queue, err: %+v", err)
		return err
	}

	return nil
}

func (s *ScanService) updateQueuedScan(ctx context.Context, scan *models.Scan) (*models.Scan, error) {
	log := zap.S()
	timeNow := time.Now()
	updateScanRequest := &UpdateScanRequest{Status: models.ScanStatusQueued, QueuedAt: &timeNow}
	updatedScan, err := s.UpdateScan(ctx, scan, updateScanRequest)
	if err != nil {
		log.Warnf("failed to update scan, err: %+v", err)
		return nil, err
	}

	return updatedScan, nil
}

// HandleResultMessage handles result returned from workers
func (s *ScanService) HandleResultMessage(ctx context.Context, result *models.ScanResultMessage) error {
	log := zap.S()
	scan := &models.Scan{ID: result.ScanID}
	updateScanRequest := &UpdateScanRequest{
		Status: result.ScanStatus,
	}
	switch result.ScanStatus {
	case models.ScanStatusInProgress:
		updateScanRequest.ScanningAt = result.ScanningAt
	case models.ScanStatusSuccess:
		updateScanRequest.FinishedAt = result.FinishedAt
		updateScanRequest.Findings = result.Findings
	case models.ScanStatusFailure:
		updateScanRequest.FinishedAt = result.FinishedAt
	default:
		log.Warnf("unsupported status")
		return nil
	}

	updatedScan, err := s.UpdateScan(ctx, scan, updateScanRequest)
	if err != nil {
		log.Warnf("failed to update scan, err: %+v", err)
		return err
	}
	log.Infof("updated scan: %+v", updatedScan)

	return nil
}
