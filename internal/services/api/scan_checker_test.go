package api

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/internal/repos"
	"go.uber.org/zap"
)

type scanCheckerSuite struct {
	suite.Suite

	config      *config.App
	mockCtrl    *gomock.Controller
	repo        *repos.MockIRepo
	scanRepo    *repos.MockIScanRepo
	scanChecker *ScanChecker
}

func TestScanCheckerSuite(t *testing.T) {
	suite.Run(t, &scanCheckerSuite{})
}

func (s *scanCheckerSuite) SetupSuite() {
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	cfg, err := config.Load("")
	if err != nil {
		s.Require().NoError(err)
	}
	s.config = cfg
}

func (s *scanCheckerSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.repo = repos.NewMockIRepo(s.mockCtrl)
	s.scanRepo = repos.NewMockIScanRepo(s.mockCtrl)
	s.scanChecker = NewScanChecker(s.repo, &s.config.ScanChecker)
}

func (s *scanCheckerSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *scanCheckerSuite) TestCheck() {
	s.scanRepo.EXPECT().
		MarkStaleScansAsFailure(gomock.Any(), s.config.ScanChecker.MaxStaleTimeInMinutes).
		Return(nil)
	s.repo.EXPECT().Scan().Return(s.scanRepo)

	err := s.scanChecker.Check(context.Background())
	s.Require().NoError(err)
}

func (s *scanCheckerSuite) TestCheckWithFailedUpdation() {
	s.scanRepo.EXPECT().
		MarkStaleScansAsFailure(gomock.Any(), s.config.ScanChecker.MaxStaleTimeInMinutes).
		Return(errors.New("failed to mark stale scans"))
	s.repo.EXPECT().Scan().Return(s.scanRepo)

	err := s.scanChecker.Check(context.Background())
	s.Require().Error(err)
}
