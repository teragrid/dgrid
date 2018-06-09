package consensus

import (
	"github.com/teragrid/go-amino"
	"github.com/teragrid/go-crypto"
)

var cdc = amino.NewCodec()

func init() {
	RegisterConsensusMessages(cdc)
	RegisterWALMessages(cdc)
	crypto.RegisterAmino(cdc)
}
