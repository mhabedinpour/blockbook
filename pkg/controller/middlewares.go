package controller

import (
	"context"
	"strings"
	"time"

	"github.com/Depado/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	metricsNamespace = "blockbook"
)

func AddLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := uuid.New().String()
		logger := logger.With(zap.String("requestId", requestId), zap.String("path", c.FullPath()), zap.String("method", c.Request.Method))
		c.Set(LoggerName, logger)
		c.Next()
	}
}

func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := GetLogger(c)

		start := time.Now()
		c.Next()

		latency := time.Since(start)

		logger.Info("Access log",
			zap.Int("status_code", c.Writer.Status()),
			zap.Float64("latency", latency.Seconds()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("url", c.Request.URL.String()),
			zap.String("referer", c.Request.Referer()),
			zap.String("method", c.Request.Method),
		)
	}
}

func Prometheus(e *gin.Engine, metricsPath, subsystem, defaultHandlerName string, apiControllers map[string]Controller) gin.HandlerFunc {
	p := ginprom.New(
		ginprom.Engine(e),
		ginprom.Path(metricsPath),
		ginprom.Namespace(metricsNamespace),
		ginprom.Subsystem(subsystem),
		ginprom.HandlerNameFunc(func(c *gin.Context) string {
			path := c.FullPath()
			for handlerName, ctrl := range apiControllers {
				if strings.Contains(path, ctrl.PathPrefix()) {
					return handlerName
				}
			}

			return defaultHandlerName
		}),
	)

	return p.Instrument()
}
