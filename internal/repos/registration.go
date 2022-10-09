package repos

import (
	"context"

	"gorm.io/gorm"

	"github.com/vumanhcuongit/scan/pkg/models"
)

type RegistrationSQLRepo struct {
	db *gorm.DB
}

// NewRegistrationSQLRepo returns a new IRegistrationRepo
func NewRegistrationSQLRepo(db *gorm.DB) IRegistrationRepo {
	return &RegistrationSQLRepo{
		db: db,
	}
}

func (r *RegistrationSQLRepo) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *RegistrationSQLRepo) GetByID(ctx context.Context, id int64) (*models.Registration, error) {
	record := &models.Registration{}
	err := r.dbWithContext(ctx).Where("id = ?", id).First(record).Error
	return record, err
}

func (r *RegistrationSQLRepo) GetByAppIdentifier(
	ctx context.Context,
	appIdentifier string,
) (*models.Registration, error) {
	record := &models.Registration{}
	err := r.dbWithContext(ctx).Where("app_identifier = ?", appIdentifier).First(record).Error

	return record, err
}

func (r *RegistrationSQLRepo) Create(ctx context.Context, record *models.Registration) (*models.Registration, error) {
	err := r.dbWithContext(ctx).Create(record).Error
	if err != nil {
		return nil, err
	}

	return record, nil

}

func (r *RegistrationSQLRepo) UpdateWithMap(
	ctx context.Context,
	record *models.Registration,
	params map[string]interface{},
) error {
	return r.dbWithContext(ctx).
		Model(record).
		Updates(params).
		Error
}

func (r *RegistrationSQLRepo) Delete(ctx context.Context, record *models.Registration) error {
	return r.dbWithContext(ctx).Delete(record).Error
}

func (r *RegistrationSQLRepo) List(
	ctx context.Context,
	size int,
	page int,
	filter *models.RegistrationFilter,
) ([]*models.Registration, error) {
	var records []*models.Registration
	query := r.buildQueryFromFilter(ctx, filter)
	offset := (page - 1) * size
	err := query.Order("id DESC").Limit(size).Offset(offset).Find(&records).Error
	return records, err
}

func (r *RegistrationSQLRepo) Count(
	ctx context.Context,
	filter *models.RegistrationFilter,
) (int, error) {
	query := r.buildQueryFromFilter(ctx, filter)
	var total int64
	err := query.Model(&models.Registration{}).Count(&total).Error
	return int(total), err
}

func (r *RegistrationSQLRepo) buildQueryFromFilter(
	ctx context.Context,
	filter *models.RegistrationFilter,
) *gorm.DB {
	query := r.dbWithContext(ctx)

	if filter == nil {
		return query
	}

	if filter.AppIdentifier != nil {
		query = query.Where("app_identifier = ?", filter.AppIdentifier)
	}

	return query
}
