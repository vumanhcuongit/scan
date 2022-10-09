package infra

import (
	"sync"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jinzhu/gorm"
	"github.com/vumanhcuongit/scan/internal/config"
	"go.uber.org/zap"
)

// IMigrateTool tool to migrate schema and data.
type IMigrateTool interface {
	// Create test db for unit test
	CreateDBAndMigrate(cfg *config.DatabaseConfig, migrationFile string) *gorm.DB

	// Migrate from current version to latest verion.
	Migrate(source string, connStr string)
}

type migrateTool struct{}

var once sync.Once         // nolint
var mutex = &sync.Mutex{}  // nolint
var singleton IMigrateTool // nolint

// GetMigrateTool get singleton instance for migrate tool
func GetMigrateTool() IMigrateTool { // nolint
	once.Do(func() {
		singleton = &migrateTool{}
	})
	return singleton
}

// Migrate execute migration in serialize.
func (mt *migrateTool) Migrate(source string, connStr string) {
	mutex.Lock()
	defer mutex.Unlock()

	zap.S().Info("Migrating....")
	zap.S().Infof("Source=%+v Connection=%+v\n", source, connStr)

	mg, err := migrate.New(source, connStr)
	if err != nil {
		zap.S().Errorf("Migrate failed with error=%+v", err.Error())
		panic(err)
	}
	defer mg.Close()

	version, dirty, err := mg.Version()
	if err != nil && err != migrate.ErrNilVersion {
		zap.S().Errorf("Migrate failed with error=%+v", err.Error())
		panic(err)
	}

	if dirty {
		mg.Force(int(version) - 1) // nolint
	}

	err = mg.Up()

	if err != nil && err != migrate.ErrNoChange {
		zap.S().Errorf("Migrate failed with error=%+v", err.Error())
		panic(err)
	}

	zap.S().Info("Migration done...")
}

// CreateDBAndMigrate create test store DB and operator DB to execute unit test.
func (mt *migrateTool) CreateDBAndMigrate(cfg *config.DatabaseConfig, migrationFile string) *gorm.DB {
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

	mt.Migrate(migrationFile, cfg.MigrationConnURL)
	return db
}
