package types

import (
	asura "github.com/teragrid/dgrid/asura/types"
	"github.com/teragrid/dgrid/pkg/crypto/merkle"
	cmn "github.com/teragrid/dgrid/pkg/common"
)

//-----------------------------------------------------------------------------

// AsuraResult is the deterministic component of a ResponseDeliverTx.
// TODO: add tags and other fields
// https://github.com/teragrid/dgrid/issues/1007
type AsuraResult struct {
	Code uint32       `json:"code"`
	Data cmn.HexBytes `json:"data"`
}

// Bytes returns the amino encoded AsuraResult
func (a AsuraResult) Bytes() []byte {
	return cdcEncode(a)
}

// AsuraResults wraps the deliver tx results to return a proof
type AsuraResults []AsuraResult

// NewResults creates AsuraResults from the list of ResponseDeliverTx.
func NewResults(responses []*asura.ResponseDeliverTx) AsuraResults {
	res := make(AsuraResults, len(responses))
	for i, d := range responses {
		res[i] = NewResultFromResponse(d)
	}
	return res
}

// NewResultFromResponse creates AsuraResult from ResponseDeliverTx.
func NewResultFromResponse(response *asura.ResponseDeliverTx) AsuraResult {
	return AsuraResult{
		Code: response.Code,
		Data: response.Data,
	}
}

// Bytes serializes the AsuraResponse using wire
func (a AsuraResults) Bytes() []byte {
	bz, err := cdc.MarshalBinaryLengthPrefixed(a)
	if err != nil {
		panic(err)
	}
	return bz
}

// Hash returns a merkle hash of all results
func (a AsuraResults) Hash() []byte {
	// NOTE: we copy the impl of the merkle tree for txs -
	// we should be consistent and either do it for both or not.
	return merkle.SimpleHashFromByteSlices(a.toByteSlices())
}

// ProveResult returns a merkle proof of one result from the set
func (a AsuraResults) ProveResult(i int) merkle.SimpleProof {
	_, proofs := merkle.SimpleProofsFromByteSlices(a.toByteSlices())
	return *proofs[i]
}

func (a AsuraResults) toByteSlices() [][]byte {
	l := len(a)
	bzs := make([][]byte, l)
	for i := 0; i < l; i++ {
		bzs[i] = a[i].Bytes()
	}
	return bzs
}
