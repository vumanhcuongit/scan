package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/runtime/middleware"
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
	// documentation for developers
	opts := middleware.SwaggerUIOpts{SpecURL: "./swagger.yaml"}
	swagger := middleware.SwaggerUI(opts, nil)
	router.StaticFile("/swagger.yaml", "api/swagger.yaml")
	router.GET("/docs", gin.WrapH(swagger))

	apiGroup := router.Group("api")
	// repositories
	apiGroup.POST("/repositories", h.createRepository)
	apiGroup.GET("/repositories", h.listRepositories)
	apiGroup.GET("/repositories/:id", h.getRepository)
	apiGroup.PATCH("/repositories/:id", h.updateRepository)
	apiGroup.DELETE("/repositories/:id", h.deleteRepository)

	// scans
	apiGroup.POST("/scans", h.createScan)
	apiGroup.GET("/scans", h.listScans)
}

func (h *Handler) ReturnData(ginCtx *gin.Context, data interface{}) {
	resp := gin.H{}
	if data != nil {
		resp["data"] = data
	}

	ginCtx.JSON(http.StatusOK, resp)
}

func (h *Handler) ReturnNoConent(ginCtx *gin.Context) {
	ginCtx.Status(http.StatusNoContent)
}

func (h *Handler) ReturnError(ginCtx *gin.Context, err error) {
	errInfo := pkgerrors.NewErrorInfo(err)

	ginCtx.JSON(http.StatusOK, gin.H{
		"data":  nil,
		"error": errInfo,
	})
}
