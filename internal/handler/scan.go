package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/vumanhcuongit/scan/internal/services/api"
)

func (h *Handler) createScan(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := ctxzap.Extract(ctx).Sugar()
	var req = &api.CreateScanRequest{}
	if err := ginCtx.ShouldBindJSON(req); err != nil {
		log.Warnf("failed to parse request, error: %v", err.Error())
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err := h.scanService.CreateScan(ctx, &api.CreateScanRequest{})
	if err != nil {
		_ = ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ginCtx.JSON(http.StatusOK, map[string]string{"success": "true"})
}
