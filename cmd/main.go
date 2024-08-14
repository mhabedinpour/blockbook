package main

import (
	"blockbook/internal/api"
	"blockbook/internal/config"
	ethclient "blockbook/pkg/bcclient/eth"
	bccparser "blockbook/pkg/bcparser/bcc"
	"blockbook/pkg/controller"
	"blockbook/pkg/errors"
	"blockbook/pkg/logging"
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("configPath", "config.yml", "The config file path")
	flag.Parse()

	log.Println("loading config ...")
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal("could not load config file", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := logging.NewLogger(cfg.Environment)
	logger.Debug("loaded config file successfully.")

	defer func() {
		if p := recover(); p != nil {
			logger.Error("captured panic, exiting...", zap.Any("panic", p))
			_ = logger.Sync()

			panic(p)
		}
	}()

	logger.Info("creating blockchain rpc client...")
	bcClient, err := ethclient.New(cfg.Parser.Client.RpcAddress)
	if err != nil {
		logger.Fatal("could not create blockchain rpc client", zap.Error(err))
	}
	logger.Debug("blockchain rpc client created successfully")

	logger.Info("creating blockchain parser...")
	parser := bccparser.New(logger, bcClient, cfg.Parser.IndexInterval)
	logger.Debug("blockchain parser created successfully")

	logger.Info("creating webserver...")
	server, err := api.NewServer(logger, api.Options{
		Controller: controller.Options{
			Addr:               cfg.Api.Server.Addr,
			ReadTimeout:        cfg.Api.Server.ReadTimeout,
			ReadHeaderTimeout:  cfg.Api.Server.ReadHeaderTimeout,
			WriteTimeout:       cfg.Api.Server.WriteTimeout,
			IdleTimeout:        cfg.Api.Server.IdleTimeout,
			RequestTimeout:     cfg.Api.Server.RequestTimeout,
			MaxHeaderBytes:     cfg.Api.Server.MaxHeaderBytes,
			MetricsPath:        cfg.Api.Server.MetricsPath,
			MetricsSubSystem:   cfg.Api.Server.MetricsSubSystem,
			DefaultHandlerName: cfg.Api.Server.DefaultHandlerName,
		},
		BlockchainParser: parser,
	})
	if err != nil {
		logger.Fatal("could not create webserver", zap.Error(err))
	}
	logger.Debug("webserver created successfully")

	go func() {
		logger.Sugar().Infof("starting webserver on address %s ...", cfg.Api.Server.Addr)
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error("webserver failed", zap.Error(err))
			}
		}
	}()

	<-ctx.Done() // Waiting for the interrupt

	logger.Debug("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
	defer cancel()

	logger.Info("stopping webserver...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("webserver failed to shut down in time", zap.Error(err))
	}

	logger.Info("stopping blockchain parser...")
	parser.Stop()

	logger.Debug("shut down successfully")
}
