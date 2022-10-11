package repos

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/vumanhcuongit/scan/pkg/models"
)

type ScanSQLRepo struct {
	db *gorm.DB
}

// NewScanSQLRepo returns a new IScanRepo
func NewScanSQLRepo(db *gorm.DB) IScanRepo {
	return &ScanSQLRepo{
		db: db,
	}
}

func (r *ScanSQLRepo) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *ScanSQLRepo) GetByID(ctx context.Context, id int64) (*models.Scan, error) {
	record := &models.Scan{}
	err := r.dbWithContext(ctx).Where("id = ?", id).First(record).Error
	return record, err
}

func (r *ScanSQLRepo) Create(ctx context.Context, record *models.Scan) (*models.Scan, error) {
	err := r.dbWithContext(ctx).Create(record).Error
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *ScanSQLRepo) UpdateWithMap(
	ctx context.Context,
	record *models.Scan,
	params map[string]interface{},
) error {
	return r.dbWithContext(ctx).
		Model(record).
		Updates(params).
		Error
}

func (r *ScanSQLRepo) Delete(ctx context.Context, record *models.Scan) error {
	return r.dbWithContext(ctx).Delete(record).Error
}

func (r *ScanSQLRepo) MarkStaleScansAsFailure(
	ctx context.Context,
	maxMinutes int,
) error {
	timeNow := time.Now()
	staleTime := timeNow.Add(-1 * time.Minute * time.Duration(maxMinutes))
	return r.dbWithContext(ctx).
		Model(models.Scan{}).
		Where("status IN (?) AND (queued_at < ? OR scanning_at < ?)",
			[]string{models.ScanStatusQueued, models.ScanStatusInProgress}, staleTime, staleTime).
		Updates(models.Scan{Status: models.ScanStatusFailure, FinishedAt: &timeNow}).Error
}

func (r *ScanSQLRepo) List(
	ctx context.Context,
	size int,
	page int,
	filter *models.ScanFilter,
) ([]*models.Scan, error) {
	var records []*models.Scan
	query := r.buildQueryFromFilter(ctx, filter)
	offset := (page - 1) * size
	err := query.Order("id DESC").Limit(size).Offset(offset).Find(&records).Error
	return records, err
}

func (r *ScanSQLRepo) buildQueryFromFilter(
	ctx context.Context,
	filter *models.ScanFilter,
) *gorm.DB {
	query := r.dbWithContext(ctx)

	if filter == nil {
		return query
	}

	if filter.RepositoryID != nil {
		query = query.Where("repository_id = ?", filter.RepositoryID)
	}

	if filter.RepositoryName != nil {
		query = query.Where("repository_name = ?", filter.RepositoryName)
	}

	if filter.Status != nil {
		query = query.Where("status = ?", filter.Status)
	}

	return query
}
