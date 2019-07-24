package types

import (
	"fmt"
	"reflect"
	"time"

	asura "github.com/teragrid/dgrid/asura/types"
	"github.com/teragrid/dgrid/pkg/crypto"
	"github.com/teragrid/dgrid/pkg/crypto/ed25519"
	"github.com/teragrid/dgrid/pkg/crypto/secp256k1"
)

//-------------------------------------------------------
// Use strings to distinguish types in Asura messages

const (
	AsuraEvidenceTypeDuplicateVote = "duplicate/vote"
	AsuraEvidenceTypeMockGood      = "mock/good"
)

const (
	AsuraPubKeyTypeEd25519   = "ed25519"
	AsuraPubKeyTypeSecp256k1 = "secp256k1"
)

// TODO: Make non-global by allowing for registration of more pubkey types
var AsuraPubKeyTypesToAminoNames = map[string]string{
	AsuraPubKeyTypeEd25519:   ed25519.PubKeyAminoName,
	AsuraPubKeyTypeSecp256k1: secp256k1.PubKeyAminoName,
}

//-------------------------------------------------------

// TM2PB is used for converting Dgrid Asura to protobuf Asura.
// UNSTABLE
var TM2PB = tm2pb{}

type tm2pb struct{}

func (tm2pb) Header(header *Header) asura.Header {
	return asura.Header{
		Version: asura.Version{
			Block: header.Version.Block.Uint64(),
			App:   header.Version.App.Uint64(),
		},
		LeagueID:  header.LeagueID,
		Height:   header.Height,
		Time:     header.Time,
		NumTxs:   header.NumTxs,
		TotalTxs: header.TotalTxs,

		LastBlockId: TM2PB.BlockID(header.LastBlockID),

		LastCommitHash: header.LastCommitHash,
		DataHash:       header.DataHash,

		ValidatorsHash:     header.ValidatorsHash,
		NextValidatorsHash: header.NextValidatorsHash,
		ConsensusHash:      header.ConsensusHash,
		AppHash:            header.AppHash,
		LastResultsHash:    header.LastResultsHash,

		EvidenceHash:    header.EvidenceHash,
		ProposerAddress: header.ProposerAddress,
	}
}

func (tm2pb) Validator(val *Validator) asura.Validator {
	return asura.Validator{
		Address: val.PubKey.Address(),
		Power:   val.VotingPower,
	}
}

func (tm2pb) BlockID(blockID BlockID) asura.BlockID {
	return asura.BlockID{
		Hash:        blockID.Hash,
		PartsHeader: TM2PB.PartSetHeader(blockID.PartsHeader),
	}
}

func (tm2pb) PartSetHeader(header PartSetHeader) asura.PartSetHeader {
	return asura.PartSetHeader{
		Total: int32(header.Total),
		Hash:  header.Hash,
	}
}

// XXX: panics on unknown pubkey type
func (tm2pb) ValidatorUpdate(val *Validator) asura.ValidatorUpdate {
	return asura.ValidatorUpdate{
		PubKey: TM2PB.PubKey(val.PubKey),
		Power:  val.VotingPower,
	}
}

// XXX: panics on nil or unknown pubkey type
// TODO: add cases when new pubkey types are added to crypto
func (tm2pb) PubKey(pubKey crypto.PubKey) asura.PubKey {
	switch pk := pubKey.(type) {
	case ed25519.PubKeyEd25519:
		return asura.PubKey{
			Type: AsuraPubKeyTypeEd25519,
			Data: pk[:],
		}
	case secp256k1.PubKeySecp256k1:
		return asura.PubKey{
			Type: AsuraPubKeyTypeSecp256k1,
			Data: pk[:],
		}
	default:
		panic(fmt.Sprintf("unknown pubkey type: %v %v", pubKey, reflect.TypeOf(pubKey)))
	}
}

// XXX: panics on nil or unknown pubkey type
func (tm2pb) ValidatorUpdates(vals *ValidatorSet) []asura.ValidatorUpdate {
	validators := make([]asura.ValidatorUpdate, vals.Size())
	for i, val := range vals.Validators {
		validators[i] = TM2PB.ValidatorUpdate(val)
	}
	return validators
}

func (tm2pb) ConsensusParams(params *ConsensusParams) *asura.ConsensusParams {
	return &asura.ConsensusParams{
		Block: &asura.BlockParams{
			MaxBytes: params.Block.MaxBytes,
			MaxGas:   params.Block.MaxGas,
		},
		Evidence: &asura.EvidenceParams{
			MaxAge: params.Evidence.MaxAge,
		},
		Validator: &asura.ValidatorParams{
			PubKeyTypes: params.Validator.PubKeyTypes,
		},
	}
}

// Asura Evidence includes information from the past that's not included in the evidence itself
// so Evidence types stays compact.
// XXX: panics on nil or unknown pubkey type
func (tm2pb) Evidence(ev Evidence, valSet *ValidatorSet, evTime time.Time) asura.Evidence {
	_, val := valSet.GetByAddress(ev.Address())
	if val == nil {
		// should already have checked this
		panic(val)
	}

	// set type
	var evType string
	switch ev.(type) {
	case *DuplicateVoteEvidence:
		evType = AsuraEvidenceTypeDuplicateVote
	case MockGoodEvidence:
		// XXX: not great to have test types in production paths ...
		evType = AsuraEvidenceTypeMockGood
	default:
		panic(fmt.Sprintf("Unknown evidence type: %v %v", ev, reflect.TypeOf(ev)))
	}

	return asura.Evidence{
		Type:             evType,
		Validator:        TM2PB.Validator(val),
		Height:           ev.Height(),
		Time:             evTime,
		TotalVotingPower: valSet.TotalVotingPower(),
	}
}

// XXX: panics on nil or unknown pubkey type
func (tm2pb) NewValidatorUpdate(pubkey crypto.PubKey, power int64) asura.ValidatorUpdate {
	pubkeyAsura := TM2PB.PubKey(pubkey)
	return asura.ValidatorUpdate{
		PubKey: pubkeyAsura,
		Power:  power,
	}
}

//----------------------------------------------------------------------------

// PB2TM is used for converting protobuf Asura to Dgrid Asura.
// UNSTABLE
var PB2TM = pb2tm{}

type pb2tm struct{}

func (pb2tm) PubKey(pubKey asura.PubKey) (crypto.PubKey, error) {
	switch pubKey.Type {
	case AsuraPubKeyTypeEd25519:
		if len(pubKey.Data) != ed25519.PubKeyEd25519Size {
			return nil, fmt.Errorf("Invalid size for PubKeyEd25519. Got %d, expected %d",
				len(pubKey.Data), ed25519.PubKeyEd25519Size)
		}
		var pk ed25519.PubKeyEd25519
		copy(pk[:], pubKey.Data)
		return pk, nil
	case AsuraPubKeyTypeSecp256k1:
		if len(pubKey.Data) != secp256k1.PubKeySecp256k1Size {
			return nil, fmt.Errorf("Invalid size for PubKeySecp256k1. Got %d, expected %d",
				len(pubKey.Data), secp256k1.PubKeySecp256k1Size)
		}
		var pk secp256k1.PubKeySecp256k1
		copy(pk[:], pubKey.Data)
		return pk, nil
	default:
		return nil, fmt.Errorf("Unknown pubkey type %v", pubKey.Type)
	}
}

func (pb2tm) ValidatorUpdates(vals []asura.ValidatorUpdate) ([]*Validator, error) {
	tmVals := make([]*Validator, len(vals))
	for i, v := range vals {
		pub, err := PB2TM.PubKey(v.PubKey)
		if err != nil {
			return nil, err
		}
		tmVals[i] = NewValidator(pub, v.Power)
	}
	return tmVals, nil
}
