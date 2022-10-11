package execution

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/internal/services/execution/job"
	"github.com/vumanhcuongit/scan/pkg/gitscan"
	"github.com/vumanhcuongit/scan/pkg/kafka"
	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

type Execution struct {
	kafkaReader  *kafka.Reader
	kafkaWriter  kafka.IWriter
	jobManager   *job.Job      // job are processed concurrently by multiple workers
	workerClient *asynq.Client // client puts tasks on a queue
	workerServer *asynq.Server // server pulls tasks off queues and starts a worker goroutine for each task
	workerMux    *asynq.ServeMux
}

func New(cfg *config.App, kafkaReader *kafka.Reader, kafkaWriter kafka.IWriter) *Execution {
	gitScan := gitscan.NewGitScan(cfg.SourceCodesDir)
	jobManager := job.NewJob(gitScan, kafkaWriter)
	workerServer, workerMux, workerClient, err := SetupWorker(&cfg.RedisWorker, jobManager)
	if err != nil {
		panic(err)
	}

	return &Execution{
		kafkaReader:  kafkaReader,
		kafkaWriter:  kafkaWriter,
		jobManager:   jobManager,
		workerServer: workerServer,
		workerMux:    workerMux,
		workerClient: workerClient,
	}
}

func (e *Execution) Stop() {
	e.workerClient.Close()
	e.workerServer.Shutdown()
}

func (e *Execution) Run(ctx context.Context) error {
	zapLogger, _ := zap.NewProduction()
	defer func() {
		_ = zapLogger.Sync()
	}()
	log := zapLogger.Sugar()
	return e.kafkaReader.Consume(ctx, func(ctx context.Context, message []byte) error {
		log.Infof("starting to scan for request: %s", message)
		var req models.ScanRequestMessage
		err := json.Unmarshal(message, &req)
		if err != nil {
			log.Warnf("failed to unmarshal message, err: %+v", err)
			return err
		}

		scanSourceCodejob, err := e.jobManager.NewScanSourceCodeJob(req.ScanID, req.Owner, req.Repository)
		if err != nil {
			log.Warnf("failed to create job: %v", err)
			return err
		}
		jobInfo, err := e.workerClient.Enqueue(scanSourceCodejob)
		if err != nil {
			log.Warnf("failed to enqueue job: %v", err)
			return err
		}
		log.Infof("enqueued job: id=%s queue=%s", jobInfo.ID, jobInfo.Queue)

		return nil
	})
}
