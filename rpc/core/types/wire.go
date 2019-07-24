package core_types

import (
	amino "github.com/teragrid/dgrid/third_party/amino"
	"github.com/teragrid/dgrid/core/types"
)

func RegisterAmino(cdc *amino.Codec) {
	types.RegisterEventDatas(cdc)
	types.RegisterBlockAmino(cdc)
}
