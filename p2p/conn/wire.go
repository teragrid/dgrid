package conn

import (
	"github.com/teragrid/go-amino"
	"github.com/teragrid/go-crypto"
)

var cdc *amino.Codec = amino.NewCodec()

func init() {
	crypto.RegisterAmino(cdc)
	RegisterPacket(cdc)
}
