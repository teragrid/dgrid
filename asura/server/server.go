/*
Package server is used to start a new Asura server.

It contains two server implementation:
 * gRPC server
 * socket server

*/

package server

import (
	"fmt"

	"github.com/teragrid/dgrid/asura/types"
	cmn "github.com/teragrid/dgrid/pkg/common"
)

func NewServer(protoAddr, transport string, app types.Application) (cmn.Service, error) {
	var s cmn.Service
	var err error
	switch transport {
	case "socket":
		s = NewSocketServer(protoAddr, app)
	case "grpc":
		s = NewGRPCServer(protoAddr, types.NewGRPCApplication(app))
	default:
		err = fmt.Errorf("Unknown server type %s", transport)
	}
	return s, err
}
