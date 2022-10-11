package api

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/vumanhcuongit/scan/internal/repos"
	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

type scanSuite struct {
	suite.Suite

	mockCtrl    *gomock.Controller
	repo        *repos.MockIRepo
	scanRepo    *repos.MockIScanRepo
	scanService *ScanService
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
	s.scanService = &ScanService{repo: s.repo}
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
