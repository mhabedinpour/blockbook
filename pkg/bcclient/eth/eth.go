package ethclient

import (
	"blockbook/pkg/bcclient"
	"blockbook/pkg/errors"
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client is an implementation of `bcclient.Client` using `go-ethereum` pkg.
type Client struct {
	cli *ethclient.Client
}

// This piece of code is to ensure that a type implements a certain interface at compile time.
// More info: https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ bcclient.Client = (*Client)(nil)

func (c Client) CurrentBlockNumber(ctx context.Context) (uint64, error) {
	num, err := c.cli.BlockNumber(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "could not get current block number")
	}

	return num, nil
}

func (c Client) Block(ctx context.Context, number uint64) (bcclient.Block, error) {
	block, err := c.cli.BlockByNumber(ctx, big.NewInt(int64(number)))
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return bcclient.Block{}, bcclient.ErrBlockNotFound
		}

		return bcclient.Block{}, errors.Wrap(err, "could not get block by number")
	}

	txs := make([]*bcclient.Transaction, 0, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		if tx.To() == nil || tx.Value() == nil {
			continue
		}

		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			continue
		}

		txs = append(txs, &bcclient.Transaction{
			Hash:        tx.Hash().String(),
			FromAddress: from.String(),
			ToAddress:   tx.To().String(),
			Amount:      tx.Value(),
			CreatedAt:   tx.Time(),
		})
	}

	return bcclient.Block{
		Number:       block.NumberU64(),
		Transactions: txs,
	}, nil
}

func New(rpcAddress string) (Client, error) {
	cli, err := ethclient.Dial(rpcAddress)
	if err != nil {
		return Client{}, errors.Wrap(err, "could not create eth rpc client")
	}

	return Client{
		cli: cli,
	}, nil
}
