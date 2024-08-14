package api

import (
	"blockbook/internal/api/controllers/address"
	"blockbook/internal/api/controllers/block"
	"blockbook/internal/api/controllers/health"
	"blockbook/pkg/bcparser"
	"blockbook/pkg/controller"
	"net/http"

	"go.uber.org/zap"
)

type Options struct {
	Controller       controller.Options
	BlockchainParser bcparser.Parser
}

func NewServer(logger *zap.Logger, options Options) (*http.Server, error) {
	apiV1Group := controller.NewGroup("/api/v1",
		block.New(options.BlockchainParser),
		address.New(options.BlockchainParser),
	)
	publicGroup := controller.NewGroup("/public", apiV1Group)

	apiControllers := map[string]controller.Controller{
		"public": publicGroup,
		"health": health.New(),
	}

	return controller.NewServer(logger, apiControllers, options.Controller)
}
