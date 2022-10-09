package api

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/vumanhcuongit/scan/pkg/models"
)

type CreateRepositoryRequest struct {
	RepositoryURL string `json:"repository_url" binding:"required"`
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
