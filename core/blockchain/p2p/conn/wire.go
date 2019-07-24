package conn

import (
	amino "github.com/teragrid/dgrid/third_party/amino"
	cryptoAmino "github.com/teragrid/dgrid/pkg/crypto/encoding/amino"
)

var cdc *amino.Codec = amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	RegisterPacket(cdc)
}
