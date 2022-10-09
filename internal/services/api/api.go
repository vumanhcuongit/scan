package api

import (
	"github.com/vumanhcuongit/scan/internal/services/base"
	"github.com/vumanhcuongit/scan/pkg/kafka"
)

type ScanService struct {
	base.Service
	kafkaWriter *kafka.Writer
}

func NewScanService(bs *base.Service, kafkaWriter *kafka.Writer) *ScanService {
	return &ScanService{
		Service:     *bs,
		kafkaWriter: kafkaWriter,
	}
}
