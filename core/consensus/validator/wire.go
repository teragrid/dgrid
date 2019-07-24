package validator

import (
	amino "github.com/teragrid/dgrid/third_party/amino"
	cryptoAmino "github.com/teragrid/dgrid/pkg/crypto/encoding/amino"
)

var cdc = amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	RegisterRemoteSignerMsg(cdc)
}
