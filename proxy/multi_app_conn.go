package proxy

import (
	"github.com/pkg/errors"

	cmn "github.com/teragrid/dgrid/pkg/common"
)

//-----------------------------

// Dgrid's interface to the application consists of multiple connections
type AppConns interface {
	cmn.Service

	Storage() AppConnStorage
	Consensus() AppConnConsensus
	Query() AppConnQuery
}

func NewAppConns(clientCreator ClientCreator) AppConns {
	return NewMultiAppConn(clientCreator)
}

//-----------------------------
// multiAppConn implements AppConns

// a multiAppConn is made of a few appConns (storage, consensus, query)
// and manages their underlying asura clients
// TODO: on app restart, clients must reboot together
type multiAppConn struct {
	cmn.BaseService

	storageConn   *appConnStorage
	consensusConn *appConnConsensus
	queryConn     *appConnQuery

	clientCreator ClientCreator
}

// Make all necessary asura connections to the application
func NewMultiAppConn(clientCreator ClientCreator) *multiAppConn {
	multiAppConn := &multiAppConn{
		clientCreator: clientCreator,
	}
	multiAppConn.BaseService = *cmn.NewBaseService(nil, "multiAppConn", multiAppConn)
	return multiAppConn
}

// Returns the storage connection
func (app *multiAppConn) Storage() AppConnStorage {
	return app.storageConn
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
	querycli, err := app.clientCreator.NewAsuraClient()
	if err != nil {
		return errors.Wrap(err, "Error creating Asura client (query connection)")
	}
	querycli.SetLogger(app.Logger.With("module", "asura-client", "connection", "query"))
	if err := querycli.Start(); err != nil {
		return errors.Wrap(err, "Error starting Asura client (query connection)")
	}
	app.queryConn = NewAppConnQuery(querycli)

	// storage connection
	memcli, err := app.clientCreator.NewAsuraClient()
	if err != nil {
		return errors.Wrap(err, "Error creating Asura client (storage connection)")
	}
	memcli.SetLogger(app.Logger.With("module", "asura-client", "connection", "storage"))
	if err := memcli.Start(); err != nil {
		return errors.Wrap(err, "Error starting Asura client (storage connection)")
	}
	app.storageConn = NewAppConnStorage(memcli)

	// consensus connection
	concli, err := app.clientCreator.NewAsuraClient()
	if err != nil {
		return errors.Wrap(err, "Error creating Asura client (consensus connection)")
	}
	concli.SetLogger(app.Logger.With("module", "asura-client", "connection", "consensus"))
	if err := concli.Start(); err != nil {
		return errors.Wrap(err, "Error starting Asura client (consensus connection)")
	}
	app.consensusConn = NewAppConnConsensus(concli)

	return nil
}
