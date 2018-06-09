package mock

import (
	asura "github.com/teragrid/asura/types"
	"github.com/teragrid/teragrid/rpc/client"
	ctypes "github.com/teragrid/teragrid/rpc/core/types"
	"github.com/teragrid/teragrid/types"
	"github.com/teragrid/teragrid/version"
	cmn "github.com/teragrid/teralibs/common"
)

// asuraApp will send all asura related request to the named app,
// so you can test app behavior from a client without needing
// an entire teragrid node
type asuraApp struct {
	App asura.Application
}

var (
	_ client.asuraClient = asuraApp{}
	_ client.asuraClient = asuraMock{}
	_ client.asuraClient = (*asuraRecorder)(nil)
)

func (a asuraApp) asuraInfo() (*ctypes.ResultasuraInfo, error) {
	return &ctypes.ResultasuraInfo{a.App.Info(asura.RequestInfo{version.Version})}, nil
}

func (a asuraApp) asuraQuery(path string, data cmn.HexBytes) (*ctypes.ResultasuraQuery, error) {
	return a.asuraQueryWithOptions(path, data, client.DefaultasuraQueryOptions)
}

func (a asuraApp) asuraQueryWithOptions(path string, data cmn.HexBytes, opts client.asuraQueryOptions) (*ctypes.ResultasuraQuery, error) {
	q := a.App.Query(asura.RequestQuery{data, path, opts.Height, opts.Trusted})
	return &ctypes.ResultasuraQuery{q}, nil
}

func (a asuraApp) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	res := ctypes.ResultBroadcastTxCommit{}
	res.CheckTx = a.App.CheckTx(tx)
	if res.CheckTx.IsErr() {
		return &res, nil
	}
	res.DeliverTx = a.App.DeliverTx(tx)
	return &res, nil
}

func (a asuraApp) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	c := a.App.CheckTx(tx)
	// and this gets written in a background thread...
	if !c.IsErr() {
		go func() { a.App.DeliverTx(tx) }() // nolint: errcheck
	}
	return &ctypes.ResultBroadcastTx{c.Code, c.Data, c.Log, tx.Hash()}, nil
}

func (a asuraApp) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	c := a.App.CheckTx(tx)
	// and this gets written in a background thread...
	if !c.IsErr() {
		go func() { a.App.DeliverTx(tx) }() // nolint: errcheck
	}
	return &ctypes.ResultBroadcastTx{c.Code, c.Data, c.Log, tx.Hash()}, nil
}

// asuraMock will send all asura related request to the named app,
// so you can test app behavior from a client without needing
// an entire teragrid node
type asuraMock struct {
	Info            Call
	Query           Call
	BroadcastCommit Call
	Broadcast       Call
}

func (m asuraMock) asuraInfo() (*ctypes.ResultasuraInfo, error) {
	res, err := m.Info.GetResponse(nil)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultasuraInfo{res.(asura.ResponseInfo)}, nil
}

func (m asuraMock) asuraQuery(path string, data cmn.HexBytes) (*ctypes.ResultasuraQuery, error) {
	return m.asuraQueryWithOptions(path, data, client.DefaultasuraQueryOptions)
}

func (m asuraMock) asuraQueryWithOptions(path string, data cmn.HexBytes, opts client.asuraQueryOptions) (*ctypes.ResultasuraQuery, error) {
	res, err := m.Query.GetResponse(QueryArgs{path, data, opts.Height, opts.Trusted})
	if err != nil {
		return nil, err
	}
	resQuery := res.(asura.ResponseQuery)
	return &ctypes.ResultasuraQuery{resQuery}, nil
}

func (m asuraMock) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	res, err := m.BroadcastCommit.GetResponse(tx)
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultBroadcastTxCommit), nil
}

func (m asuraMock) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := m.Broadcast.GetResponse(tx)
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultBroadcastTx), nil
}

func (m asuraMock) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := m.Broadcast.GetResponse(tx)
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultBroadcastTx), nil
}

// asuraRecorder can wrap another type (asuraApp, asuraMock, or Client)
// and record all asura related calls.
type asuraRecorder struct {
	Client client.asuraClient
	Calls  []Call
}

func NewasuraRecorder(client client.asuraClient) *asuraRecorder {
	return &asuraRecorder{
		Client: client,
		Calls:  []Call{},
	}
}

type QueryArgs struct {
	Path    string
	Data    cmn.HexBytes
	Height  int64
	Trusted bool
}

func (r *asuraRecorder) addCall(call Call) {
	r.Calls = append(r.Calls, call)
}

func (r *asuraRecorder) asuraInfo() (*ctypes.ResultasuraInfo, error) {
	res, err := r.Client.asuraInfo()
	r.addCall(Call{
		Name:     "asura_info",
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *asuraRecorder) asuraQuery(path string, data cmn.HexBytes) (*ctypes.ResultasuraQuery, error) {
	return r.asuraQueryWithOptions(path, data, client.DefaultasuraQueryOptions)
}

func (r *asuraRecorder) asuraQueryWithOptions(path string, data cmn.HexBytes, opts client.asuraQueryOptions) (*ctypes.ResultasuraQuery, error) {
	res, err := r.Client.asuraQueryWithOptions(path, data, opts)
	r.addCall(Call{
		Name:     "asura_query",
		Args:     QueryArgs{path, data, opts.Height, opts.Trusted},
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *asuraRecorder) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	res, err := r.Client.BroadcastTxCommit(tx)
	r.addCall(Call{
		Name:     "broadcast_tx_commit",
		Args:     tx,
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *asuraRecorder) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := r.Client.BroadcastTxAsync(tx)
	r.addCall(Call{
		Name:     "broadcast_tx_async",
		Args:     tx,
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *asuraRecorder) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := r.Client.BroadcastTxSync(tx)
	r.addCall(Call{
		Name:     "broadcast_tx_sync",
		Args:     tx,
		Response: res,
		Error:    err,
	})
	return res, err
}
