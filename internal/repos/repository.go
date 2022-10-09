package repos

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/vumanhcuongit/scan/pkg/models"
)

type RepositorySQLRepo struct {
	db *gorm.DB
}

// NewRepositorySQLRepo returns a new IRepositoryRepo
func NewRepositorySQLRepo(db *gorm.DB) IRepositoryRepo {
	return &RepositorySQLRepo{
		db: db,
	}
}

func (r *RepositorySQLRepo) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *RepositorySQLRepo) GetByID(ctx context.Context, id int64) (*models.Repository, error) {
	record := &models.Repository{}
	err := r.dbWithContext(ctx).Where("id = ?", id).First(record).Error
	return record, err
}

func (r *RepositorySQLRepo) Create(ctx context.Context, record *models.Repository) (*models.Repository, error) {
	err := r.dbWithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	}).Create(record).Error
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *RepositorySQLRepo) UpdateWithMap(
	ctx context.Context,
	record *models.Repository,
	params map[string]interface{},
) error {
	return r.dbWithContext(ctx).
		Model(record).
		Updates(params).
		Error
}

func (r *RepositorySQLRepo) Delete(ctx context.Context, record *models.Repository) error {
	return r.dbWithContext(ctx).Delete(record).Error
}

func (r *RepositorySQLRepo) List(
	ctx context.Context,
	size int,
	page int,
	filter *models.RepositoryFilter,
) ([]*models.Repository, error) {
	var records []*models.Repository
	query := r.buildQueryFromFilter(ctx, filter)
	offset := (page - 1) * size
	err := query.Order("id DESC").Limit(size).Offset(offset).Find(&records).Error
	return records, err
}

func (r *RepositorySQLRepo) buildQueryFromFilter(
	ctx context.Context,
	filter *models.RepositoryFilter,
) *gorm.DB {
	query := r.dbWithContext(ctx)

	if filter == nil {
		return query
	}

	if filter.RepositoryID != nil {
		query = query.Where("repository_id = ?", filter.RepositoryID)
	}

	return query
}
