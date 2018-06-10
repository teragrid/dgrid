package client

import (
	"context"

	nm "github.com/teragrid/teragrid/node"
	"github.com/teragrid/teragrid/rpc/core"
	ctypes "github.com/teragrid/teragrid/rpc/core/types"
	"github.com/teragrid/teragrid/types"
	cmn "github.com/teragrid/teralibs/common"
	tmpubsub "github.com/teragrid/teralibs/pubsub"
)

/*
Local is a Client implementation that directly executes the rpc
functions on a given node, without going through HTTP or GRPC.

This implementation is useful for:

* Running tests against a node in-process without the overhead
of going through an http server
* Communication between an asura app and teragrid core when they
are compiled in process.

For real clients, you probably want to use client.HTTP.  For more
powerful control during testing, you probably want the "client/mock" package.
*/
type Local struct {
	*types.EventBus
}

// NewLocal configures a client that calls the Node directly.
//
// Note that given how rpc/core works with package singletons, that
// you can only have one node per process.  So make sure test cases
// don't run in parallel, or try to simulate an entire network in
// one process...
func NewLocal(node *nm.Node) *Local {
	node.ConfigureRPC()
	return &Local{
		EventBus: node.EventBus(),
	}
}

var (
	_ Client        = (*Local)(nil)
	_ NetworkClient = Local{}
	_ EventsClient  = (*Local)(nil)
)

func (Local) Status() (*ctypes.ResultStatus, error) {
	return core.Status()
}

func (Local) AsuraInfo() (*ctypes.ResultAsuraInfo, error) {
	return core.AsuraInfo()
}

func (c *Local) AsuraQuery(path string, data cmn.HexBytes) (*ctypes.ResultAsuraQuery, error) {
	return c.AsuraQueryWithOptions(path, data, DefaultAsuraQueryOptions)
}

func (Local) AsuraQueryWithOptions(path string, data cmn.HexBytes, opts AsuraQueryOptions) (*ctypes.ResultAsuraQuery, error) {
	return core.AsuraQuery(path, data, opts.Height, opts.Trusted)
}

func (Local) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	return core.BroadcastTxCommit(tx)
}

func (Local) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return core.BroadcastTxAsync(tx)
}

func (Local) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return core.BroadcastTxSync(tx)
}

func (Local) NetInfo() (*ctypes.ResultNetInfo, error) {
	return core.NetInfo()
}

func (Local) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	return core.DumpConsensusState()
}

func (Local) Health() (*ctypes.ResultHealth, error) {
	return core.Health()
}

func (Local) DialSeeds(seeds []string) (*ctypes.ResultDialSeeds, error) {
	return core.UnsafeDialSeeds(seeds)
}

func (Local) DialPeers(peers []string, persistent bool) (*ctypes.ResultDialPeers, error) {
	return core.UnsafeDialPeers(peers, persistent)
}

func (Local) BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	return core.BlockchainInfo(minHeight, maxHeight)
}

func (Local) Genesis() (*ctypes.ResultGenesis, error) {
	return core.Genesis()
}

func (Local) Block(height *int64) (*ctypes.ResultBlock, error) {
	return core.Block(height)
}

func (Local) BlockResults(height *int64) (*ctypes.ResultBlockResults, error) {
	return core.BlockResults(height)
}

func (Local) Commit(height *int64) (*ctypes.ResultCommit, error) {
	return core.Commit(height)
}

func (Local) Validators(height *int64) (*ctypes.ResultValidators, error) {
	return core.Validators(height)
}

func (Local) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	return core.Tx(hash, prove)
}

func (Local) TxSearch(query string, prove bool) ([]*ctypes.ResultTx, error) {
	return core.TxSearch(query, prove)
}

func (c *Local) Subscribe(ctx context.Context, subscriber string, query tmpubsub.Query, out chan<- interface{}) error {
	return c.EventBus.Subscribe(ctx, subscriber, query, out)
}

func (c *Local) Unsubscribe(ctx context.Context, subscriber string, query tmpubsub.Query) error {
	return c.EventBus.Unsubscribe(ctx, subscriber, query)
}

func (c *Local) UnsubscribeAll(ctx context.Context, subscriber string) error {
	return c.EventBus.UnsubscribeAll(ctx, subscriber)
}
