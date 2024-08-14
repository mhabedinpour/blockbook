package api

import (
	"blockbook/pkg/bcclient"
	ethclient "blockbook/pkg/bcclient/eth"
	bccparser "blockbook/pkg/bcparser/bcc"
	"blockbook/pkg/errors"
	"bytes"
	"context"
	"encoding/json"
	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	goeth "github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"math/big"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	parserRefreshInterval = 5 * time.Second
	ganacheRpcAddress     = "http://localhost:8545"

	wallet1PublicAddress = "0x627306090abaB3A6e1400e9345bC60c78a8BEf57"
	wallet1PrivateKey    = "c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3"

	wallet2PublicAddress = "0xf17f52151EbEF6C7334FAD080c5704D77216b732"
	wallet2PrivateKey    = "ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f"

	maxTransactionsNum = 10

	newBlockCheckInterval = 1 * time.Second
	newBlockCheckTimeout  = 10 * time.Second
	newBlocksThreshold    = 1
)

type apiResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Result  T      `json:"result"`
}

type currentBlockResponse struct {
	LastIndexedBlock uint64 `json:"lastIndexedBlock"`
}

type subscribeRequestBody struct {
	Address string `json:"address"`
}

type transactionsResponse struct {
	Transactions []bcclient.Transaction `json:"transactions"`
}

type randomTx struct {
	fromWallet1 bool
	amount      *big.Int
}

func setupServer(logger *zap.Logger) (http.Handler, func()) {
	bcClient, err := ethclient.New(ganacheRpcAddress)
	if err != nil {
		panic(err)
	}

	parser := bccparser.New(logger, bcClient, parserRefreshInterval)
	server, err := NewServer(logger, Options{
		BlockchainParser: parser,
	})
	if err != nil {
		panic(err)
	}

	logger.Info("waiting for parser to become ready...")
	<-parser.Ready()

	return server.Handler, func() {
		parser.Stop()
	}
}

func parseApiResponse[T any](t *testing.T, rec *httptest.ResponseRecorder, res *apiResponse[T]) {
	assert.Equal(t, 200, rec.Code)

	err := json.Unmarshal(rec.Body.Bytes(), res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
}

func getCurrentBlock(t *testing.T, handler http.Handler) uint64 {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/public/api/v1/block/current", nil)
	if err != nil {
		panic(err)
	}
	handler.ServeHTTP(rec, req)

	var res apiResponse[currentBlockResponse]
	parseApiResponse(t, rec, &res)

	return res.Result.LastIndexedBlock
}

func subscribeAddress(t *testing.T, handler http.Handler, address string) {
	body, err := json.Marshal(subscribeRequestBody{
		Address: address,
	})
	if err != nil {
		panic(err)
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/public/api/v1/address/subscribe", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	handler.ServeHTTP(rec, req)

	var res apiResponse[struct{}]
	parseApiResponse(t, rec, &res)
}

func generateRandomTxs() []randomTx {
	txNum := rand.UintN(maxTransactionsNum) + 1
	txs := make([]randomTx, 0, txNum)

	for range txNum {
		txs = append(txs, randomTx{
			fromWallet1: rand.IntN(2) == 1,
			amount:      big.NewInt(rand.Int64N(1000000000000000000) + 1000000000000000000), // between 1 and 2 eth
		})
	}

	return txs
}

func getTransactions(t *testing.T, handler http.Handler, address string) []bcclient.Transaction {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/public/api/v1/address/"+address+"/transactions", nil)
	if err != nil {
		panic(err)
	}
	handler.ServeHTTP(rec, req)

	var res apiResponse[transactionsResponse]
	parseApiResponse(t, rec, &res)

	return res.Result.Transactions
}

func sendTransaction(t *testing.T, privateHexKey, publicHexKey, toHexAddress string, amount *big.Int) {
	client, err := goeth.Dial(ganacheRpcAddress)
	if err != nil {
		panic(err)
	}

	privateKey, err := crypto.HexToECDSA(privateHexKey)
	if err != nil {
		panic(err)
	}

	fromAddress := common.HexToAddress(publicHexKey)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	assert.NoError(t, err)

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	assert.NoError(t, err)

	toAddress := common.HexToAddress(toHexAddress)
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    amount,
		Gas:      gasLimit,
		GasPrice: gasPrice,
	})

	chainID, err := client.NetworkID(context.Background())
	assert.NoError(t, err)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	assert.NoError(t, err)

	err = client.SendTransaction(context.Background(), signedTx)
	assert.NoError(t, err)
}

func waitForNewBlocks(t *testing.T, logger *zap.Logger, handler http.Handler, lastKnownBlock uint64) uint64 {
	ctx, cancel := context.WithTimeout(context.Background(), newBlockCheckTimeout)
	defer cancel()

	ticker := time.NewTicker(newBlockCheckInterval)
	defer ticker.Stop()

	lastIndexedBlock := lastKnownBlock

	for {
		select {
		case <-ticker.C:
			logger.Info("checking for new blocks...")

			lastIndexedBlock = getCurrentBlock(t, handler)
			if lastIndexedBlock >= lastKnownBlock+newBlocksThreshold {
				logger.Sugar().Infof("founded new indexed block: %d", lastIndexedBlock)

				return lastIndexedBlock
			}
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())

			return lastIndexedBlock
		}
	}
}

func compareTxs(t *testing.T, handler http.Handler, wanted []bcclient.Transaction) bool {
	wallet1Txs := getTransactions(t, handler, wallet1PublicAddress)
	wallet2Txs := getTransactions(t, handler, wallet2PublicAddress)

	opts := []cmp.Option{
		cmpopts.SortSlices(func(a, b bcclient.Transaction) bool {
			return a.Amount.Cmp(b.Amount) == 1
		}),
		cmp.Comparer(func(a, b bcclient.Transaction) bool {
			return a.FromAddress == b.FromAddress && a.ToAddress == b.ToAddress && a.Amount.Cmp(b.Amount) == 0
		}),
	}
	if !cmp.Equal(wanted, wallet1Txs, opts...) {
		return false
	}
	if !cmp.Equal(wanted, wallet2Txs, opts...) {
		return false
	}

	return true
}

func TestTransactions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	httpHandler, teardown := setupServer(logger)
	defer teardown()

	lastIndexedBlock := getCurrentBlock(t, httpHandler)
	logger.Sugar().Infof("last known indexed blocked: %d", lastIndexedBlock)

	// subscribe wallets
	logger.Info("subscribing wallet 1....")
	subscribeAddress(t, httpHandler, wallet1PublicAddress)
	logger.Info("subscribing wallet 2....")
	subscribeAddress(t, httpHandler, wallet2PublicAddress)

	// generate random transactions
	txs := generateRandomTxs()
	expectedTxs := make([]bcclient.Transaction, 0, len(txs))
	for i, tx := range txs {
		fromPublic := wallet1PublicAddress
		fromPrivate := wallet1PrivateKey
		toPublic := wallet2PublicAddress

		if !tx.fromWallet1 {
			fromPublic = wallet2PublicAddress
			fromPrivate = wallet2PrivateKey
			toPublic = wallet1PublicAddress
		}

		logger.Sugar().Infof("Sending tx #%d from %s to %s with value of %d", i, fromPublic, toPublic, tx.amount)
		sendTransaction(t, fromPrivate, fromPublic, toPublic, tx.amount)

		expectedTxs = append(expectedTxs, bcclient.Transaction{
			FromAddress: fromPublic,
			ToAddress:   toPublic,
			Amount:      tx.amount,
		})
	}

	// we may need to check transactions several times because sometimes it takes longer to generate the blocks, and we can't know exactly how much time we have to wait.
	b := backoff.NewExponentialBackOff()
	b.Reset()
	err := backoff.Retry(func() error {
		// wait for new blocks with the generated random transactions
		lastIndexedBlock = waitForNewBlocks(t, logger, httpHandler, lastIndexedBlock)

		// make sure generated random transactions exists in the api response
		if !compareTxs(t, httpHandler, expectedTxs) {
			return errors.New("founded mismatched transactions")
		}

		return nil
	}, b)
	assert.NoError(t, err)
}
