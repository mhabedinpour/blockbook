package bcclient

import (
	"context"
	"math/big"
	"time"
)

// Transaction is a single transaction inside a blockchain block.
type Transaction struct {
	Hash        string    `json:"hash"`
	FromAddress string    `json:"fromAddress"`
	ToAddress   string    `json:"toAddress"`
	Amount      *big.Int  `json:"amount"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Block represents a block on a blockchain network.
type Block struct {
	Number       uint64
	Transactions []*Transaction
}

// Client can be used to interact with different blockchains through RPC calls.
type Client interface {
	CurrentBlockNumber(ctx context.Context) (uint64, error)
	Block(ctx context.Context, number uint64) (Block, error)
}
