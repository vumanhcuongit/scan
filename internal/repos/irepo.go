package repos

import (
	"context"

	"github.com/vumanhcuongit/scan/pkg/models"
)

//go:generate mockgen -source=irepo.go -destination=irepo.mock.go -package=repos

type IRepo interface {
	Stop()
	CleanDB()
	WithTransaction(ctx context.Context, fn func(IRepo) error) (err error)
	Repository() IRepositoryRepo
}

type IRepositoryRepo interface {
	Create(ctx context.Context, record *models.Repository) (*models.Repository, error)
	GetByID(ctx context.Context, id int64) (*models.Repository, error)
	UpdateWithMap(
		ctx context.Context,
		record *models.Repository,
		params map[string]interface{},
	) error
	Delete(ctx context.Context, record *models.Repository) error
	List(
		ctx context.Context,
		size int,
		page int,
		filter *models.RepositoryFilter,
	) ([]*models.Repository, error)
}

type IScanRepo interface {
	Create(ctx context.Context, record *models.Scan) (*models.Scan, error)
	GetByID(ctx context.Context, id int64) (*models.Scan, error)
	UpdateWithMap(
		ctx context.Context,
		record *models.Scan,
		params map[string]interface{},
	) error
	Delete(ctx context.Context, record *models.Scan) error
	List(
		ctx context.Context,
		size int,
		page int,
		filter *models.ScanFilter,
	) ([]*models.Scan, error)
}
