package mock

import (
	asura "github.com/teragrid/dgrid/asura/types"
	cmn "github.com/teragrid/dgrid/pkg/common"
	"github.com/teragrid/dgrid/proxy"
	"github.com/teragrid/dgrid/rpc/client"
	ctypes "github.com/teragrid/dgrid/rpc/core/types"
	"github.com/teragrid/dgrid/core/types"
)

// AsuraApp will send all asura related request to the named app,
// so you can test app behavior from a client without needing
// an entire teragrid node
type AsuraApp struct {
	App asura.Application
}

var (
	_ client.AsuraClient = AsuraApp{}
	_ client.AsuraClient = AsuraMock{}
	_ client.AsuraClient = (*AsuraRecorder)(nil)
)

func (a AsuraApp) AsuraInfo() (*ctypes.ResultAsuraInfo, error) {
	return &ctypes.ResultAsuraInfo{Response: a.App.Info(proxy.RequestInfo)}, nil
}

func (a AsuraApp) AsuraQuery(path string, data cmn.HexBytes) (*ctypes.ResultAsuraQuery, error) {
	return a.AsuraQueryWithOptions(path, data, client.DefaultAsuraQueryOptions)
}

func (a AsuraApp) AsuraQueryWithOptions(path string, data cmn.HexBytes, opts client.AsuraQueryOptions) (*ctypes.ResultAsuraQuery, error) {
	q := a.App.Query(asura.RequestQuery{
		Data:   data,
		Path:   path,
		Height: opts.Height,
		Prove:  opts.Prove,
	})
	return &ctypes.ResultAsuraQuery{Response: q}, nil
}

// NOTE: Caller should call a.App.Commit() separately,
// this function does not actually wait for a commit.
// TODO: Make it wait for a commit and set res.Height appropriately.
func (a AsuraApp) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	res := ctypes.ResultBroadcastTxCommit{}
	res.CheckTx = a.App.CheckTx(tx)
	if res.CheckTx.IsErr() {
		return &res, nil
	}
	res.DeliverTx = a.App.DeliverTx(tx)
	res.Height = -1 // TODO
	return &res, nil
}

func (a AsuraApp) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	c := a.App.CheckTx(tx)
	// and this gets written in a background thread...
	if !c.IsErr() {
		go func() { a.App.DeliverTx(tx) }() // nolint: errcheck
	}
	return &ctypes.ResultBroadcastTx{Code: c.Code, Data: c.Data, Log: c.Log, Hash: tx.Hash()}, nil
}

func (a AsuraApp) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	c := a.App.CheckTx(tx)
	// and this gets written in a background thread...
	if !c.IsErr() {
		go func() { a.App.DeliverTx(tx) }() // nolint: errcheck
	}
	return &ctypes.ResultBroadcastTx{Code: c.Code, Data: c.Data, Log: c.Log, Hash: tx.Hash()}, nil
}

// AsuraMock will send all asura related request to the named app,
// so you can test app behavior from a client without needing
// an entire teragrid node
type AsuraMock struct {
	Info            Call
	Query           Call
	BroadcastCommit Call
	Broadcast       Call
}

func (m AsuraMock) AsuraInfo() (*ctypes.ResultAsuraInfo, error) {
	res, err := m.Info.GetResponse(nil)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultAsuraInfo{Response: res.(asura.ResponseInfo)}, nil
}

func (m AsuraMock) AsuraQuery(path string, data cmn.HexBytes) (*ctypes.ResultAsuraQuery, error) {
	return m.AsuraQueryWithOptions(path, data, client.DefaultAsuraQueryOptions)
}

func (m AsuraMock) AsuraQueryWithOptions(path string, data cmn.HexBytes, opts client.AsuraQueryOptions) (*ctypes.ResultAsuraQuery, error) {
	res, err := m.Query.GetResponse(QueryArgs{path, data, opts.Height, opts.Prove})
	if err != nil {
		return nil, err
	}
	resQuery := res.(asura.ResponseQuery)
	return &ctypes.ResultAsuraQuery{Response: resQuery}, nil
}

func (m AsuraMock) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	res, err := m.BroadcastCommit.GetResponse(tx)
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultBroadcastTxCommit), nil
}

func (m AsuraMock) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := m.Broadcast.GetResponse(tx)
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultBroadcastTx), nil
}

func (m AsuraMock) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := m.Broadcast.GetResponse(tx)
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultBroadcastTx), nil
}

// AsuraRecorder can wrap another type (AsuraApp, AsuraMock, or Client)
// and record all Asura related calls.
type AsuraRecorder struct {
	Client client.AsuraClient
	Calls  []Call
}

func NewAsuraRecorder(client client.AsuraClient) *AsuraRecorder {
	return &AsuraRecorder{
		Client: client,
		Calls:  []Call{},
	}
}

type QueryArgs struct {
	Path   string
	Data   cmn.HexBytes
	Height int64
	Prove  bool
}

func (r *AsuraRecorder) addCall(call Call) {
	r.Calls = append(r.Calls, call)
}

func (r *AsuraRecorder) AsuraInfo() (*ctypes.ResultAsuraInfo, error) {
	res, err := r.Client.AsuraInfo()
	r.addCall(Call{
		Name:     "abci_info",
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *AsuraRecorder) AsuraQuery(path string, data cmn.HexBytes) (*ctypes.ResultAsuraQuery, error) {
	return r.AsuraQueryWithOptions(path, data, client.DefaultAsuraQueryOptions)
}

func (r *AsuraRecorder) AsuraQueryWithOptions(path string, data cmn.HexBytes, opts client.AsuraQueryOptions) (*ctypes.ResultAsuraQuery, error) {
	res, err := r.Client.AsuraQueryWithOptions(path, data, opts)
	r.addCall(Call{
		Name:     "abci_query",
		Args:     QueryArgs{path, data, opts.Height, opts.Prove},
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *AsuraRecorder) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	res, err := r.Client.BroadcastTxCommit(tx)
	r.addCall(Call{
		Name:     "broadcast_tx_commit",
		Args:     tx,
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *AsuraRecorder) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := r.Client.BroadcastTxAsync(tx)
	r.addCall(Call{
		Name:     "broadcast_tx_async",
		Args:     tx,
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *AsuraRecorder) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := r.Client.BroadcastTxSync(tx)
	r.addCall(Call{
		Name:     "broadcast_tx_sync",
		Args:     tx,
		Response: res,
		Error:    err,
	})
	return res, err
}
