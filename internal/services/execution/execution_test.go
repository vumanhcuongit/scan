package execution

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/pkg/kafka"
)

func TestNewExecution(t *testing.T) {
	cfg, err := config.Load("")
	if err != nil {
		require.NoError(t, err)
	}
	execution := New(cfg, &kafka.Reader{}, &kafka.Writer{})
	require.NotNil(t, execution)
}

func TestStop(t *testing.T) {
	cfg, err := config.Load("")
	if err != nil {
		require.NoError(t, err)
	}
	execution := New(cfg, &kafka.Reader{}, &kafka.Writer{})
	require.NotNil(t, execution)

	execution.Stop()
}
