package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vumanhcuongit/scan/internal/config"
	"github.com/vumanhcuongit/scan/internal/handler"
	"github.com/vumanhcuongit/scan/internal/services/api"
	"github.com/vumanhcuongit/scan/internal/services/base"
	"github.com/vumanhcuongit/scan/pkg/kafka"
	"go.uber.org/zap"
)

type Server struct {
	apiSvc     api.IScanService
	cfg        *config.App
	router     *gin.Engine
	httpServer *http.Server
}

func NewServer(cfg *config.App, kafkaWriter kafka.IWriter, kafkaReader *kafka.Reader) *Server {
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	bs := base.NewService(cfg)
	router := gin.New()
	return &Server{
		cfg:    cfg,
		apiSvc: api.NewScanService(bs, kafkaWriter, kafkaReader),
		router: router,
	}
}

func (s *Server) initPing() {
	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"time_now": time.Now().String(),
		})
	})
}

func (s *Server) Start() error {
	return s.apiSvc.Start(context.Background())
}

// Listen listen on tcp port and serve http server.
func (s *Server) Listen() error {
	s.initPing()
	apiHandler := handler.NewHandler(s.apiSvc)
	apiHandler.Register(s.router)
	s.httpServer = &http.Server{
		Handler: s.router,
		Addr:    s.cfg.HTTPAddr,
	}

	return s.httpServer.ListenAndServe()
}
