package storage

import (
	amino "github.com/teragrid/dgrid/third_party/amino"
)

var cdc = amino.NewCodec()

func init() {
	RegisterStorageMessages(cdc)
}
