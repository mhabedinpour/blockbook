package bcparser

import "blockbook/pkg/bcclient"

// Parser can be used to retrieve latest transactions of subscribed addresses on a blockchain.
type Parser interface {
	// CurrentBlockNumber returns the latest indexed block number.
	CurrentBlockNumber() uint64
	// Subscribe can be used to add an address to the watchlist. Returns false if address is already subscribed, Otherwise returns true.
	Subscribe(address string) bool
	// Unsubscribe can be used to remove an address from the watchlist. Returns false if address is not subscribed, Otherwise returns true.
	Unsubscribe(address string) bool
	// Transactions returns the latest transaction of an address. If address is not subscribed, nil is returned.
	Transactions(address string) []*bcclient.Transaction
	// Ready is used to be aware of when the parser has done its initial scan, and it's ready for usage.
	Ready() <-chan struct{}
}
