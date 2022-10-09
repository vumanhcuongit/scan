package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/vumanhcuongit/scan/internal/services/api"
)

type Handler struct {
	scanService *api.ScanService
}

func NewHandler(scanService *api.ScanService) *Handler {
	return &Handler{
		scanService: scanService,
	}
}

func (h *Handler) Register(router gin.IRouter) {
	apiGroup := router.Group("api")
	apiGroup.POST("/scans/:repository_id", h.createScan)
}

func (h *Handler) createScan(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := ctxzap.Extract(ctx).Sugar()

	log.Infof("starting to scan")
	err := h.scanService.CreateScan(ctx, &api.CreateScanRequest{})
	if err != nil {
		_ = ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ginCtx.JSON(http.StatusOK, map[string]string{"success": "true"})
}
