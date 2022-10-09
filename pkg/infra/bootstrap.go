package infra

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/getsentry/sentry-go"

	"go.uber.org/zap"

	"github.com/vumanhcuongit/scan/internal/config"
)

func setUpSentry(cfg *config.EnvConfig) error {
	if cfg.IsLocalEnvironment() {
		return nil
	}

	return sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.Environment,
	})
}

// ConfigApplication ...
func ConfigApplication(cfg *config.EnvConfig) {
	configBytes, err := json.MarshalIndent(cfg, "", "   ")
	if err != nil {
		zap.S().Warnf("could not convert config to JSON: %v", err)
	} else {
		zap.S().Infof("load config %s", string(configBytes))
	}

	// set up sentry
	err = setUpSentry(cfg)
	CheckError(err)
}

// CheckError log error and panic
func CheckError(err error) {
	if err != nil {
		zap.S().Error("Application Init Error ", err)
		panic(err)
	}
}

// WaitOSSignal ...
func WaitOSSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	s := <-c
	zap.S().Infof("Receive os.Signal: %s", s.String())
}
