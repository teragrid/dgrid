package proxy

import (
	"sync"

	"github.com/pkg/errors"

	asura "github.com/teragrid/dgrid/asura/client"
	"github.com/teragrid/dgrid/asura/example/counter"
	"github.com/teragrid/dgrid/asura/example/kvstore"
	"github.com/teragrid/dgrid/asura/types"
)

// NewAsuraClient returns newly connected client
type ClientCreator interface {
	NewAsuraClient() (asura.Client, error)
}

//----------------------------------------------------
// local proxy uses a mutex on an in-proc app

type localClientCreator struct {
	mtx *sync.Mutex
	app types.Application
}

func NewLocalClientCreator(app types.Application) ClientCreator {
	return &localClientCreator{
		mtx: new(sync.Mutex),
		app: app,
	}
}

func (l *localClientCreator) NewAsuraClient() (asura.Client, error) {
	return asura.NewLocalClient(l.mtx, l.app), nil
}

//---------------------------------------------------------------
// remote proxy opens new connections to an external app process

type remoteClientCreator struct {
	addr        string
	transport   string
	mustConnect bool
}

func NewRemoteClientCreator(addr, transport string, mustConnect bool) ClientCreator {
	return &remoteClientCreator{
		addr:        addr,
		transport:   transport,
		mustConnect: mustConnect,
	}
}

func (r *remoteClientCreator) NewAsuraClient() (asura.Client, error) {
	remoteApp, err := asura.NewClient(r.addr, r.transport, r.mustConnect)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to proxy")
	}
	return remoteApp, nil
}

//-----------------------------------------------------------------
// default

func DefaultClientCreator(addr, transport, dbDir string) ClientCreator {
	switch addr {
	case "counter":
		return NewLocalClientCreator(counter.NewCounterApplication(false))
	case "counter_serial":
		return NewLocalClientCreator(counter.NewCounterApplication(true))
	case "kvstore":
		return NewLocalClientCreator(kvstore.NewKVStoreApplication())
	case "persistent_kvstore":
		return NewLocalClientCreator(kvstore.NewPersistentKVStoreApplication(dbDir))
	case "noop":
		return NewLocalClientCreator(types.NewBaseApplication())
	default:
		mustConnect := false // loop retrying
		return NewRemoteClientCreator(addr, transport, mustConnect)
	}
}
