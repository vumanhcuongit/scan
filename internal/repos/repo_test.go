package repos

import (
	"os"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/vumanhcuongit/scan/pkg/infra"
	"gorm.io/gorm"

	"github.com/vumanhcuongit/scan/internal/config"
)

type repoTestState struct {
	db   *gorm.DB
	repo IRepo
}

var state *repoTestState // nolint

func TestMain(m *testing.M) {
	cfg, err := config.Load("")
	if err != nil {
		panic(err)
	}

	testDB, err := infra.NewDatabaseWithBackOff(cfg.DB)
	if err != nil {
		panic(err)
	}

	repo := NewSQLRepo(testDB)
	state = &repoTestState{
		db:   testDB,
		repo: repo,
	}

	repo.CleanDB()
	exitVal := m.Run()
	repo.CleanDB()
	repo.Stop()
	os.Exit(exitVal)
}
