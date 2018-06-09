package proxy

import (
	"github.com/pkg/errors"

	cmn "github.com/teragrid/teralibs/common"
)

//-----------------------------

// teragrid's interface to the application consists of multiple connections
type AppConns interface {
	cmn.Service

	Mempool() AppConnMempool
	Consensus() AppConnConsensus
	Query() AppConnQuery
}

func NewAppConns(clientCreator ClientCreator, handshaker Handshaker) AppConns {
	return NewMultiAppConn(clientCreator, handshaker)
}

//-----------------------------
// multiAppConn implements AppConns

type Handshaker interface {
	Handshake(AppConns) error
}

// a multiAppConn is made of a few appConns (mempool, consensus, query)
// and manages their underlying asura clients, including the handshake
// which ensures the app and teragrid are synced.
// TODO: on app restart, clients must reboot together
type multiAppConn struct {
	cmn.BaseService

	handshaker Handshaker

	mempoolConn   *appConnMempool
	consensusConn *appConnConsensus
	queryConn     *appConnQuery

	clientCreator ClientCreator
}

// Make all necessary asura connections to the application
func NewMultiAppConn(clientCreator ClientCreator, handshaker Handshaker) *multiAppConn {
	multiAppConn := &multiAppConn{
		handshaker:    handshaker,
		clientCreator: clientCreator,
	}
	multiAppConn.BaseService = *cmn.NewBaseService(nil, "multiAppConn", multiAppConn)
	return multiAppConn
}

// Returns the mempool connection
func (app *multiAppConn) Mempool() AppConnMempool {
	return app.mempoolConn
}

// Returns the consensus Connection
func (app *multiAppConn) Consensus() AppConnConsensus {
	return app.consensusConn
}

// Returns the query Connection
func (app *multiAppConn) Query() AppConnQuery {
	return app.queryConn
}

func (app *multiAppConn) OnStart() error {
	// query connection
	querycli, err := app.clientCreator.NewasuraClient()
	if err != nil {
		return errors.Wrap(err, "Error creating asura client (query connection)")
	}
	querycli.SetLogger(app.Logger.With("module", "asura-client", "connection", "query"))
	if err := querycli.Start(); err != nil {
		return errors.Wrap(err, "Error starting asura client (query connection)")
	}
	app.queryConn = NewAppConnQuery(querycli)

	// mempool connection
	memcli, err := app.clientCreator.NewasuraClient()
	if err != nil {
		return errors.Wrap(err, "Error creating asura client (mempool connection)")
	}
	memcli.SetLogger(app.Logger.With("module", "asura-client", "connection", "mempool"))
	if err := memcli.Start(); err != nil {
		return errors.Wrap(err, "Error starting asura client (mempool connection)")
	}
	app.mempoolConn = NewAppConnMempool(memcli)

	// consensus connection
	concli, err := app.clientCreator.NewasuraClient()
	if err != nil {
		return errors.Wrap(err, "Error creating asura client (consensus connection)")
	}
	concli.SetLogger(app.Logger.With("module", "asura-client", "connection", "consensus"))
	if err := concli.Start(); err != nil {
		return errors.Wrap(err, "Error starting asura client (consensus connection)")
	}
	app.consensusConn = NewAppConnConsensus(concli)

	// ensure app is synced to the latest state
	if app.handshaker != nil {
		return app.handshaker.Handshake(app)
	}

	return nil
}
