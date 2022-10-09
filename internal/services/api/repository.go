package api

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/vumanhcuongit/scan/pkg/models"
)

type CreateRepositoryRequest struct {
	RepositoryURL string `json:"repository_url" binding:"required"`
}

type UpdateRepositoryRequest struct {
	Name          string `json:"name"`
	Owner         string `json:"owner"`
	RepositoryURL string `json:"repository_url"`
}

type ListRepositoriesRequest struct {
	RepositoryID *int64 `json:"repository_id" form:"repository_id"`
	Size         int    `json:"size" form:"size"`
	Page         int    `json:"page" form:"page"`
}

func (s *ScanService) CreateRepository(ctx context.Context, request *CreateRepositoryRequest) (*models.Repository, error) {
	log := ctxzap.Extract(ctx).Sugar()
	log.Infof("starting to create a repository with request %+v", request)

	record, err := models.NewRepository(request.RepositoryURL)
	if err != nil {
		log.Warnf("failed to init repository, err: %+v", err)
		return nil, err
	}

	repository, err := s.repo.Repository().Create(ctx, record)
	if err != nil {
		log.Warnf("failed to create repository, err: %+v", err)
		return nil, err
	}

	return repository, nil
}

func (s *ScanService) GetRepository(ctx context.Context, repositoryID int64) (*models.Repository, error) {
	log := ctxzap.Extract(ctx).Sugar()
	log.Infof("starting to get repository with id %d", repositoryID)

	repository, err := s.repo.Repository().GetByID(ctx, repositoryID)
	if err != nil {
		log.Warnf("failed to get repository, err: %+v", err)
		return nil, err
	}

	return repository, nil
}

func (s *ScanService) ListRepositories(ctx context.Context, request *ListRepositoriesRequest) ([]*models.Repository, error) {
	log := ctxzap.Extract(ctx).Sugar()
	log.Infof("starting to list repositories with request %+v", request)

	filter := &models.RepositoryFilter{}
	if request.RepositoryID != nil {
		filter.RepositoryID = request.RepositoryID
	}
	repositories, err := s.repo.Repository().List(ctx, request.Size, request.Page, filter)
	if err != nil {
		log.Warnf("failed to list repositories, err: %+v", err)
		return nil, err
	}

	return repositories, nil
}

func (s *ScanService) UpdateRepository(ctx context.Context, repositoryID int64, request *UpdateRepositoryRequest) (*models.Repository, error) {
	log := ctxzap.Extract(ctx).Sugar()
	log.Infof("starting to update repository with request %+v", request)

	repository, err := s.repo.Repository().GetByID(ctx, repositoryID)
	if err != nil {
		log.Warnf("failed to get repository, err: %+v", err)
		return nil, err
	}

	changesets := map[string]interface{}{}
	if request.Name != "" {
		changesets["name"] = request.Name
		repository.Name = request.Name
	}
	if request.Owner != "" {
		changesets["owner"] = request.Owner
		repository.Owner = request.Owner
	}
	if request.RepositoryURL != "" {
		changesets["repository_url"] = request.RepositoryURL
		repository.RepositoryURL = request.RepositoryURL
	}

	err = s.repo.Repository().UpdateWithMap(ctx, repository, changesets)
	if err != nil {
		log.Warnf("failed to update repository, err: +%v", err)
		return nil, err
	}

	return repository, nil
}

func (s *ScanService) DeleteRepository(ctx context.Context, repositoryID int64) error {
	log := ctxzap.Extract(ctx).Sugar()
	log.Infof("starting to delete repository with id %d", repositoryID)

	repository, err := s.repo.Repository().GetByID(ctx, repositoryID)
	if err != nil {
		log.Warnf("failed to get repository, err: %+v", err)
		return err
	}

	err = s.repo.Repository().Delete(ctx, repository)
	if err != nil {
		log.Warnf("failed to delete repository, err: %+v", err)
		return err
	}

	return nil
}
