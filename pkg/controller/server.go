package controller

import (
	"net/http"
	"time"

	"blockbook/pkg/logging"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Options struct {
	Addr               string
	ReadTimeout        time.Duration
	ReadHeaderTimeout  time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	RequestTimeout     time.Duration
	MaxHeaderBytes     int
	MetricsPath        string
	MetricsSubSystem   string
	DefaultHandlerName string
}

// NewServer can be used to create a HTTP server based on GIN with some extra features like access log, prometheus metrics, pprof handlers, error handling, etc.
func NewServer(logger *zap.Logger, apiControllers map[string]Controller, options Options) (*http.Server, error) {
	if !logging.IsDebug(logger) {
		gin.SetMode(gin.ReleaseMode)
	}

	e := gin.New()

	pprof.Register(e)

	err := setupValidator()
	if err != nil {
		return nil, err
	}

	e.Use(AddLogger(logger))
	e.Use(AccessLog())
	e.Use(RequestTimeout(options.RequestTimeout))
	e.Use(Prometheus(e, options.MetricsPath, options.MetricsSubSystem, options.DefaultHandlerName, apiControllers))

	for _, ctrl := range apiControllers {
		ctrl.RegisterHandlers(e.Group(ctrl.PathPrefix()))
	}

	return &http.Server{
		Addr:              options.Addr,
		Handler:           e,
		ReadTimeout:       options.ReadTimeout,
		ReadHeaderTimeout: options.ReadHeaderTimeout,
		WriteTimeout:      options.WriteTimeout,
		IdleTimeout:       options.IdleTimeout,
		MaxHeaderBytes:    options.MaxHeaderBytes,
	}, nil
}
