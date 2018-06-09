package core_types

import (
	"github.com/teragrid/go-amino"
	"github.com/teragrid/go-crypto"
	"github.com/teragrid/teragrid/types"
)

func RegisterAmino(cdc *amino.Codec) {
	types.RegisterEventDatas(cdc)
	types.RegisterEvidences(cdc)
	crypto.RegisterAmino(cdc)
}
