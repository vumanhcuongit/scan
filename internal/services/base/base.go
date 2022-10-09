package base

import (
	"github.com/vumanhcuongit/scan/pkg/infra"

	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/internal/repos"
)

type Service struct {
	cfg  *config.App
	repo repos.IRepo
}

func NewService(cfg *config.App) *Service {
	// init repo
	db, err := infra.NewDatabase(cfg.DB)
	if err != nil {
		panic(err)
	}

	repo := repos.NewSQLRepo(db)
	return &Service{
		cfg:  cfg,
		repo: repo,
	}
}

func (s *Service) SetRepo(repo repos.IRepo) {
	s.repo = repo
}

func (s *Service) Repo() repos.IRepo {
	return s.repo
}

func (s *Service) Config() *config.App {
	return s.cfg
}
