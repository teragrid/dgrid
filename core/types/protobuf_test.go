package types

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	asura "github.com/teragrid/dgrid/asura/types"
	"github.com/teragrid/dgrid/pkg/crypto"
	"github.com/teragrid/dgrid/pkg/crypto/ed25519"
	"github.com/teragrid/dgrid/pkg/crypto/secp256k1"
	amino "github.com/teragrid/dgrid/third_party/amino"
	"github.com/teragrid/dgrid/version"
)

func TestAsuraPubKey(t *testing.T) {
	pkEd := ed25519.GenPrivKey().PubKey()
	pkSecp := secp256k1.GenPrivKey().PubKey()
	testAsuraPubKey(t, pkEd, AsuraPubKeyTypeEd25519)
	testAsuraPubKey(t, pkSecp, AsuraPubKeyTypeSecp256k1)
}

func testAsuraPubKey(t *testing.T, pk crypto.PubKey, typeStr string) {
	asuraPubKey := TM2PB.PubKey(pk)
	pk2, err := PB2TM.PubKey(asuraPubKey)
	assert.Nil(t, err)
	assert.Equal(t, pk, pk2)
}

func TestAsuraValidators(t *testing.T) {
	pkEd := ed25519.GenPrivKey().PubKey()

	// correct validator
	tmValExpected := NewValidator(pkEd, 10)

	tmVal := NewValidator(pkEd, 10)

	asuraVal := TM2PB.ValidatorUpdate(tmVal)
	tmVals, err := PB2TM.ValidatorUpdates([]asura.ValidatorUpdate{asuraVal})
	assert.Nil(t, err)
	assert.Equal(t, tmValExpected, tmVals[0])

	asuraVals := TM2PB.ValidatorUpdates(NewValidatorSet(tmVals))
	assert.Equal(t, []asura.ValidatorUpdate{asuraVal}, asuraVals)

	// val with address
	tmVal.Address = pkEd.Address()

	asuraVal = TM2PB.ValidatorUpdate(tmVal)
	tmVals, err = PB2TM.ValidatorUpdates([]asura.ValidatorUpdate{asuraVal})
	assert.Nil(t, err)
	assert.Equal(t, tmValExpected, tmVals[0])

	// val with incorrect pubkey data
	asuraVal = TM2PB.ValidatorUpdate(tmVal)
	asuraVal.PubKey.Data = []byte("incorrect!")
	tmVals, err = PB2TM.ValidatorUpdates([]asura.ValidatorUpdate{asuraVal})
	assert.NotNil(t, err)
	assert.Nil(t, tmVals)
}

func TestAsuraConsensusParams(t *testing.T) {
	cp := DefaultConsensusParams()
	asuraCP := TM2PB.ConsensusParams(cp)
	cp2 := cp.Update(asuraCP)

	assert.Equal(t, *cp, cp2)
}

func newHeader(
	height, numTxs int64,
	commitHash, dataHash, evidenceHash []byte,
) *Header {
	return &Header{
		Height:         height,
		NumTxs:         numTxs,
		LastCommitHash: commitHash,
		DataHash:       dataHash,
		EvidenceHash:   evidenceHash,
	}
}

func TestAsuraHeader(t *testing.T) {
	// build a full header
	var height int64 = 5
	var numTxs int64 = 3
	header := newHeader(
		height, numTxs,
		[]byte("lastCommitHash"), []byte("dataHash"), []byte("evidenceHash"),
	)
	protocolVersion := version.Consensus{7, 8}
	timestamp := time.Now()
	lastBlockID := BlockID{
		Hash: []byte("hash"),
		PartsHeader: PartSetHeader{
			Total: 10,
			Hash:  []byte("hash"),
		},
	}
	var totalTxs int64 = 100
	header.Populate(
		protocolVersion, "leagueID",
		timestamp, lastBlockID, totalTxs,
		[]byte("valHash"), []byte("nextValHash"),
		[]byte("consHash"), []byte("appHash"), []byte("lastResultsHash"),
		[]byte("proposerAddress"),
	)

	cdc := amino.NewCodec()
	headerBz := cdc.MustMarshalBinaryBare(header)

	pbHeader := TM2PB.Header(header)
	pbHeaderBz, err := proto.Marshal(&pbHeader)
	assert.NoError(t, err)

	// assert some fields match
	assert.EqualValues(t, protocolVersion.Block, pbHeader.Version.Block)
	assert.EqualValues(t, protocolVersion.App, pbHeader.Version.App)
	assert.EqualValues(t, "leagueID", pbHeader.LeagueID)
	assert.EqualValues(t, height, pbHeader.Height)
	assert.EqualValues(t, timestamp, pbHeader.Time)
	assert.EqualValues(t, numTxs, pbHeader.NumTxs)
	assert.EqualValues(t, totalTxs, pbHeader.TotalTxs)
	assert.EqualValues(t, lastBlockID.Hash, pbHeader.LastBlockId.Hash)
	assert.EqualValues(t, []byte("lastCommitHash"), pbHeader.LastCommitHash)
	assert.Equal(t, []byte("proposerAddress"), pbHeader.ProposerAddress)

	// assert the encodings match
	// NOTE: they don't yet because Amino encodes
	// int64 as zig-zag and we're using non-zigzag in the protobuf.
	_, _ = headerBz, pbHeaderBz
	// assert.EqualValues(t, headerBz, pbHeaderBz)

}

func TestAsuraEvidence(t *testing.T) {
	val := NewMockPV()
	blockID := makeBlockID([]byte("blockhash"), 1000, []byte("partshash"))
	blockID2 := makeBlockID([]byte("blockhash2"), 1000, []byte("partshash"))
	const leagueID = "mychain"
	pubKey := val.GetPubKey()
	ev := &DuplicateVoteEvidence{
		PubKey: pubKey,
		VoteA:  makeVote(val, leagueID, 0, 10, 2, 1, blockID),
		VoteB:  makeVote(val, leagueID, 0, 10, 2, 1, blockID2),
	}
	asuraEv := TM2PB.Evidence(
		ev,
		NewValidatorSet([]*Validator{NewValidator(pubKey, 10)}),
		time.Now(),
	)

	assert.Equal(t, "duplicate/vote", asuraEv.Type)
}

type pubKeyEddie struct{}

func (pubKeyEddie) Address() Address                        { return []byte{} }
func (pubKeyEddie) Bytes() []byte                           { return []byte{} }
func (pubKeyEddie) VerifyBytes(msg []byte, sig []byte) bool { return false }
func (pubKeyEddie) Equals(crypto.PubKey) bool               { return false }

func TestAsuraValidatorFromPubKeyAndPower(t *testing.T) {
	pubkey := ed25519.GenPrivKey().PubKey()

	asuraVal := TM2PB.NewValidatorUpdate(pubkey, 10)
	assert.Equal(t, int64(10), asuraVal.Power)

	assert.Panics(t, func() { TM2PB.NewValidatorUpdate(nil, 10) })
	assert.Panics(t, func() { TM2PB.NewValidatorUpdate(pubKeyEddie{}, 10) })
}

func TestAsuraValidatorWithoutPubKey(t *testing.T) {
	pkEd := ed25519.GenPrivKey().PubKey()

	asuraVal := TM2PB.Validator(NewValidator(pkEd, 10))

	// pubkey must be nil
	tmValExpected := asura.Validator{
		Address: pkEd.Address(),
		Power:   10,
	}

	assert.Equal(t, tmValExpected, asuraVal)
}
