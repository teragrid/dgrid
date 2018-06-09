package types

import (
	asura "github.com/teragrid/asura/types"
	cmn "github.com/teragrid/teralibs/common"
	"github.com/teragrid/teralibs/merkle"
)

//-----------------------------------------------------------------------------

// asuraResult is the deterministic component of a ResponseDeliverTx.
// TODO: add Tags
type asuraResult struct {
	Code uint32       `json:"code"`
	Data cmn.HexBytes `json:"data"`
}

// Hash returns the canonical hash of the asuraResult
func (a asuraResult) Hash() []byte {
	bz := aminoHash(a)
	return bz
}

// asuraResults wraps the deliver tx results to return a proof
type asuraResults []asuraResult

// NewResults creates asuraResults from ResponseDeliverTx
func NewResults(del []*asura.ResponseDeliverTx) asuraResults {
	res := make(asuraResults, len(del))
	for i, d := range del {
		res[i] = NewResultFromResponse(d)
	}
	return res
}

func NewResultFromResponse(response *asura.ResponseDeliverTx) asuraResult {
	return asuraResult{
		Code: response.Code,
		Data: response.Data,
	}
}

// Bytes serializes the asuraResponse using wire
func (a asuraResults) Bytes() []byte {
	bz, err := cdc.MarshalBinary(a)
	if err != nil {
		panic(err)
	}
	return bz
}

// Hash returns a merkle hash of all results
func (a asuraResults) Hash() []byte {
	return merkle.SimpleHashFromHashers(a.toHashers())
}

// ProveResult returns a merkle proof of one result from the set
func (a asuraResults) ProveResult(i int) merkle.SimpleProof {
	_, proofs := merkle.SimpleProofsFromHashers(a.toHashers())
	return *proofs[i]
}

func (a asuraResults) toHashers() []merkle.Hasher {
	l := len(a)
	hashers := make([]merkle.Hasher, l)
	for i := 0; i < l; i++ {
		hashers[i] = a[i]
	}
	return hashers
}
