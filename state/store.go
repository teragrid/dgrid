package state

import (
	"fmt"

	asura "github.com/teragrid/asura/types"
	"github.com/teragrid/teragrid/types"
	cmn "github.com/teragrid/teralibs/common"
	dbm "github.com/teragrid/teralibs/db"
)

//------------------------------------------------------------------------

func calcValidatorsKey(height int64) []byte {
	return []byte(cmn.Fmt("validatorsKey:%v", height))
}

func calcConsensusParamsKey(height int64) []byte {
	return []byte(cmn.Fmt("consensusParamsKey:%v", height))
}

func calcAsuraResponsesKey(height int64) []byte {
	return []byte(cmn.Fmt("AsuraResponsesKey:%v", height))
}

// LoadStateFromDBOrGenesisFile loads the most recent state from the database,
// or creates a new one from the given genesisFilePath and persists the result
// to the database.
func LoadStateFromDBOrGenesisFile(stateDB dbm.DB, genesisFilePath string) (State, error) {
	state := LoadState(stateDB)
	if state.IsEmpty() {
		var err error
		state, err = MakeGenesisStateFromFile(genesisFilePath)
		if err != nil {
			return state, err
		}
		SaveState(stateDB, state)
	}

	return state, nil
}

// LoadStateFromDBOrGenesisDoc loads the most recent state from the database,
// or creates a new one from the given genesisDoc and persists the result
// to the database.
func LoadStateFromDBOrGenesisDoc(stateDB dbm.DB, genesisDoc *types.GenesisDoc) (State, error) {
	state := LoadState(stateDB)
	if state.IsEmpty() {
		var err error
		state, err = MakeGenesisState(genesisDoc)
		if err != nil {
			return state, err
		}
		SaveState(stateDB, state)
	}

	return state, nil
}

// LoadState loads the State from the database.
func LoadState(db dbm.DB) State {
	return loadState(db, stateKey)
}

func loadState(db dbm.DB, key []byte) (state State) {
	buf := db.Get(key)
	if len(buf) == 0 {
		return state
	}

	err := cdc.UnmarshalBinaryBare(buf, &state)
	if err != nil {
		// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
		cmn.Exit(cmn.Fmt(`LoadState: Data has been corrupted or its spec has changed:
                %v\n`, err))
	}
	// TODO: ensure that buf is completely read.

	return state
}

// SaveState persists the State, the ValidatorsInfo, and the ConsensusParamsInfo to the database.
func SaveState(db dbm.DB, s State) {
	saveState(db, s, stateKey)
}

func saveState(db dbm.DB, s State, key []byte) {
	nextHeight := s.LastBlockHeight + 1
	saveValidatorsInfo(db, nextHeight, s.LastHeightValidatorsChanged, s.Validators)
	saveConsensusParamsInfo(db, nextHeight, s.LastHeightConsensusParamsChanged, s.ConsensusParams)
	db.SetSync(stateKey, s.Bytes())
}

//------------------------------------------------------------------------

// AsuraResponses retains the responses
// of the various asura calls during block processing.
// It is persisted to disk for each height before calling Commit.
type AsuraResponses struct {
	DeliverTx []*asura.ResponseDeliverTx
	EndBlock  *asura.ResponseEndBlock
}

// NewAsuraResponses returns a new AsuraResponses
func NewAsuraResponses(block *types.Block) *AsuraResponses {
	resDeliverTxs := make([]*asura.ResponseDeliverTx, block.NumTxs)
	if block.NumTxs == 0 {
		// This makes Amino encoding/decoding consistent.
		resDeliverTxs = nil
	}
	return &AsuraResponses{
		DeliverTx: resDeliverTxs,
	}
}

// Bytes serializes the asuraResponse using go-amino.
func (arz *AsuraResponses) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(arz)
}

func (arz *AsuraResponses) ResultsHash() []byte {
	results := types.NewResults(arz.DeliverTx)
	return results.Hash()
}

// LoadAsuraResponses loads the AsuraResponses for the given height from the database.
// This is useful for recovering from crashes where we called app.Commit and before we called
// s.Save(). It can also be used to produce Merkle proofs of the result of txs.
func LoadAsuraResponses(db dbm.DB, height int64) (*AsuraResponses, error) {
	buf := db.Get(calcAsuraResponsesKey(height))
	if len(buf) == 0 {
		return nil, ErrNoAsuraResponsesForHeight{height}
	}

	AsuraResponses := new(AsuraResponses)
	err := cdc.UnmarshalBinaryBare(buf, AsuraResponses)
	if err != nil {
		// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
		cmn.Exit(cmn.Fmt(`LoadAsuraResponses: Data has been corrupted or its spec has
                changed: %v\n`, err))
	}
	// TODO: ensure that buf is completely read.

	return AsuraResponses, nil
}

// SaveAsuraResponses persists the AsuraResponses to the database.
// This is useful in case we crash after app.Commit and before s.Save().
// Responses are indexed by height so they can also be loaded later to produce Merkle proofs.
func saveAsuraResponses(db dbm.DB, height int64, AsuraResponses *AsuraResponses) {
	db.SetSync(calcAsuraResponsesKey(height), AsuraResponses.Bytes())
}

//-----------------------------------------------------------------------------

// ValidatorsInfo represents the latest validator set, or the last height it changed
type ValidatorsInfo struct {
	ValidatorSet      *types.ValidatorSet
	LastHeightChanged int64
}

// Bytes serializes the ValidatorsInfo using go-amino.
func (valInfo *ValidatorsInfo) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(valInfo)
}

// LoadValidators loads the ValidatorSet for a given height.
// Returns ErrNoValSetForHeight if the validator set can't be found for this height.
func LoadValidators(db dbm.DB, height int64) (*types.ValidatorSet, error) {
	valInfo := loadValidatorsInfo(db, height)
	if valInfo == nil {
		return nil, ErrNoValSetForHeight{height}
	}

	if valInfo.ValidatorSet == nil {
		valInfo = loadValidatorsInfo(db, valInfo.LastHeightChanged)
		if valInfo == nil {
			cmn.PanicSanity(fmt.Sprintf(`Couldn't find validators at height %d as
                        last changed from height %d`, valInfo.LastHeightChanged, height))
		}
	}

	return valInfo.ValidatorSet, nil
}

func loadValidatorsInfo(db dbm.DB, height int64) *ValidatorsInfo {
	buf := db.Get(calcValidatorsKey(height))
	if len(buf) == 0 {
		return nil
	}

	v := new(ValidatorsInfo)
	err := cdc.UnmarshalBinaryBare(buf, v)
	if err != nil {
		// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
		cmn.Exit(cmn.Fmt(`LoadValidators: Data has been corrupted or its spec has changed:
                %v\n`, err))
	}
	// TODO: ensure that buf is completely read.

	return v
}

// saveValidatorsInfo persists the validator set for the next block to disk.
// It should be called from s.Save(), right before the state itself is persisted.
// If the validator set did not change after processing the latest block,
// only the last height for which the validators changed is persisted.
func saveValidatorsInfo(db dbm.DB, nextHeight, changeHeight int64, valSet *types.ValidatorSet) {
	valInfo := &ValidatorsInfo{
		LastHeightChanged: changeHeight,
	}
	if changeHeight == nextHeight {
		valInfo.ValidatorSet = valSet
	}
	db.SetSync(calcValidatorsKey(nextHeight), valInfo.Bytes())
}

//-----------------------------------------------------------------------------

// ConsensusParamsInfo represents the latest consensus params, or the last height it changed
type ConsensusParamsInfo struct {
	ConsensusParams   types.ConsensusParams
	LastHeightChanged int64
}

// Bytes serializes the ConsensusParamsInfo using go-amino.
func (params ConsensusParamsInfo) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(params)
}

// LoadConsensusParams loads the ConsensusParams for a given height.
func LoadConsensusParams(db dbm.DB, height int64) (types.ConsensusParams, error) {
	empty := types.ConsensusParams{}

	paramsInfo := loadConsensusParamsInfo(db, height)
	if paramsInfo == nil {
		return empty, ErrNoConsensusParamsForHeight{height}
	}

	if paramsInfo.ConsensusParams == empty {
		paramsInfo = loadConsensusParamsInfo(db, paramsInfo.LastHeightChanged)
		if paramsInfo == nil {
			cmn.PanicSanity(fmt.Sprintf(`Couldn't find consensus params at height %d as
                        last changed from height %d`, paramsInfo.LastHeightChanged, height))
		}
	}

	return paramsInfo.ConsensusParams, nil
}

func loadConsensusParamsInfo(db dbm.DB, height int64) *ConsensusParamsInfo {
	buf := db.Get(calcConsensusParamsKey(height))
	if len(buf) == 0 {
		return nil
	}

	paramsInfo := new(ConsensusParamsInfo)
	err := cdc.UnmarshalBinaryBare(buf, paramsInfo)
	if err != nil {
		// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
		cmn.Exit(cmn.Fmt(`LoadConsensusParams: Data has been corrupted or its spec has changed:
                %v\n`, err))
	}
	// TODO: ensure that buf is completely read.

	return paramsInfo
}

// saveConsensusParamsInfo persists the consensus params for the next block to disk.
// It should be called from s.Save(), right before the state itself is persisted.
// If the consensus params did not change after processing the latest block,
// only the last height for which they changed is persisted.
func saveConsensusParamsInfo(db dbm.DB, nextHeight, changeHeight int64, params types.ConsensusParams) {
	paramsInfo := &ConsensusParamsInfo{
		LastHeightChanged: changeHeight,
	}
	if changeHeight == nextHeight {
		paramsInfo.ConsensusParams = params
	}
	db.SetSync(calcConsensusParamsKey(nextHeight), paramsInfo.Bytes())
}