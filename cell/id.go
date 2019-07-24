package cell

import (
	"time"

	"github.com/teragrid/dgrid/pkg/crypto"
)

type CellID struct {
	Name   string
	PubKey crypto.PubKey
}

type PrivCellID struct {
	CellID
	PrivKey crypto.PrivKey
}

type CellGreeting struct {
	CellID
	Version  string
	LeagueID string
	Message  string
	Time     time.Time
}

type SignedCellGreeting struct {
	CellGreeting
	Signature []byte
}

func (pnid *PrivCellID) SignGreeting() *SignedCellGreeting {
	//greeting := CellGreeting{}
	return nil
}
