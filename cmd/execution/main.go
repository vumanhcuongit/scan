package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/internal/services/execution"
	"github.com/vumanhcuongit/scan/pkg/infra"
	"github.com/vumanhcuongit/scan/pkg/kafka"
	"go.uber.org/zap"
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

	kafkaReader := kafka.NewReader(cfg.MessageQueue.Broker, cfg.MessageQueue.TopicRequest, cfg.MessageQueue.WorkerGroupID)
	kafkaWriter := kafka.NewWriter(cfg.MessageQueue.Broker, cfg.MessageQueue.TopicReply)
	exec := execution.New(cfg, kafkaReader, kafkaWriter)
	defer exec.Stop()
	log.Printf("Starting execution service")
	log.Fatal(exec.Run(context.Background()))
}
