package execution

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/internal/services/execution/job"
	"go.uber.org/zap"
)

func SetupWorker(cfg *config.RedisWorkerConfig, jobManager *job.Job) (*asynq.Server, *asynq.ServeMux, *asynq.Client, error) {
	log := zap.S()
	redisClientOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		panic(err)
	}

	workerClient := asynq.NewClient(redisClientOpt)
	workerServer := asynq.NewServer(
		redisClientOpt,
		asynq.Config{
			Concurrency: int(cfg.TotalConcurrencyWorkers),
		},
	)

	workerMux := asynq.NewServeMux()
	workerMux.Use(setupLog())
	workerMux.HandleFunc(job.TypeScanSourceCode, jobManager.HandleScanSourceCodeJob)

	if err := workerServer.Start(workerMux); err != nil {
		log.Fatalf("failed to run server: %+v", err)
		return nil, nil, nil, err
	}

	return workerServer, workerMux, workerClient, nil
}

func setupLog() asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			logger := zap.L()
			payload := string(t.Payload())
			logger.Sugar().Infow("starting task", "type", t.Type(), "payload", payload)
			err := next.ProcessTask(ctx, t)
			if err != nil {
				maxRetry, _ := asynq.GetMaxRetry(ctx)
				retried, _ := asynq.GetRetryCount(ctx)
				logger.Sugar().Warnw(
					"Retry task", "type", t.Type(), "error", err, "retried", retried, "max_retry", maxRetry,
				)
				return err
			}

			logger.Sugar().Infow("Finished task", "type", t.Type())
			return nil
		})
	}
}
