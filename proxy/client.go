package proxy

import (
	"sync"

	"github.com/pkg/errors"

	asuracli "github.com/teragrid/asura/client"
	"github.com/teragrid/asura/example/kvstore"
	"github.com/teragrid/asura/types"
)

// NewAsuraClient returns newly connected client
type ClientCreator interface {
	NewAsuraClient() (asuracli.Client, error)
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

func (l *localClientCreator) NewAsuraClient() (asuracli.Client, error) {
	return asuracli.NewLocalClient(l.mtx, l.app), nil
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

func (r *remoteClientCreator) NewAsuraClient() (asuracli.Client, error) {
	remoteApp, err := asuracli.NewClient(r.addr, r.transport, r.mustConnect)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to proxy")
	}
	return remoteApp, nil
}

//-----------------------------------------------------------------
// default

func DefaultClientCreator(addr, transport, dbDir string) ClientCreator {
	switch addr {
	case "kvstore":
		fallthrough
	case "dummy":
		return NewLocalClientCreator(kvstore.NewKVStoreApplication())
	case "persistent_kvstore":
		fallthrough
	case "persistent_dummy":
		return NewLocalClientCreator(kvstore.NewPersistentKVStoreApplication(dbDir))
	case "nilapp":
		return NewLocalClientCreator(types.NewBaseApplication())
	default:
		mustConnect := false // loop retrying
		return NewRemoteClientCreator(addr, transport, mustConnect)
	}
}
