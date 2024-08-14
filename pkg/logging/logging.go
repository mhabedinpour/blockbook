package logging

import (
	"log"

	"go.uber.org/zap"
)

const (
	DevelopmentEnvironment = "development"
	ProductionEnvironment  = "production"
)

func NewLogger(environment string) *zap.Logger {
	var logger *zap.Logger
	if environment == ProductionEnvironment {
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			log.Fatal("could not create production logger")
		}
	} else {
		logger, _ = zap.NewDevelopment()
	}

	return logger
}

func AddComponent(parentLogger *zap.Logger, componentName string) *zap.Logger {
	return parentLogger.With(zap.String(componentName, "component"))
}

func IsDebug(logger *zap.Logger) bool {
	return logger.Level() == zap.DebugLevel
}
