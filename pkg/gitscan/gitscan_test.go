package gitscan

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScanSourceCode(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	sourceCodesDir := "./source_codes"
	os.RemoveAll(sourceCodesDir)
	gitScanSrv := NewGitScan(sourceCodesDir)
	findings, err := gitScanSrv.Scan(ctx, "vumanhcuongit", "workshop")
	require.NoError(t, err)
	require.Equal(t, 3, len(findings))
	os.RemoveAll(sourceCodesDir)
}
