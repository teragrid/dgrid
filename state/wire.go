package state

import (
	"github.com/teragrid/go-amino"
	"github.com/teragrid/go-crypto"
)

var cdc = amino.NewCodec()

func init() {
	crypto.RegisterAmino(cdc)
}
