package pex

import (
	amino "github.com/teragrid/dgrid/third_party/amino"
)

var cdc *amino.Codec = amino.NewCodec()

func init() {
	RegisterPexMessage(cdc)
}
