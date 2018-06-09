package evidence

import (
	"github.com/teragrid/go-amino"
	"github.com/teragrid/go-crypto"
	"github.com/teragrid/teragrid/types"
)

var cdc = amino.NewCodec()

func init() {
	RegisterEvidenceMessages(cdc)
	crypto.RegisterAmino(cdc)
	types.RegisterEvidences(cdc)
	RegisterMockEvidences(cdc) // For testing
}

//-------------------------------------------

func RegisterMockEvidences(cdc *amino.Codec) {
	cdc.RegisterConcrete(types.MockGoodEvidence{},
		"teragrid/MockGoodEvidence", nil)
	cdc.RegisterConcrete(types.MockBadEvidence{},
		"teragrid/MockBadEvidence", nil)
}
