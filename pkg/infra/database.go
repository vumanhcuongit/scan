package infra

import (
	"fmt"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"gorm.io/driver/mysql"
	gormio "gorm.io/gorm"

	"go.uber.org/zap"

	gormlogger "gorm.io/gorm/logger"

	"github.com/vumanhcuongit/scan/internal/config"

	tLogger "github.com/vumanhcuongit/scan/pkg/logger"
)

// GormLogger ...
type GormLogger struct {
	zap *zap.SugaredLogger
}

// NewLogger ...
func NewLogger(logger *zap.SugaredLogger) GormLogger {
	return GormLogger{zap: logger}
}

// Print ...
func (l GormLogger) Print(v ...interface{}) {
	switch v[0] {
	case "sql":
		l.zap.With(
			[]interface{}{
				"module", "gorm",
				"type", "sql",
				"rows", v[5],
				"src_ref", v[1],
				"values", v[4],
				"duration", v[2],
			}...,
		).Info(v[3])
	case "log":
		l.zap.With(
			[]interface{}{
				"module", "gorm",
				"type", "log",
			}...,
		).Info(v[2])
	}
}

// InitDatabase ...
func InitDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(cfg.DriverName, cfg.DataSource)
	if err != nil {
		fmt.Println("---------------------")
		fmt.Println(err)
		return db, nil
	}

	fmt.Println("---------------------")
	fmt.Println(cfg)

	db.SetLogger(NewLogger(zap.S()))
	db.LogMode(cfg.IsDevMode)
	db.DB().SetMaxOpenConns(cfg.MaxOpenConns)
	db.DB().SetMaxIdleConns(0)
	db.DB().SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeTimeMiliseconds) * time.Millisecond)

	return db, nil
}

// InitDatabaseWithBackoff ...
func InitDatabaseWithBackoff(cfg *config.DatabaseConfig) *gorm.DB {
	var db *gorm.DB

	boff := backoff.NewExponentialBackOff()
	err := backoff.Retry(func() error {
		var e error
		db, e = InitDatabase(cfg)
		if e != nil {
			zap.S().Warnf("Connect database failed, err: %+v", e.Error())
			return e
		}
		e = db.DB().Ping()
		if e != nil {
			zap.S().Warnf("Connect database failed, err: %+v", e.Error())
			return e
		}
		return nil
	}, boff)
	if err != nil {
		zap.S().Errorf("Connect database failed, err: %+v", err.Error())
		panic(err)
	}

	return db
}

// NewDatabase ...
func NewDatabase(cfg *config.DatabaseConfig) (*gormio.DB, error) {
	logger := tLogger.NewGormLogger()
	logger.LogLevel = gormlogger.Info
	logger.SetAsDefault()
	db, err := gormio.Open(mysql.Open(cfg.DataSource), &gormio.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger,
	})

	if err != nil {
		return db, nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeTimeMiliseconds) * time.Millisecond)
	return db, nil
}

// NewDatabaseWithBackOff ...
func NewDatabaseWithBackOff(cfg *config.DatabaseConfig) (*gormio.DB, error) {
	var db *gormio.DB
	exponentialBackOff := backoff.NewExponentialBackOff()
	err := backoff.Retry(func() error {
		var err error
		db, err = NewDatabase(cfg)
		if err != nil {
			zap.S().Warnf("Connect database failed, err: %+v", err.Error())
			return err
		}

		sqlDB, err := db.DB()
		if err != nil {
			zap.S().Warnf("Connect database failed, err: %+v", err.Error())
			return err
		}

		err = sqlDB.Ping()
		if err != nil {
			zap.S().Warnf("Connect database failed, err: %+v", err.Error())
			return err
		}

		return nil
	}, exponentialBackOff)

	if err != nil {
		zap.S().Errorf("Connect database failed, err: %+v", err.Error())
		return nil, err
	}

	return db, nil
}
