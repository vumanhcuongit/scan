package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/vumanhcuongit/scan/internal/repos"
	"github.com/vumanhcuongit/scan/pkg/kafka"
	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

type scanSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	repo           *repos.MockIRepo
	scanRepo       *repos.MockIScanRepo
	repositoryRepo *repos.MockIRepositoryRepo
	kafkaWriter    *kafka.MockIWriter
	scanService    *ScanService
}

func TestScanSuite(t *testing.T) {
	suite.Run(t, &scanSuite{})
}

func (s *scanSuite) SetupSuite() {
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()
	undo := zap.ReplaceGlobals(logger)
	defer undo()
}

func (s *scanSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.repo = repos.NewMockIRepo(s.mockCtrl)
	s.scanRepo = repos.NewMockIScanRepo(s.mockCtrl)
	s.kafkaWriter = kafka.NewMockIWriter(s.mockCtrl)
	s.scanService = &ScanService{repo: s.repo, kafkaWriter: s.kafkaWriter}
	s.repositoryRepo = repos.NewMockIRepositoryRepo(s.mockCtrl)
}

func (s *scanSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *scanSuite) TestListScans() {
	repoID := int64(1)
	repoName := "scan"
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	scanID := int64(2)
	expectedScan := &models.Scan{
		ID:             scanID,
		RepositoryID:   repoID,
		RepositoryName: repoName,
		RepositoryURL:  repositoryURL,
	}
	request := &ListScansRequest{
		Page:         1,
		Size:         20,
		RepositoryID: &repoID,
	}

	s.scanRepo.EXPECT().List(gomock.Any(), request.Size, request.Page, gomock.Any()).
		Return([]*models.Scan{expectedScan}, nil)
	s.repo.EXPECT().Scan().Return(s.scanRepo)

	scans, err := s.scanService.ListScans(context.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(1, len(scans))
	s.Require().Equal(scanID, scans[0].ID)
	s.Require().Equal(repoID, scans[0].RepositoryID)
	s.Require().Equal(repositoryURL, scans[0].RepositoryURL)
	s.Require().Equal(repoName, scans[0].RepositoryName)
}

func (s *scanSuite) TestListScansWithError() {
	repoID := int64(1)
	request := &ListScansRequest{
		Page:         1,
		Size:         20,
		RepositoryID: &repoID,
	}

	s.scanRepo.EXPECT().List(gomock.Any(), request.Size, request.Page, gomock.Any()).
		Return(nil, errors.New("failed to list scans"))
	s.repo.EXPECT().Scan().Return(s.scanRepo)

	scans, err := s.scanService.ListScans(context.Background(), request)
	s.Require().Error(err)
	s.Require().Nil(scans)
}

func (s *scanSuite) TestUpdateScan() {
	scan := &models.Scan{}
	timeNow := time.Now()
	request := &UpdateScanRequest{
		Status:     models.ScanStatusQueued,
		Findings:   []byte{},
		QueuedAt:   &timeNow,
		ScanningAt: &timeNow,
		FinishedAt: &timeNow,
	}
	changesets := map[string]interface{}{
		"status":      models.ScanStatusQueued,
		"findings":    []byte{},
		"queued_at":   &timeNow,
		"scanning_at": &timeNow,
		"finished_at": &timeNow,
	}
	s.scanRepo.EXPECT().UpdateWithMap(gomock.Any(), scan, changesets).Return(nil)
	s.repo.EXPECT().Scan().Return(s.scanRepo)

	_, err := s.scanService.UpdateScan(context.Background(), scan, request)
	s.Require().NoError(err)
}

func (s *scanSuite) TestUpdateScanWithFailedUpdation() {
	scan := &models.Scan{}
	timeNow := time.Now()
	request := &UpdateScanRequest{
		Status:     models.ScanStatusQueued,
		Findings:   []byte{},
		QueuedAt:   &timeNow,
		ScanningAt: &timeNow,
		FinishedAt: &timeNow,
	}
	changesets := map[string]interface{}{
		"status":      models.ScanStatusQueued,
		"findings":    []byte{},
		"queued_at":   &timeNow,
		"scanning_at": &timeNow,
		"finished_at": &timeNow,
	}
	s.scanRepo.EXPECT().UpdateWithMap(gomock.Any(), scan, changesets).Return(errors.New("invalid data"))
	s.repo.EXPECT().Scan().Return(s.scanRepo)

	_, err := s.scanService.UpdateScan(context.Background(), scan, request)
	s.Require().Error(err)
}

func (s *scanSuite) TestTriggerScan() {
	repoID := int64(1)
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	expectedScan, _ := models.NewScan(expectedRepository)
	request := &TriggerScanRequest{
		RepositoryID: repoID,
	}

	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(expectedRepository, nil)
	s.scanRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(expectedScan, nil)
	s.kafkaWriter.EXPECT().WriteMessage(gomock.Any(), gomock.Any()).Return(nil)
	s.scanRepo.EXPECT().UpdateWithMap(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	s.repo.EXPECT().Scan().Return(s.scanRepo).Times(2)
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)

	scan, err := s.scanService.TriggerScan(context.Background(), request)
	s.Require().NoError(err)
	s.Require().NotNil(scan)
	s.Require().NotNil(scan.QueuedAt)
	s.Require().Equal(models.ScanStatusQueued, scan.Status)
}

func (s *scanSuite) TestTriggerScanWithNotFoundRepo() {
	repoID := int64(1)
	request := &TriggerScanRequest{
		RepositoryID: repoID,
	}

	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(nil, errors.New("record not found"))
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)

	scan, err := s.scanService.TriggerScan(context.Background(), request)
	s.Require().Error(err)
	s.Require().Nil(scan)
}

func (s *scanSuite) TestTriggerScanWithFailedWritingMessage() {
	repoID := int64(1)
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	expectedScan, _ := models.NewScan(expectedRepository)
	request := &TriggerScanRequest{
		RepositoryID: repoID,
	}

	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(expectedRepository, nil)
	s.scanRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(expectedScan, nil)
	s.kafkaWriter.EXPECT().WriteMessage(gomock.Any(), gomock.Any()).Return(errors.New("failed to write message"))
	s.repo.EXPECT().Scan().Return(s.scanRepo)
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)

	scan, err := s.scanService.TriggerScan(context.Background(), request)
	s.Require().Error(err)
	s.Require().Nil(scan)
}
func (s *scanSuite) TestHandleResultMessageWithSuccess() {
	scanID := int64(1)
	timeNow := time.Now()
	messageResult := &models.ScanResultMessage{
		ScanID:     scanID,
		ScanStatus: models.ScanStatusSuccess,
		FinishedAt: &timeNow,
	}
	changesets := map[string]interface{}{
		"status":      models.ScanStatusSuccess,
		"finished_at": &timeNow,
	}
	s.scanRepo.EXPECT().UpdateWithMap(gomock.Any(), gomock.Any(), changesets).Return(nil)
	s.repo.EXPECT().Scan().Return(s.scanRepo)

	err := s.scanService.HandleResultMessage(context.Background(), messageResult)
	s.Require().NoError(err)
}

func (s *scanSuite) TestHandleResultMessageWithInProgress() {
	scanID := int64(1)
	timeNow := time.Now()
	messageResult := &models.ScanResultMessage{
		ScanID:     scanID,
		ScanStatus: models.ScanStatusInProgress,
		ScanningAt: &timeNow,
	}
	changesets := map[string]interface{}{
		"status":      models.ScanStatusInProgress,
		"scanning_at": &timeNow,
	}
	s.scanRepo.EXPECT().UpdateWithMap(gomock.Any(), gomock.Any(), changesets).Return(nil)
	s.repo.EXPECT().Scan().Return(s.scanRepo)

	err := s.scanService.HandleResultMessage(context.Background(), messageResult)
	s.Require().NoError(err)
}

func (s *scanSuite) TestHandleResultMessageWithFailure() {
	scanID := int64(1)
	timeNow := time.Now()
	messageResult := &models.ScanResultMessage{
		ScanID:     scanID,
		ScanStatus: models.ScanStatusFailure,
		FinishedAt: &timeNow,
	}
	changesets := map[string]interface{}{
		"status":      models.ScanStatusFailure,
		"finished_at": &timeNow,
	}
	s.scanRepo.EXPECT().UpdateWithMap(gomock.Any(), gomock.Any(), changesets).Return(nil)
	s.repo.EXPECT().Scan().Return(s.scanRepo)

	err := s.scanService.HandleResultMessage(context.Background(), messageResult)
	s.Require().NoError(err)
}

func (s *scanSuite) TestHandleResultMessageWithInvalidStatus() {
	scanID := int64(1)
	timeNow := time.Now()
	messageResult := &models.ScanResultMessage{
		ScanID:     scanID,
		ScanStatus: "invalid_status",
		FinishedAt: &timeNow,
	}

	err := s.scanService.HandleResultMessage(context.Background(), messageResult)
	s.Require().NoError(err)
}
