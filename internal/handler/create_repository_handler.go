package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/vumanhcuongit/scan/internal/services/api"
)

func (h *Handler) createRepository(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := ctxzap.Extract(ctx).Sugar()
	var req = &api.CreateRepositoryRequest{}
	if err := ginCtx.ShouldBindJSON(req); err != nil {
		log.Warnf("failed to parse request, error: %v", err.Error())
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	repository, err := h.scanService.CreateRepository(ctx, req)
	if err != nil {
		h.ReturnError(ginCtx, err)
		return
	}

	h.ReturnData(ginCtx, repository)
}
