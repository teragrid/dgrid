package client_test

import (
	"os"
	"testing"

	"github.com/teragrid/dgrid/asura/example/kvstore"
	nm "github.com/teragrid/dgrid/node"
	rpctest "github.com/teragrid/dgrid/rpc/test"
)

var node *nm.Node

func TestMain(m *testing.M) {
	// start a teragrid node (and kvstore) in the background to test against
	app := kvstore.NewKVStoreApplication()
	node = rpctest.StartTeragrid(app)

	code := m.Run()

	// and shut down proper at the end
	rpctest.StopTeragrid(node)
	os.Exit(code)
}
