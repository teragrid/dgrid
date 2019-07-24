package proxy

import (
	asura "github.com/teragrid/dgrid/asura/client"
	"github.com/teragrid/dgrid/asura/types"
)

//----------------------------------------------------------------------------------------
// Enforce which asura msgs can be sent on a connection at the type level

type AppConnConsensus interface {
	SetResponseCallback(asura.Callback)
	Error() error

	InitChainSync(types.RequestInitChain) (*types.ResponseInitChain, error)

	BeginBlockSync(types.RequestBeginBlock) (*types.ResponseBeginBlock, error)
	DeliverTxAsync(tx []byte) *asura.ReqRes
	EndBlockSync(types.RequestEndBlock) (*types.ResponseEndBlock, error)
	CommitSync() (*types.ResponseCommit, error)
}

type AppConnStorage interface {
	SetResponseCallback(asura.Callback)
	Error() error

	CheckTxAsync(tx []byte) *asura.ReqRes

	FlushAsync() *asura.ReqRes
	FlushSync() error
}

type AppConnQuery interface {
	Error() error

	EchoSync(string) (*types.ResponseEcho, error)
	InfoSync(types.RequestInfo) (*types.ResponseInfo, error)
	QuerySync(types.RequestQuery) (*types.ResponseQuery, error)

	//	SetOptionSync(key string, value string) (res types.Result)
}

//-----------------------------------------------------------------------------------------
// Implements AppConnConsensus (subset of asura.Client)

type appConnConsensus struct {
	appConn asura.Client
}

func NewAppConnConsensus(appConn asura.Client) *appConnConsensus {
	return &appConnConsensus{
		appConn: appConn,
	}
}

func (app *appConnConsensus) SetResponseCallback(cb asura.Callback) {
	app.appConn.SetResponseCallback(cb)
}

func (app *appConnConsensus) Error() error {
	return app.appConn.Error()
}

func (app *appConnConsensus) InitChainSync(req types.RequestInitChain) (*types.ResponseInitChain, error) {
	return app.appConn.InitChainSync(req)
}

func (app *appConnConsensus) BeginBlockSync(req types.RequestBeginBlock) (*types.ResponseBeginBlock, error) {
	return app.appConn.BeginBlockSync(req)
}

func (app *appConnConsensus) DeliverTxAsync(tx []byte) *asura.ReqRes {
	return app.appConn.DeliverTxAsync(tx)
}

func (app *appConnConsensus) EndBlockSync(req types.RequestEndBlock) (*types.ResponseEndBlock, error) {
	return app.appConn.EndBlockSync(req)
}

func (app *appConnConsensus) CommitSync() (*types.ResponseCommit, error) {
	return app.appConn.CommitSync()
}

//------------------------------------------------
// Implements AppConnStorage (subset of asura.Client)

type appConnStorage struct {
	appConn asura.Client
}

func NewAppConnStorage(appConn asura.Client) *appConnStorage {
	return &appConnStorage{
		appConn: appConn,
	}
}

func (app *appConnStorage) SetResponseCallback(cb asura.Callback) {
	app.appConn.SetResponseCallback(cb)
}

func (app *appConnStorage) Error() error {
	return app.appConn.Error()
}

func (app *appConnStorage) FlushAsync() *asura.ReqRes {
	return app.appConn.FlushAsync()
}

func (app *appConnStorage) FlushSync() error {
	return app.appConn.FlushSync()
}

func (app *appConnStorage) CheckTxAsync(tx []byte) *asura.ReqRes {
	return app.appConn.CheckTxAsync(tx)
}

//------------------------------------------------
// Implements AppConnQuery (subset of asura.Client)

type appConnQuery struct {
	appConn asura.Client
}

func NewAppConnQuery(appConn asura.Client) *appConnQuery {
	return &appConnQuery{
		appConn: appConn,
	}
}

func (app *appConnQuery) Error() error {
	return app.appConn.Error()
}

func (app *appConnQuery) EchoSync(msg string) (*types.ResponseEcho, error) {
	return app.appConn.EchoSync(msg)
}

func (app *appConnQuery) InfoSync(req types.RequestInfo) (*types.ResponseInfo, error) {
	return app.appConn.InfoSync(req)
}

func (app *appConnQuery) QuerySync(reqQuery types.RequestQuery) (*types.ResponseQuery, error) {
	return app.appConn.QuerySync(reqQuery)
}
