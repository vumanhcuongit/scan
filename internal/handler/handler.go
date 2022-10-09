package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vumanhcuongit/scan/internal/services/api"
	pkgerrors "github.com/vumanhcuongit/scan/pkg/errors"
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
	apiGroup.POST("/repositories", h.createRepository)
	apiGroup.POST("/scans", h.createScan)
}

func (h *Handler) ReturnData(ginCtx *gin.Context, data interface{}) {
	resp := gin.H{}
	if data != nil {
		resp["data"] = data
	}

	ginCtx.JSON(http.StatusOK, resp)
}

func (h *Handler) ReturnError(ginCtx *gin.Context, err error) {
	errInfo := pkgerrors.NewErrorInfo(err)

	ginCtx.JSON(http.StatusOK, gin.H{
		"data":  nil,
		"error": errInfo,
	})
}
