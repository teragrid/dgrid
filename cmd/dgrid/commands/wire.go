package commands

import (
	"github.com/teragrid/dgrid/pkg/amino"
	"github.com/teragrid/go-crypto"
)

var cdc = amino.NewCodec()

func init() {
	crypto.RegisterAmino(cdc)
}
