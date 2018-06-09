package state

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/teragrid/asura/example/kvstore"
	asura "github.com/teragrid/asura/types"
	crypto "github.com/teragrid/go-crypto"
	"github.com/teragrid/teragrid/proxy"
	"github.com/teragrid/teragrid/types"
	cmn "github.com/teragrid/teralibs/common"
	dbm "github.com/teragrid/teralibs/db"
	"github.com/teragrid/teralibs/log"
)

var (
	privKey      = crypto.GenPrivKeyEd25519FromSecret([]byte("execution_test"))
	chainID      = "execution_chain"
	testPartSize = 65536
	nTxsPerBlock = 10
)

func TestApplyBlock(t *testing.T) {
	cc := proxy.NewLocalClientCreator(kvstore.NewKVStoreApplication())
	proxyApp := proxy.NewAppConns(cc, nil)
	err := proxyApp.Start()
	require.Nil(t, err)
	defer proxyApp.Stop()

	state, stateDB := state(), dbm.NewMemDB()

	blockExec := NewBlockExecutor(stateDB, log.TestingLogger(), proxyApp.Consensus(),
		types.MockMempool{}, types.MockEvidencePool{})

	block := makeBlock(state, 1)
	blockID := types.BlockID{block.Hash(), block.MakePartSet(testPartSize).Header()}

	state, err = blockExec.ApplyBlock(state, blockID, block)
	require.Nil(t, err)

	// TODO check state and mempool
}

// TestBeginBlockAbsentValidators ensures we send absent validators list.
func TestBeginBlockAbsentValidators(t *testing.T) {
	app := &testApp{}
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := proxy.NewAppConns(cc, nil)
	err := proxyApp.Start()
	require.Nil(t, err)
	defer proxyApp.Stop()

	state := state()

	prevHash := state.LastBlockID.Hash
	prevParts := types.PartSetHeader{}
	prevBlockID := types.BlockID{prevHash, prevParts}

	now := time.Now().UTC()
	testCases := []struct {
		desc                     string
		lastCommitPrecommits     []*types.Vote
		expectedAbsentValidators []int32
	}{
		{"none absent", []*types.Vote{{ValidatorIndex: 0, Timestamp: now, Type: types.VoteTypePrecommit}, {ValidatorIndex: 1, Timestamp: now}}, []int32{}},
		{"one absent", []*types.Vote{{ValidatorIndex: 0, Timestamp: now, Type: types.VoteTypePrecommit}, nil}, []int32{1}},
		{"multiple absent", []*types.Vote{nil, nil}, []int32{0, 1}},
	}

	for _, tc := range testCases {
		lastCommit := &types.Commit{BlockID: prevBlockID, Precommits: tc.lastCommitPrecommits}

		block, _ := state.MakeBlock(2, makeTxs(2), lastCommit)
		_, err = ExecCommitBlock(proxyApp.Consensus(), block, log.TestingLogger())
		require.Nil(t, err, tc.desc)

		// -> app must receive an index of the absent validator
		assert.Equal(t, tc.expectedAbsentValidators, app.AbsentValidators, tc.desc)
	}
}

// TestBeginBlockByzantineValidators ensures we send byzantine validators list.
func TestBeginBlockByzantineValidators(t *testing.T) {
	app := &testApp{}
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := proxy.NewAppConns(cc, nil)
	err := proxyApp.Start()
	require.Nil(t, err)
	defer proxyApp.Stop()

	state := state()

	prevHash := state.LastBlockID.Hash
	prevParts := types.PartSetHeader{}
	prevBlockID := types.BlockID{prevHash, prevParts}

	height1, idx1, val1 := int64(8), 0, []byte("val1")
	height2, idx2, val2 := int64(3), 1, []byte("val2")
	ev1 := types.NewMockGoodEvidence(height1, idx1, val1)
	ev2 := types.NewMockGoodEvidence(height2, idx2, val2)

	testCases := []struct {
		desc                        string
		evidence                    []types.Evidence
		expectedByzantineValidators []asura.Evidence
	}{
		{"none byzantine", []types.Evidence{}, []asura.Evidence{}},
		{"one byzantine", []types.Evidence{ev1}, []asura.Evidence{{ev1.Address(), ev1.Height()}}},
		{"multiple byzantine", []types.Evidence{ev1, ev2}, []asura.Evidence{
			{ev1.Address(), ev1.Height()},
			{ev2.Address(), ev2.Height()}}},
	}

	for _, tc := range testCases {
		lastCommit := &types.Commit{BlockID: prevBlockID}

		block, _ := state.MakeBlock(10, makeTxs(2), lastCommit)
		block.Evidence.Evidence = tc.evidence
		_, err = ExecCommitBlock(proxyApp.Consensus(), block, log.TestingLogger())
		require.Nil(t, err, tc.desc)

		// -> app must receive an index of the byzantine validator
		assert.Equal(t, tc.expectedByzantineValidators, app.ByzantineValidators, tc.desc)
	}
}

//----------------------------------------------------------------------------

// make some bogus txs
func makeTxs(height int64) (txs []types.Tx) {
	for i := 0; i < nTxsPerBlock; i++ {
		txs = append(txs, types.Tx([]byte{byte(height), byte(i)}))
	}
	return txs
}

func state() State {
	s, _ := MakeGenesisState(&types.GenesisDoc{
		ChainID: chainID,
		Validators: []types.GenesisValidator{
			{privKey.PubKey(), 10000, "test"},
		},
		AppHash: nil,
	})
	return s
}

func makeBlock(state State, height int64) *types.Block {
	block, _ := state.MakeBlock(height, makeTxs(state.LastBlockHeight), new(types.Commit))
	return block
}

//----------------------------------------------------------------------------

var _ asura.Application = (*testApp)(nil)

type testApp struct {
	asura.BaseApplication

	AbsentValidators    []int32
	ByzantineValidators []asura.Evidence
}

func NewKVStoreApplication() *testApp {
	return &testApp{}
}

func (app *testApp) Info(req asura.RequestInfo) (resInfo asura.ResponseInfo) {
	return asura.ResponseInfo{}
}

func (app *testApp) BeginBlock(req asura.RequestBeginBlock) asura.ResponseBeginBlock {
	app.AbsentValidators = req.AbsentValidators
	app.ByzantineValidators = req.ByzantineValidators
	return asura.ResponseBeginBlock{}
}

func (app *testApp) DeliverTx(tx []byte) asura.ResponseDeliverTx {
	return asura.ResponseDeliverTx{Tags: []cmn.KVPair{}}
}

func (app *testApp) CheckTx(tx []byte) asura.ResponseCheckTx {
	return asura.ResponseCheckTx{}
}

func (app *testApp) Commit() asura.ResponseCommit {
	return asura.ResponseCommit{}
}

func (app *testApp) Query(reqQuery asura.RequestQuery) (resQuery asura.ResponseQuery) {
	return
}
