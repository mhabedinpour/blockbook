package bccparser

import (
	"blockbook/pkg/bcclient"
	"blockbook/pkg/bcparser"
	"blockbook/pkg/errors"
	"blockbook/pkg/logging"
	"blockbook/pkg/set"
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/cenkalti/backoff/v4"
)

const (
	BackoffMaxInterval    = 30 * time.Second
	BackoffInitialFactor  = 250
	BackoffMaxElapsedTime = 60 * time.Second

	MaxTxsToKeep = 100
)

// Parser is an implementation of `bcparser.Parser` based on `bcclient.Client`.
type Parser struct {
	client              bcclient.Client
	lastIndexedBlock    atomic.Uint64
	subscribedAddresses *set.Set[string]
	// mu is used to synchronize access to transactions.
	mu           sync.RWMutex
	transactions map[string][]*bcclient.Transaction
	logger       *zap.Logger
	// ctxCancel is used by Stop() to stop the indexer goroutine.
	ctxCancel context.CancelFunc
	readyChan chan struct{}
}

// This piece of code is to ensure that a type implements a certain interface at compile time.
// More info: https://github.com/uber-go/guide/blob/master/style.md#verify-interface-compliance
var _ bcparser.Parser = (*Parser)(nil)

func (p *Parser) CurrentBlockNumber() uint64 {
	return p.lastIndexedBlock.Load()
}

func (p *Parser) Subscribe(address string) bool {
	return p.subscribedAddresses.Add(address)
}

func (p *Parser) Unsubscribe(address string) bool {
	return p.subscribedAddresses.Remove(address)
}

func (p *Parser) Transactions(address string) []*bcclient.Transaction {
	if !p.subscribedAddresses.Contains(address) {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	txs, ok := p.transactions[address]
	if !ok {
		// return an empty array to distinguish between when the address is not subscribed and when we just don't have
		// any transactions for a subscribed address yet.
		return make([]*bcclient.Transaction, 0)
	}

	return txs
}

// lookForNewBlocks checks if any new blocks are added to the chain since the last time and index all transactions inside new blocks if required.
func (p *Parser) lookForNewBlocks(ctx context.Context, firstScan bool) error {
	currentBlockNum, err := p.client.CurrentBlockNumber(ctx)
	if err != nil {
		return errors.Wrap(err, "could not get current block number from client")
	}

	lastIndexedBlock := p.lastIndexedBlock.Load()
	blockToIndex := lastIndexedBlock + 1
	// in case of the first scan, Start from the current block.
	if firstScan {
		blockToIndex = currentBlockNum
	}

	// continue indexing until we reach the current block.
	for blockToIndex <= currentBlockNum {
		p.logger.Sugar().Infof("indexing block %d...", blockToIndex)

		block, err := p.client.Block(ctx, blockToIndex)
		if err != nil {
			return errors.Wrap(err, "could not get block from client")
		}

		p.processBlock(block)
		p.logger.Sugar().Infof("block %d indexed", blockToIndex)

		blockToIndex++
	}

	return nil
}

// processBlock stores transactions inside a block in-memory if required.
func (p *Parser) processBlock(block bcclient.Block) {
	defer p.lastIndexedBlock.Store(block.Number)

	// first look for transactions involving subscribed addresses
	watchlist := p.subscribedAddresses.ToSimpleMap() // clone the set as a map to avoid constantly locking-and-unlocking the set mutex in the for loop.
	txToStore := make([]*bcclient.Transaction, 0)
	for _, tx := range block.Transactions {
		_, senderSubscribed := watchlist[tx.FromAddress]
		_, receiverSubscribed := watchlist[tx.ToAddress]
		if !senderSubscribed && !receiverSubscribed {
			continue
		}

		txToStore = append(txToStore, tx)
	}

	if len(txToStore) == 0 {
		return
	}

	// acquire write lock and start indexing transactions in-memory.
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, tx := range txToStore {
		p.transactions[tx.FromAddress] = append(p.transactions[tx.FromAddress], tx)
		if len(p.transactions[tx.FromAddress]) > MaxTxsToKeep {
			p.transactions[tx.FromAddress] = p.transactions[tx.FromAddress][len(p.transactions[tx.FromAddress])-MaxTxsToKeep:]
		}

		p.transactions[tx.ToAddress] = append(p.transactions[tx.ToAddress], tx)
		if len(p.transactions[tx.ToAddress]) > MaxTxsToKeep {
			p.transactions[tx.ToAddress] = p.transactions[tx.ToAddress][len(p.transactions[tx.ToAddress])-MaxTxsToKeep:]
		}
	}
}

// startIndexing launches the indexer goroutine which periodically checks for new blocks in the background and indexes transactions in-memory if required.
func (p *Parser) startIndexing(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	firstScan := true
	markReady := func() {
		if !firstScan {
			return
		}

		firstScan = false
		close(p.readyChan)
	}

	err := p.lookForNewBlocks(ctx, true)
	if err != nil {
		p.logger.Error("could not do initial block scan", zap.Error(err))
	} else {
		markReady()
	}

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()

			return

		case <-ticker.C:
			b := &backoff.ExponentialBackOff{
				InitialInterval:     BackoffInitialFactor,
				RandomizationFactor: backoff.DefaultRandomizationFactor,
				Multiplier:          backoff.DefaultMultiplier,
				MaxInterval:         BackoffMaxInterval,
				MaxElapsedTime:      BackoffMaxElapsedTime,
				Stop:                backoff.Stop,
				Clock:               backoff.SystemClock,
			}
			b.Reset()

			_ = backoff.Retry(func() error {
				err := p.lookForNewBlocks(ctx, firstScan)
				if err != nil {
					p.logger.Error("could not scan blocks", zap.Error(err))
				} else {
					markReady()
				}

				return err
			}, b)
		}
	}
}

func (p *Parser) Ready() <-chan struct{} {
	return p.readyChan
}

// Stop terminates the indexer goroutine by cancelling its context.
func (p *Parser) Stop() {
	p.ctxCancel()
}

func New(logger *zap.Logger, client bcclient.Client, indexInterval time.Duration) *Parser {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Parser{
		client:              client,
		logger:              logging.AddComponent(logger, "block-parser"),
		ctxCancel:           cancel,
		subscribedAddresses: set.New[string](),
		transactions:        make(map[string][]*bcclient.Transaction),
		readyChan:           make(chan struct{}),
	}

	go p.startIndexing(ctx, indexInterval)

	return p
}
