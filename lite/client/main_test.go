package client_test

import (
	"os"
	"testing"

	"github.com/teragrid/asura/example/kvstore"

	nm "github.com/teragrid/teragrid/node"
	rpctest "github.com/teragrid/teragrid/rpc/test"
)

var node *nm.Node

func TestMain(m *testing.M) {
	// start a teragrid node (and merkleeyes) in the background to test against
	app := kvstore.NewKVStoreApplication()
	node = rpctest.Startteragrid(app)
	code := m.Run()

	// and shut down proper at the end
	node.Stop()
	node.Wait()
	os.Exit(code)
}
