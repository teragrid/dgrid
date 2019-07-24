package node

import (
	"context"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/teragrid/dgrid/asura/example/kvstore"
	"github.com/teragrid/dgrid/core/blockchain/p2p"
	cfg "github.com/teragrid/dgrid/core/config"
	"github.com/teragrid/dgrid/core/types"
	ttime "github.com/teragrid/dgrid/core/types/time"
	"github.com/teragrid/dgrid/evidence"
	mempl "github.com/teragrid/dgrid/storage"
	cmn "github.com/teragrid/dgrid/pkg/common"
	"github.com/teragrid/dgrid/pkg/crypto/ed25519"
	dbm "github.com/teragrid/dgrid/pkg/db"
	"github.com/teragrid/dgrid/pkg/log"
	"github.com/teragrid/dgrid/proxy"
	sm "github.com/teragrid/dgrid/state"
	"github.com/teragrid/dgrid/version"
	"github.com/teragrid/dgridcore/consensus/validator"
)

func TestNodeStartStop(t *testing.T) {
	config := cfg.ResetTestRoot("node_node_test")
	defer os.RemoveAll(config.RootDir)

	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err)
	err = n.Start()
	require.NoError(t, err)

	t.Logf("Started node %v", n.sw.NodeInfo())

	// wait for the node to produce a block
	blocksSub, err := n.EventBus().Subscribe(context.Background(), "node_test", types.EventQueryNewBlock)
	require.NoError(t, err)
	select {
	case <-blocksSub.Out():
	case <-blocksSub.Cancelled():
		t.Fatal("blocksSub was cancelled")
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for the node to produce a block")
	}

	// stop the node
	go func() {
		n.Stop()
	}()

	select {
	case <-n.Quit():
	case <-time.After(5 * time.Second):
		pid := os.Getpid()
		p, err := os.FindProcess(pid)
		if err != nil {
			panic(err)
		}
		err = p.Signal(syscall.SIGABRT)
		fmt.Println(err)
		t.Fatal("timed out waiting for shutdown")
	}
}

func TestSplitAndTrimEmpty(t *testing.T) {
	testCases := []struct {
		s        string
		sep      string
		cutset   string
		expected []string
	}{
		{"a,b,c", ",", " ", []string{"a", "b", "c"}},
		{" a , b , c ", ",", " ", []string{"a", "b", "c"}},
		{" a, b, c ", ",", " ", []string{"a", "b", "c"}},
		{" a, ", ",", " ", []string{"a"}},
		{"   ", ",", " ", []string{}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, splitAndTrimEmpty(tc.s, tc.sep, tc.cutset), "%s", tc.s)
	}
}

func TestNodeDelayedStart(t *testing.T) {
	config := cfg.ResetTestRoot("node_delayed_start_test")
	defer os.RemoveAll(config.RootDir)
	now := ttime.Now()

	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger())
	n.GenesisDoc().GenesisTime = now.Add(2 * time.Second)
	require.NoError(t, err)

	n.Start()
	startTime := ttime.Now()
	assert.Equal(t, true, startTime.After(n.GenesisDoc().GenesisTime))
}

func TestNodeSetAppVersion(t *testing.T) {
	config := cfg.ResetTestRoot("node_app_version_test")
	defer os.RemoveAll(config.RootDir)

	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err)

	// default config uses the kvstore app
	var appVersion version.Protocol = kvstore.ProtocolVersion

	// check version is set in state
	state := sm.LoadState(n.stateDB)
	assert.Equal(t, state.Version.Consensus.App, appVersion)

	// check version is set in node info
	assert.Equal(t, n.nodeInfo.(p2p.DefaultNodeInfo).ProtocolVersion.App, appVersion)
}

func TestNodeSetPrivValTCP(t *testing.T) {
	addr := "tcp://" + testFreeAddr(t)

	config := cfg.ResetTestRoot("node_priv_val_tcp_test")
	defer os.RemoveAll(config.RootDir)
	config.BaseConfig.ValidatorListenAddr = addr

	dialer := validator.DialTCPFn(addr, 100*time.Millisecond, ed25519.GenPrivKey())
	pvsc := validator.NewSignerServiceEndpoint(
		log.TestingLogger(),
		config.LeagueID(),
		types.NewMockPV(),
		dialer,
	)
	validator.SignerServiceEndpointTimeoutReadWrite(100 * time.Millisecond)(pvsc)

	go func() {
		err := pvsc.Start()
		if err != nil {
			panic(err)
		}
	}()
	defer pvsc.Stop()

	n, err := DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err)
	assert.IsType(t, &validator.SignerValidatorEndpoint{}, n.Validator())
}

// address without a protocol must result in error
func TestValidatorListenAddrNoProtocol(t *testing.T) {
	addrNoPrefix := testFreeAddr(t)

	config := cfg.ResetTestRoot("node_priv_val_tcp_test")
	defer os.RemoveAll(config.RootDir)
	config.BaseConfig.ValidatorListenAddr = addrNoPrefix

	_, err := DefaultNewNode(config, log.TestingLogger())
	assert.Error(t, err)
}

func TestNodeSetPrivValIPC(t *testing.T) {
	tmpfile := "/tmp/kms." + cmn.RandStr(6) + ".sock"
	defer os.Remove(tmpfile) // clean up

	config := cfg.ResetTestRoot("node_priv_val_tcp_test")
	defer os.RemoveAll(config.RootDir)
	config.BaseConfig.ValidatorListenAddr = "unix://" + tmpfile

	dialer := validator.DialUnixFn(tmpfile)
	pvsc := validator.NewSignerServiceEndpoint(
		log.TestingLogger(),
		config.LeagueID(),
		types.NewMockPV(),
		dialer,
	)
	validator.SignerServiceEndpointTimeoutReadWrite(100 * time.Millisecond)(pvsc)

	go func() {
		err := pvsc.Start()
		require.NoError(t, err)
	}()
	defer pvsc.Stop()

	n, err := DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err)
	assert.IsType(t, &validator.SignerValidatorEndpoint{}, n.Validator())

}

// testFreeAddr claims a free port so we don't block on listener being ready.
func testFreeAddr(t *testing.T) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	return fmt.Sprintf("127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)
}

// create a proposal block using real and full
// storage and evidence pool and validate it.
func TestCreateProposalBlock(t *testing.T) {
	config := cfg.ResetTestRoot("node_create_proposal")
	defer os.RemoveAll(config.RootDir)
	cc := proxy.NewLocalClientCreator(kvstore.NewKVStoreApplication())
	proxyApp := proxy.NewAppConns(cc)
	err := proxyApp.Start()
	require.Nil(t, err)
	defer proxyApp.Stop()

	logger := log.TestingLogger()

	var height int64 = 1
	state, stateDB := state(1, height)
	maxBytes := 16384
	state.ConsensusParams.Block.MaxBytes = int64(maxBytes)
	proposerAddr, _ := state.Validators.GetByIndex(0)

	// Make Storage
	memplMetrics := mempl.PrometheusMetrics("node_test")
	storage := mempl.NewStorage(
		config.Storage,
		proxyApp.Storage(),
		state.LastBlockHeight,
		mempl.WithMetrics(memplMetrics),
		mempl.WithPreCheck(sm.TxPreCheck(state)),
		mempl.WithPostCheck(sm.TxPostCheck(state)),
	)
	storage.SetLogger(logger)

	// Make EvidencePool
	types.RegisterMockEvidencesGlobal() // XXX!
	evidence.RegisterMockEvidences()
	evidenceDB := dbm.NewMemDB()
	evidencePool := evidence.NewEvidencePool(stateDB, evidenceDB)
	evidencePool.SetLogger(logger)

	// fill the evidence pool with more evidence
	// than can fit in a block
	minEvSize := 12
	numEv := (maxBytes / types.MaxEvidenceBytesDenominator) / minEvSize
	for i := 0; i < numEv; i++ {
		ev := types.NewMockRandomGoodEvidence(1, proposerAddr, cmn.RandBytes(minEvSize))
		err := evidencePool.AddEvidence(ev)
		assert.NoError(t, err)
	}

	// fill the storage with more txs
	// than can fit in a block
	txLength := 1000
	for i := 0; i < maxBytes/txLength; i++ {
		tx := cmn.RandBytes(txLength)
		err := storage.CheckTx(tx, nil)
		assert.NoError(t, err)
	}

	blockExec := sm.NewBlockExecutor(
		stateDB,
		logger,
		proxyApp.Consensus(),
		storage,
		evidencePool,
	)

	commit := types.NewCommit(types.BlockID{}, nil)
	block, _ := blockExec.CreateProposalBlock(
		height,
		state, commit,
		proposerAddr,
	)

	err = blockExec.ValidateBlock(state, block)
	assert.NoError(t, err)
}

func state(nVals int, height int64) (sm.State, dbm.DB) {
	vals := make([]types.GenesisValidator, nVals)
	for i := 0; i < nVals; i++ {
		secret := []byte(fmt.Sprintf("test%d", i))
		pk := ed25519.GenPrivKeyFromSecret(secret)
		vals[i] = types.GenesisValidator{
			pk.PubKey().Address(),
			pk.PubKey(),
			1000,
			fmt.Sprintf("test%d", i),
		}
	}
	s, _ := sm.MakeGenesisState(&types.GenesisDoc{
		LeagueID:   "test-chain",
		Validators: vals,
		AppHash:    nil,
	})

	// save validators to db for 2 heights
	stateDB := dbm.NewMemDB()
	sm.SaveState(stateDB, s)

	for i := 1; i < int(height); i++ {
		s.LastBlockHeight++
		s.LastValidators = s.Validators.Copy()
		sm.SaveState(stateDB, s)
	}
	return s, stateDB
}
