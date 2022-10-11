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

type repositorySuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	repo           *repos.MockIRepo
	repositoryRepo *repos.MockIRepositoryRepo
	scanService    *ScanService
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, &repositorySuite{})
}

func (s *repositorySuite) SetupSuite() {
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()
	undo := zap.ReplaceGlobals(logger)
	defer undo()
}

func (s *repositorySuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.repo = repos.NewMockIRepo(s.mockCtrl)
	s.repositoryRepo = repos.NewMockIRepositoryRepo(s.mockCtrl)
	s.scanService = &ScanService{repo: s.repo}
}

func (s *repositorySuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *repositorySuite) TestCreateRepository() {
	repoID := int64(1)
	ownerName := "vumanhcuongit"
	repoName := "scan"
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	s.repositoryRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(expectedRepository, nil)
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)
	request := &CreateRepositoryRequest{
		RepositoryURL: repositoryURL,
	}

	repository, err := s.scanService.CreateRepository(context.Background(), request)
	s.Require().NoError(err)
	s.Require().NotNil(repository)
	s.Require().Equal(repoID, repository.ID)
	s.Require().Equal(repositoryURL, repository.RepositoryURL)
	s.Require().Equal(ownerName, repository.Owner)
	s.Require().Equal(repoName, repository.Name)
}

func (s *repositorySuite) TestCreateRepositoryWithInvalidURL() {
	repositoryURL := "this-is-an-invalid-URL"
	request := &CreateRepositoryRequest{
		RepositoryURL: repositoryURL,
	}

	repository, err := s.scanService.CreateRepository(context.Background(), request)
	s.Require().Error(err)
	s.Require().Nil(repository)
}

func (s *repositorySuite) TestCreateRepositoryWithFailedCreation() {
	repoID := int64(1)
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	s.repositoryRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("invalid data"))
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)
	request := &CreateRepositoryRequest{
		RepositoryURL: repositoryURL,
	}

	repository, err := s.scanService.CreateRepository(context.Background(), request)
	s.Require().Error(err)
	s.Require().Nil(repository)
}

func (s *repositorySuite) TestGetRepository() {
	repoID := int64(1)
	ownerName := "vumanhcuongit"
	repoName := "scan"
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(expectedRepository, nil)
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)

	repository, err := s.scanService.GetRepository(context.Background(), repoID)
	s.Require().NoError(err)
	s.Require().NotNil(repository)
	s.Require().Equal(repoID, repository.ID)
	s.Require().Equal(repositoryURL, repository.RepositoryURL)
	s.Require().Equal(ownerName, repository.Owner)
	s.Require().Equal(repoName, repository.Name)
}

func (s *repositorySuite) TestGetRepositoryWithRecordNotFound() {
	repoID := int64(1)
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(nil, errors.New("record not found"))
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)

	repository, err := s.scanService.GetRepository(context.Background(), repoID)
	s.Require().Error(err)
	s.Require().Nil(repository)
}

func (s *repositorySuite) TestListRepositories() {
	repoID := int64(1)
	ownerName := "vumanhcuongit"
	repoName := "scan"
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	request := &ListRepositoriesRequest{
		Page: 1,
		Size: 20,
	}

	s.repositoryRepo.EXPECT().List(gomock.Any(), request.Size, request.Page, &models.RepositoryFilter{}).
		Return([]*models.Repository{expectedRepository}, nil)
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)

	repositories, err := s.scanService.ListRepositories(context.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(1, len(repositories))
	s.Require().Equal(repoID, repositories[0].ID)
	s.Require().Equal(repositoryURL, repositories[0].RepositoryURL)
	s.Require().Equal(ownerName, repositories[0].Owner)
	s.Require().Equal(repoName, repositories[0].Name)
}

func (s *repositorySuite) TestListRepositoriesWithError() {
	request := &ListRepositoriesRequest{
		Page: 1,
		Size: 20,
	}
	s.repositoryRepo.EXPECT().List(gomock.Any(), request.Size, request.Page, &models.RepositoryFilter{}).
		Return(nil, errors.New("failed to list repositories"))
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)

	repositories, err := s.scanService.ListRepositories(context.Background(), request)
	s.Require().Error(err)
	s.Require().Nil(repositories)
}

func (s *repositorySuite) TestUpdateRepository() {
	repoID := int64(1)
	ownerName := "vumanhcuongit"
	repoName := "scan"
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	changesets := map[string]interface{}{
		"name":           repoName,
		"owner":          ownerName,
		"repository_url": repositoryURL,
	}
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(expectedRepository, nil)
	s.repositoryRepo.EXPECT().UpdateWithMap(gomock.Any(), expectedRepository, changesets).Return(nil)
	s.repo.EXPECT().Repository().Return(s.repositoryRepo).Times(2)

	request := &UpdateRepositoryRequest{
		Name:          repoName,
		Owner:         ownerName,
		RepositoryURL: repositoryURL,
	}
	repository, err := s.scanService.UpdateRepository(context.Background(), repoID, request)
	s.Require().NoError(err)
	s.Require().NotNil(repository)
	s.Require().Equal(repoName, repository.Name)
	s.Require().Equal(ownerName, repository.Owner)
	s.Require().Equal(repositoryURL, repository.RepositoryURL)
}

func (s *repositorySuite) TestUpdateRepositoryWithFailedUpdation() {
	repoID := int64(1)
	ownerName := "vumanhcuongit"
	repoName := "scan"
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	changesets := map[string]interface{}{
		"name":           repoName,
		"owner":          ownerName,
		"repository_url": repositoryURL,
	}
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(expectedRepository, nil)
	s.repositoryRepo.EXPECT().UpdateWithMap(gomock.Any(), expectedRepository, changesets).Return(errors.New("failed to update"))
	s.repo.EXPECT().Repository().Return(s.repositoryRepo).Times(2)

	request := &UpdateRepositoryRequest{
		Name:          repoName,
		Owner:         ownerName,
		RepositoryURL: repositoryURL,
	}
	repository, err := s.scanService.UpdateRepository(context.Background(), repoID, request)
	s.Require().Error(err)
	s.Require().Nil(repository)
}

func (s *repositorySuite) TestUpdateRepositoryWithRecordNotFound() {
	repoID := int64(1)
	ownerName := "vumanhcuongit"
	repoName := "scan"
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(nil, errors.New("record not found"))
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)

	request := &UpdateRepositoryRequest{
		Name:          repoName,
		Owner:         ownerName,
		RepositoryURL: repositoryURL,
	}
	repository, err := s.scanService.UpdateRepository(context.Background(), repoID, request)
	s.Require().Error(err)
	s.Require().Nil(repository)
}

func (s *repositorySuite) TestDeleteRepository() {
	repoID := int64(1)
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(expectedRepository, nil)
	s.repositoryRepo.EXPECT().Delete(gomock.Any(), expectedRepository).Return(nil)
	s.repo.EXPECT().Repository().Return(s.repositoryRepo).Times(2)

	err := s.scanService.DeleteRepository(context.Background(), repoID)
	s.Require().NoError(err)
}

func (s *repositorySuite) TestDeleteRepositoryWithFailedDeletion() {
	repoID := int64(1)
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(expectedRepository, nil)
	s.repositoryRepo.EXPECT().Delete(gomock.Any(), expectedRepository).Return(errors.New("failed to delete"))
	s.repo.EXPECT().Repository().Return(s.repositoryRepo).Times(2)

	err := s.scanService.DeleteRepository(context.Background(), repoID)
	s.Require().Error(err)
}

func (s *repositorySuite) TestDeleteRepositoryWithRecordNotFound() {
	repoID := int64(1)
	repositoryURL := "https://github.com/vumanhcuongit/scan"
	expectedRepository, _ := models.NewRepository(repositoryURL)
	expectedRepository.ID = repoID
	s.repositoryRepo.EXPECT().GetByID(gomock.Any(), repoID).Return(nil, errors.New("record not found"))
	s.repo.EXPECT().Repository().Return(s.repositoryRepo)

	err := s.scanService.DeleteRepository(context.Background(), repoID)
	s.Require().Error(err)
}
