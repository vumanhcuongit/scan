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
	Registration() IRegistrationRepo
}

type IRegistrationRepo interface {
	Create(ctx context.Context, record *models.Registration) (*models.Registration, error)
	GetByAppIdentifier(ctx context.Context, appIdentifier string) (*models.Registration, error)
	GetByID(ctx context.Context, id int64) (*models.Registration, error)
	UpdateWithMap(
		ctx context.Context,
		record *models.Registration,
		params map[string]interface{},
	) error
	Delete(ctx context.Context, record *models.Registration) error
	List(
		ctx context.Context,
		size int,
		page int,
		filter *models.RegistrationFilter,
	) ([]*models.Registration, error)
}
