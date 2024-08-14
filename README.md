# Blockbook

## Packages:

1. `pkg/bccclient`: Contains an interface for a blockchain client that can be used for interacting with blockchains through RPC.
2. `pkg/bccclient/eth`: An implementation of the `pkg/bccclient` for the ETH blockchain based on the `go-ethereum` pkg.
3. `pkg/bcparser`: Contains an interface for a blockchain parser.
4. `pkg/bcparser/bcc`: An in-memory implementation of `pkg/bcparser` which uses `pkg/bccclient` for interacting with the blockchain.
5. `pkg/controller`: Some handy helpers for writing REST controllers based on Gin.
6. `pkg/errors`: A custom error struct with some extra features like error type and status code.
7. `pkg/logging`: Some helpers for working with Zap logger.
8. `pkg/set`: Set can be used to check if a given key exists in a set or not. It uses a map with an empty struct as values to prevent extra memory allocations.
9. `internal/api`: REST api implementation for the blockchain parser.
10. `internal/config`: Project configuration parsing.

## Commands:

1. `make depedency`: Installs required dependencies, Including `golangci-lint` and `ganache-cli`.
2. `make lint`: Runs `golangci-lint` on the project.
3. `make test`: Runs project test suites. This command automatically starts the `ganache-cli` since this is required for integration tests.

## Configuration

The application can be configured using a yaml configuration file or environment variables.

### Environment Variables

The configuration struct is designed to be flexible and can be customized using environment variables. The table below lists the environment variables corresponding to each field in the struct:

| Field                      | Environment Variable        | Default Value           |
|----------------------------|-----------------------------|-------------------------|
| `Environment`              | `ENVIRONMENT`               | `development`           |
| `Api.Server.Addr`          | `API_SERVER_ADDR`           | `:8080`                 |
| `Parser.Client.RpcAddress` | `PARSER_CLIENT_RPC_ADDRESS` | `http://127.0.0.1:8545` |
| `Parser.IndexInterval`     | `PARSER_INDEX_INTERVAL`     | `10s`                   |
| `GracefulShutdownTimeout`  | `GRACEFUL_SHUTDOWN_TIMEOUT` | `30s`                   |

### Configuration File

Below are sample configurations in YAML format for different scenarios:

```yaml
environment: development
api:
  server:
    addr: ":8080"
parser:
  client:
    rpcAddress: "https://eth-mainnet.public.blastapi.io"
  indexInterval: 10s
gracefulShutdownTimeout: 30s
```

## Executables:

1. `cmd/main.go`: Main entrypoint for the project. This file starts the blockchain parser and its REST api server. Pass configuration file using the `-configPath` flag: `go run cmd/main.go -configPath config.yml`

## API:
1. `GET /public/api/v1/block/current`: Returns the latest indexed block number.
2. `POST /public/api/v1/address/subscribe`: Adds an address to the watchlist.
3. `DELETE /public/api/v1/address/unsubscribe`: Removes an address from the watchlist.
4. `GET /public/api/v1/address/:address/transactions`: Returns last 100 transactions for a given address.
5. `GET /metrics`: Returns Prometheus metrics.
6. `GET /-/ready` and `GET /-/live`: Health checks.
7. `/debug/pprof`: Pprof endpoints for debugging.

[Postman collection for public endpoints](https://api.postman.com/collections/33040356-a2813210-110a-42f7-9b6f-e7724b2eabf2?access_key=PMAT-01J581JRQAQG2ZNW0ZSGVHHKFX)