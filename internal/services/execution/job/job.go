package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/vumanhcuongit/scan/pkg/gitscan"
	"github.com/vumanhcuongit/scan/pkg/kafka"
	"github.com/vumanhcuongit/scan/pkg/models"
	"go.uber.org/zap"
)

const (
	TypeScanSourceCode = "scan_source_code"
)

type Job struct {
	gitScan     *gitscan.GitScan
	kafkaWriter *kafka.Writer
}

func NewJob(sourcesCodeDir string, kafkaWriter *kafka.Writer) *Job {
	return &Job{
		gitScan:     gitscan.NewGitScan(sourcesCodeDir),
		kafkaWriter: kafkaWriter,
	}
}

type ScanSourceCodePayload struct {
	ScanID    int64
	OwnerName string
	RepoName  string
}

func (j *Job) NewScanSourceCodeJob(scanID int64, ownerName string, repoName string) (*asynq.Task, error) {
	payload, err := json.Marshal(ScanSourceCodePayload{
		ScanID:    scanID,
		OwnerName: ownerName,
		RepoName:  repoName,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeScanSourceCode, payload), nil
}

func (j *Job) HandleScanSourceCodeJob(ctx context.Context, t *asynq.Task) error {
	log := zap.S()
	var payload ScanSourceCodePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	fmt.Println("starting to scan source code with payload")
	log.Infof("starting to scan source code with payload %+v", payload)

	err := j.produceInProgressResultMessage(ctx, &payload)
	if err != nil {
		log.Infof("failed to produce in progress message, err: +%v", err)
		return err
	}

	findings, err := j.gitScan.Scan(ctx, payload.OwnerName, payload.RepoName)
	if err != nil {
		err := j.produceFailedResultMessage(ctx, &payload)
		if err != nil {
			log.Infof("failed to produce in progress message, err: +%v", err)
		}
		return err
	} else {
		err := j.produceSuccessfulResultMessage(ctx, &payload, findings)
		if err != nil {
			log.Infof("failed to produce succesful message, err: +%v", err)
			return err
		}
	}

	return nil
}

func (j *Job) produceInProgressResultMessage(ctx context.Context, payload *ScanSourceCodePayload) error {
	log := zap.S()
	timeNow := time.Now()
	message, err := json.Marshal(models.ScanResultMessage{
		ScanID:     payload.ScanID,
		ScanStatus: models.ScanStatusInProgress,
		ScanningAt: &timeNow,
	})
	if err != nil {
		log.Warnf("failed to marshal message, err: %+v", err)
		return err
	}

	err = j.doWriteMessage(ctx, message)
	if err != nil {
		log.Warnf("failed to write message to queue, err: %+v", err)
		return err
	}

	return nil
}

func (j *Job) produceSuccessfulResultMessage(
	ctx context.Context,
	payload *ScanSourceCodePayload,
	findings []models.Finding,
) error {
	log := zap.S()

	var findingRepoJSON []byte
	var marshalErr error
	if len(findings) > 0 {
		findingRepoJSON, marshalErr = json.Marshal(findings)
		if marshalErr != nil {
			log.Warnf("failed to marshal finding report, err: %+v", marshalErr)
			return marshalErr
		}
	}

	timeNow := time.Now()
	message, err := json.Marshal(models.ScanResultMessage{
		ScanID:     payload.ScanID,
		ScanStatus: models.ScanStatusSuccess,
		FinishedAt: &timeNow,
		Findings:   findingRepoJSON,
	})
	if err != nil {
		log.Warnf("failed to marshal message, err: %+v", err)
		return err
	}

	err = j.doWriteMessage(ctx, message)
	if err != nil {
		log.Warnf("failed to write message to queue, err: %+v", err)
		return err
	}

	return nil
}

func (j *Job) produceFailedResultMessage(ctx context.Context, payload *ScanSourceCodePayload) error {
	log := zap.S()

	timeNow := time.Now()
	message, err := json.Marshal(models.ScanResultMessage{
		ScanID:     payload.ScanID,
		ScanStatus: models.ScanStatusFailure,
		FinishedAt: &timeNow,
	})
	if err != nil {
		log.Warnf("failed to marshal message, err: %+v", err)
		return err
	}

	err = j.doWriteMessage(ctx, message)
	if err != nil {
		log.Warnf("failed to write message to queue, err: %+v", err)
		return err
	}

	return nil
}

func (j *Job) doWriteMessage(ctx context.Context, message []byte) error {
	err := j.kafkaWriter.WriteMessage(ctx, message)
	if err != nil {
		return err
	}
	return nil
}
