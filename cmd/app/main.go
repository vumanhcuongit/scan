package main

import (
	"encoding/json"
	"flag"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/vumanhcuongit/scan/internal"
	"github.com/vumanhcuongit/scan/pkg/infra"
	"github.com/vumanhcuongit/scan/pkg/kafka"
	"go.uber.org/zap"

	"github.com/vumanhcuongit/scan/internal/config"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config-file", "", "Specify config file path")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	cfg, err := config.Load(configFile)
	infra.CheckError(err)

	configBytes, err := json.MarshalIndent(cfg, "", "   ")
	if err != nil {
		zap.S().Warnf("could not convert config to JSON: %v", err)
	} else {
		zap.S().Debugf("load config %s", string(configBytes))
	}

	infra.ConfigApplication(cfg.EnvConfig)
	migrateDB(cfg)
	startApp(cfg)
}

func migrateDB(cfg *config.App) {
	mgTool := infra.GetMigrateTool()
	zap.S().Infof("migrate service with url: %v", cfg.DB.MigrationConnURL)
	db := mgTool.CreateDBAndMigrate(cfg.DB, "file://migrations/sql")
	defer func() {
		err := db.Close()
		if err != nil {
			zap.S().Errorf("failed to close db")
		}
	}()
}

func startApp(cfg *config.App) {
	kafkaWriter := kafka.NewWriter(cfg.MessageQueue.Broker, cfg.MessageQueue.TopicRequest)
	kafkaReader := kafka.NewReader(cfg.MessageQueue.Broker, cfg.MessageQueue.TopicReply, cfg.MessageQueue.ScanningGroupID)
	s := internal.NewServer(cfg, kafkaWriter, kafkaReader)
	go func() {
		err := s.Listen()
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		err := s.Start()
		if err != nil {
			panic(err)
		}
	}()
	infra.WaitOSSignal()
}
