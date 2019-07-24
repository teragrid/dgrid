package proxy

import (
	asura "github.com/teragrid/dgrid/asura/types"
	"github.com/teragrid/dgrid/version"
)

// RequestInfo contains all the information for sending
// the asura.RequestInfo message during handshake with the app.
// It contains only compile-time version information.
var RequestInfo = asura.RequestInfo{
	Version:      version.Version,
	BlockVersion: version.BlockProtocol.Uint64(),
	P2PVersion:   version.P2PProtocol.Uint64(),
}
