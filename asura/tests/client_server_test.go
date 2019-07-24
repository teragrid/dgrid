package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abciclient "github.com/teragrid/dgrid/asura/client"
	"github.com/teragrid/dgrid/asura/example/kvstore"
	asuraserver "github.com/teragrid/dgrid/asura/server"
)

func TestClientServerNoAddrPrefix(t *testing.T) {
	addr := "localhost:26658"
	transport := "socket"
	app := kvstore.NewKVStoreApplication()

	server, err := asuraserver.NewServer(addr, transport, app)
	assert.NoError(t, err, "expected no error on NewServer")
	err = server.Start()
	assert.NoError(t, err, "expected no error on server.Start")

	client, err := abciclient.NewClient(addr, transport, true)
	assert.NoError(t, err, "expected no error on NewClient")
	err = client.Start()
	assert.NoError(t, err, "expected no error on client.Start")
}
