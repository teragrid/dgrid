package core_grpc_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/teragrid/dgrid/asura/example/kvstore"
	core_grpc "github.com/teragrid/dgrid/rpc/grpc"
	rpctest "github.com/teragrid/dgrid/rpc/test"
)

func TestMain(m *testing.M) {
	// start a teragrid node in the background to test against
	app := kvstore.NewKVStoreApplication()
	node := rpctest.StartTeragrid(app)

	code := m.Run()

	// and shut down proper at the end
	rpctest.StopTeragrid(node)
	os.Exit(code)
}

func TestBroadcastTx(t *testing.T) {
	res, err := rpctest.GetGRPCClient().BroadcastTx(context.Background(), &core_grpc.RequestBroadcastTx{Tx: []byte("this is a tx")})
	require.NoError(t, err)
	require.EqualValues(t, 0, res.CheckTx.Code)
	require.EqualValues(t, 0, res.DeliverTx.Code)
}
