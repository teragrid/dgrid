/*
package client provides a general purpose interface (Client) for connecting
to a teragrid node, as well as higher-level functionality.

The main implementation for production code is client.HTTP, which
connects via http to the jsonrpc interface of the teragrid node.

For connecting to a node running in the same process (eg. when
compiling the asura app in the same process), you can use the client.Local
implementation.

For mocking out server responses during testing to see behavior for
arbitrary return values, use the mock package.

In addition to the Client interface, which should be used externally
for maximum flexibility and testability, and two implementations,
this package also provides helper functions that work on any Client
implementation.
*/
package client

import (
	ctypes "github.com/teragrid/teragrid/rpc/core/types"
	"github.com/teragrid/teragrid/types"
	cmn "github.com/teragrid/teralibs/common"
)

// AsuraClient groups together the functionality that principally
// affects the asura app. In many cases this will be all we want,
// so we can accept an interface which is easier to mock
type AsuraClient interface {
	// reading from asura app
	AsuraInfo() (*ctypes.ResultAsuraInfo, error)
	AsuraQuery(path string, data cmn.HexBytes) (*ctypes.ResultAsuraQuery, error)
	AsuraQueryWithOptions(path string, data cmn.HexBytes,
		opts AsuraQueryOptions) (*ctypes.ResultAsuraQuery, error)

	// writing to asura app
	BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error)
	BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error)
	BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error)
}

// SignClient groups together the interfaces need to get valid
// signatures and prove anything about the chain
type SignClient interface {
	Block(height *int64) (*ctypes.ResultBlock, error)
	BlockResults(height *int64) (*ctypes.ResultBlockResults, error)
	Commit(height *int64) (*ctypes.ResultCommit, error)
	Validators(height *int64) (*ctypes.ResultValidators, error)
	Tx(hash []byte, prove bool) (*ctypes.ResultTx, error)
	TxSearch(query string, prove bool) ([]*ctypes.ResultTx, error)
}

// HistoryClient shows us data from genesis to now in large chunks.
type HistoryClient interface {
	Genesis() (*ctypes.ResultGenesis, error)
	BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error)
}

type StatusClient interface {
	// general chain info
	Status() (*ctypes.ResultStatus, error)
}

// Client wraps most important rpc calls a client would make
// if you want to listen for events, test if it also
// implements events.EventSwitch
type Client interface {
	cmn.Service
	AsuraClient
	SignClient
	HistoryClient
	StatusClient
	EventsClient
}

// NetworkClient is general info about the network state.  May not
// be needed usually.
//
// Not included in the Client interface, but generally implemented
// by concrete implementations.
type NetworkClient interface {
	NetInfo() (*ctypes.ResultNetInfo, error)
	DumpConsensusState() (*ctypes.ResultDumpConsensusState, error)
	Health() (*ctypes.ResultHealth, error)
}

// EventsClient is reactive, you can subscribe to any message, given the proper
// string. see teragrid/types/events.go
type EventsClient interface {
	types.EventBusSubscriber
}
