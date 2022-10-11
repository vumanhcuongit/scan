package api

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/internal/services/base"
	"github.com/vumanhcuongit/scan/pkg/kafka"
)

func TestNewScanService(t *testing.T) {
	cfg, err := config.Load("")
	if err != nil {
		require.NoError(t, err)
	}
	baseService := &base.Service{}
	baseService.SetConfig(cfg)
	scanService := NewScanService(baseService, &kafka.Writer{}, &kafka.Reader{})
	require.NotNil(t, scanService)
}
