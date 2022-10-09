package repos

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

// NewSQLRepo returns a IRepo
func NewSQLRepo(db *gorm.DB) IRepo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) Stop() {
	sqlDB, _ := r.db.DB()
	if sqlDB != nil {
		_ = sqlDB.Close()
	}
}

func (r *Repo) CleanDB() {
	listModels := []string{}

	for _, model := range listModels {
		query := fmt.Sprintf("TRUNCATE TABLE %s;", model)
		r.db.Exec(query)
	}
}

func (r *Repo) WithTransaction(ctx context.Context, fn func(IRepo) error) (err error) {
	log := ctxzap.Extract(ctx).Sugar()

	log.Info("Starting transaction")

	tx := r.db.Begin()
	tr := &Repo{
		db: tx,
	}
	err = tx.Error
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			log.Warnf("Transaction failed with panic: %+v", p)
			// a panic occurred, rollback and repanic
			tx.Rollback()
			panic(p)
		}

		if err != nil {
			log.Warnf("Transaction failure with error: %+v", err)
			// something went wrong, rollback
			tx.Rollback()
		} else {
			log.Info("Finishing transaction")
			// all good, commit
			err = tx.Commit().Error
		}
	}()

	err = fn(tr)

	return err
}

func (r *Repo) Repository() IRepositoryRepo {
	return NewRepositorySQLRepo(r.db)
}

func (r *Repo) Scan() IScanRepo {
	return NewScanSQLRepo(r.db)
}
