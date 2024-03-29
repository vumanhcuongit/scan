package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vumanhcuongit/scan/internal/services/api"
	"go.uber.org/zap"
)

func (h *Handler) createScan(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := zap.S()
	var req = &api.TriggerScanRequest{}
	if err := ginCtx.ShouldBindJSON(req); err != nil {
		log.Warnf("failed to parse request, error: %v", err.Error())
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	scan, err := h.scanService.TriggerScan(ctx, req)
	if err != nil {
		h.ReturnError(ginCtx, err)
		return
	}

	h.ReturnData(ginCtx, scan)
}

func (h *Handler) listScans(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := zap.S()
	var req = &api.ListScansRequest{}
	if err := ginCtx.ShouldBindQuery(req); err != nil {
		log.Warnf("failed to parse request, error: %v", err.Error())
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if req.Size == 0 {
		req.Size = 20
	}

	scans, err := h.scanService.ListScans(ctx, req)
	if err != nil {
		h.ReturnError(ginCtx, err)
		return
	}

	h.ReturnData(ginCtx, scans)
}
