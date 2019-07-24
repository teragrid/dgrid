package merkle

import (
	"github.com/teragrid/dgrid/third_party/amino"
)

var cdc *amino.Codec

func init() {
	cdc = amino.NewCodec()
	cdc.Seal()
}
