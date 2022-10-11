package job

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/suite"
	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/pkg/gitscan"
	"github.com/vumanhcuongit/scan/pkg/kafka"
	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

const (
	exampleScanID    = 1
	exampleOwnerName = "vumanhcuongit"
	exampleRepoName  = "scan"
)

type jobSuite struct {
	suite.Suite

	mockCtrl    *gomock.Controller
	config      *config.App
	kafkaWriter *kafka.MockIWriter
	gitscan     *gitscan.MockIGitScan
	job         *Job
	exampleTask *asynq.Task
}

func TestJobSuite(t *testing.T) {
	suite.Run(t, &jobSuite{})
}

func (s *jobSuite) SetupSuite() {
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

func (s *jobSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.kafkaWriter = kafka.NewMockIWriter(s.mockCtrl)
	s.gitscan = gitscan.NewMockIGitScan(s.mockCtrl)
	s.job = NewJob(s.gitscan, s.kafkaWriter)
	payload, _ := json.Marshal(ScanSourceCodePayload{
		ScanID:    exampleScanID,
		OwnerName: exampleOwnerName,
		RepoName:  exampleRepoName,
	})
	s.exampleTask = asynq.NewTask(TypeScanSourceCode, payload)
}

func (s *jobSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *jobSuite) TestNewScanSourceCodeJob() {
	job, err := s.job.NewScanSourceCodeJob(exampleScanID, exampleOwnerName, exampleRepoName)
	s.Require().NoError(err)
	s.Require().NotNil(job)
}

func (s *jobSuite) TestHandleScanSourceCodeJobWithSuccessfulScan() {
	// produce in progress message
	s.kafkaWriter.EXPECT().WriteMessage(gomock.Any(), gomock.Any()).Return(nil)

	// scan successfully
	exampleFindings := []models.Finding{
		{
			Type:   "sast",
			RuleID: "1",
		},
	}
	s.gitscan.EXPECT().Scan(gomock.Any(), exampleOwnerName, exampleRepoName).Return(exampleFindings, nil)

	// produce successful result message
	s.kafkaWriter.EXPECT().WriteMessage(gomock.Any(), gomock.Any()).Return(nil)

	err := s.job.HandleScanSourceCodeJob(context.Background(), s.exampleTask)
	s.Require().NoError(err)
}

func (s *jobSuite) TestHandleScanSourceCodeJobWithFailedScan() {
	// produce in progress message
	s.kafkaWriter.EXPECT().WriteMessage(gomock.Any(), gomock.Any()).Return(nil)

	// scan failed
	s.gitscan.EXPECT().Scan(gomock.Any(), exampleOwnerName, exampleRepoName).Return(nil, errors.New("example error"))

	// produce failed result message
	s.kafkaWriter.EXPECT().WriteMessage(gomock.Any(), gomock.Any()).Return(nil)

	err := s.job.HandleScanSourceCodeJob(context.Background(), s.exampleTask)
	s.Require().NoError(err)
}

func (s *jobSuite) TestdoWriteMessage() {
	// failed to write message
	s.kafkaWriter.EXPECT().WriteMessage(gomock.Any(), gomock.Any()).Return(errors.New("example error"))
	err := s.job.doWriteMessage(context.Background(), []byte{})
	s.Require().Error(err)
}
