package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vumanhcuongit/scan/internal/services/api"
	"go.uber.org/zap"
)

func (h *Handler) createRepository(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := zap.S()
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

func (h *Handler) getRepository(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := zap.S()
	repositoryIDStr := ginCtx.Param("id")
	if repositoryIDStr == "" {
		log.Warnf("missing repository id")
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	repositoryID, err := strconv.ParseInt(repositoryIDStr, 10, 64)
	if err != nil {
		log.Warnf("invalid repository id, err: %+v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	repository, err := h.scanService.GetRepository(ctx, repositoryID)
	if err != nil {
		h.ReturnError(ginCtx, err)
		return
	}

	h.ReturnData(ginCtx, repository)
}

func (h *Handler) listRepositories(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := zap.S()
	var req = &api.ListRepositoriesRequest{}
	if err := ginCtx.ShouldBindQuery(req); err != nil {
		log.Warnf("failed to parse request, error: %v", err.Error())
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if req.Size == 0 {
		req.Size = 20
	}

	repositories, err := h.scanService.ListRepositories(ctx, req)
	if err != nil {
		h.ReturnError(ginCtx, err)
		return
	}

	h.ReturnData(ginCtx, repositories)
}

func (h *Handler) updateRepository(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := zap.S()

	repositoryIDStr := ginCtx.Param("id")
	if repositoryIDStr == "" {
		log.Warnf("missing repository id")
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	repositoryID, err := strconv.ParseInt(repositoryIDStr, 10, 64)
	if err != nil {
		log.Warnf("invalid repository id, err: %+v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var req = &api.UpdateRepositoryRequest{}
	if err := ginCtx.ShouldBindJSON(req); err != nil {
		log.Warnf("failed to parse request, error: %v", err.Error())
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	repository, err := h.scanService.UpdateRepository(ctx, repositoryID, req)
	if err != nil {
		h.ReturnError(ginCtx, err)
		return
	}

	h.ReturnData(ginCtx, repository)
}

func (h *Handler) deleteRepository(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()
	log := zap.S()
	repositoryIDStr := ginCtx.Param("id")
	if repositoryIDStr == "" {
		log.Warnf("missing repository id")
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	repositoryID, err := strconv.ParseInt(repositoryIDStr, 10, 64)
	if err != nil {
		log.Warnf("invalid repository id, err: %+v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = h.scanService.DeleteRepository(ctx, repositoryID)
	if err != nil {
		h.ReturnError(ginCtx, err)
		return
	}

	h.ReturnNoConent(ginCtx)
}
