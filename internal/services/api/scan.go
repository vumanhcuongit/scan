package api

import (
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/vumanhcuongit/scan/internal/services/worker"
)

type CreateScanRequest struct{}

func (s *ScanService) CreateScan(ctx context.Context, request *CreateScanRequest) error {
	log := ctxzap.Extract(ctx).Sugar()
	log.Infof("start to create a scan with request %+v", request)

	message, err := json.Marshal(worker.ScanRequest{
		Owner:      "vumanhcuongit",
		Repository: "workshop",
	})
	if err != nil {
		log.Errorf("failed to marshal message, err: %+v", err)
		return err
	}

	err = s.kafkaWriter.WriteMessage(ctx, message)
	if err != nil {
		log.Errorf("failed to write message to queue, err: %+v", err)
		return err
	}

	return nil
}
