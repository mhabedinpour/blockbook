package config

import (
	"time"

	"blockbook/pkg/errors"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment string `env:"ENVIRONMENT" env-default:"development" yaml:"environment"`
	Api         struct {
		Server struct {
			Addr               string        `env:"API_SERVER_ADDR" env-default:":8080" yaml:"addr"`
			ReadTimeout        time.Duration `env:"API_SERVER_READ_TIMEOUT" env-default:"1m" yaml:"readTimeout"`
			ReadHeaderTimeout  time.Duration `env:"API_SERVER_READ_HEADER_TIMEOUT" env-default:"10s" yaml:"readHeaderTimeout"`
			WriteTimeout       time.Duration `env:"API_SERVER_WRITE_TIMEOUT" env-default:"2m" yaml:"writeTimeout"`
			IdleTimeout        time.Duration `env:"API_SERVER_IDLE_TIMEOUT" env-default:"2m" yaml:"idleTimeout"`
			RequestTimeout     time.Duration `env:"API_SERVER_REQUEST_TIMEOUT" env-default:"10s" yaml:"requestTimeout"`
			MaxHeaderBytes     int           `env:"API_SERVER_MAX_HEADER_BYTES" env-default:"0" yaml:"maxHeaderBytes"`
			MetricsPath        string        `env:"API_SERVER_METRICS_PATH" env-default:"/metrics" yaml:"metricsPath"`
			MetricsSubSystem   string        `env:"API_SERVER_METRICS_SUBSYSTEM" env-default:"api" yaml:"metricsSubSystem"`
			DefaultHandlerName string        `env:"API_SERVER_DEFAULT_HANDLER_NAME" env-default:"api-unknown" yaml:"defaultHandlerName"`
		} `yaml:"server"`
	} `yaml:"api"`
	Parser struct {
		Client struct {
			RpcAddress string `env:"PARSER_CLIENT_RPC_ADDRESS" env-default:"http://127.0.0.1:8545" yaml:"rpcAddress"`
		} `yaml:"client"`
		IndexInterval time.Duration `env:"PARSER_INDEX_INTERVAL" env-default:"10s" yaml:"indexInterval"`
	} `yaml:"parser"`
	GracefulShutdownTimeout time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT" env-default:"30s" yaml:"gracefulShutdownTimeout"`
}

// Load receives the path for yaml config file and returns a filled Config struct.
func Load(configPath string) (Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return Config{}, errors.Wrap(err, "could not read config")
	}

	return cfg, nil
}
