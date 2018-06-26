package consensus

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/teragrid/asura/example/kvstore"
	bc "github.com/teragrid/teragrid/blockchain"
	cfg "github.com/teragrid/teragrid/config"
	"github.com/teragrid/teragrid/proxy"
	sm "github.com/teragrid/teragrid/state"
	"github.com/teragrid/teragrid/types"
	pvm "github.com/teragrid/teragrid/types/priv_validator"
	auto "github.com/teragrid/teralibs/autofile"
	cmn "github.com/teragrid/teralibs/common"
	"github.com/teragrid/teralibs/db"
	"github.com/teragrid/teralibs/log"
)

// WALWithNBlocks generates a consensus WAL. It does this by spining up a
// stripped down version of node (proxy app, event bus, consensus state) with a
// persistent kvstore application and special consensus wal instance
// (byteBufferWAL) and waits until numBlocks are created. Then it returns a WAL
// content.
func WALWithNBlocks(numBlocks int) (data []byte, err error) {
	config := getConfig().ChainConfigs[0]

	app := kvstore.NewPersistentKVStoreApplication(filepath.Join(config.DBDir(), "wal_generator"))

	logger := log.TestingLogger().With("wal_generator", "wal_generator")
	logger.Info("generating WAL (last height msg excluded)", "numBlocks", numBlocks)

	/////////////////////////////////////////////////////////////////////////////
	// COPY PASTE FROM node.go WITH A FEW MODIFICATIONS
	// NOTE: we can't import node package because of circular dependency
	privValidatorFile := config.PrivValidatorFile()
	privValidator := pvm.LoadOrGenFilePV(privValidatorFile)
	genDoc, err := types.GenesisDocFromFile(config.GenesisFile())
	if err != nil {
		return nil, errors.Wrap(err, "failed to read genesis file")
	}
	stateDB := db.NewMemDB()
	blockStoreDB := db.NewMemDB()
	state, err := sm.MakeGenesisState(genDoc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make genesis state")
	}
	blockStore := bc.NewBlockStore(blockStoreDB)
	handshaker := NewHandshaker(stateDB, state, blockStore, genDoc.AppState())
	proxyApp := proxy.NewAppConns(proxy.NewLocalClientCreator(app), handshaker)
	proxyApp.SetLogger(logger.With("module", "proxy"))
	if err := proxyApp.Start(); err != nil {
		return nil, errors.Wrap(err, "failed to start proxy app connections")
	}
	defer proxyApp.Stop()
	eventBus := types.NewEventBus()
	eventBus.SetLogger(logger.With("module", "events"))
	if err := eventBus.Start(); err != nil {
		return nil, errors.Wrap(err, "failed to start event bus")
	}
	defer eventBus.Stop()
	mempool := types.MockMempool{}
	evpool := types.MockEvidencePool{}
	blockExec := sm.NewBlockExecutor(stateDB, log.TestingLogger(), proxyApp.Consensus(), mempool, evpool)
	consensusState := NewConsensusState(config.Consensus, state.Copy(), blockExec, blockStore, mempool, evpool)
	consensusState.SetLogger(logger)
	consensusState.SetEventBus(eventBus)
	if privValidator != nil {
		consensusState.SetPrivValidator(privValidator)
	}
	// END OF COPY PASTE
	/////////////////////////////////////////////////////////////////////////////

	// set consensus wal to buffered WAL, which will write all incoming msgs to buffer
	var b bytes.Buffer
	wr := bufio.NewWriter(&b)
	numBlocksWritten := make(chan struct{})
	wal := newByteBufferWAL(logger, NewWALEncoder(wr), int64(numBlocks), numBlocksWritten)
	// see wal.go#103
	wal.Save(EndHeightMessage{0})
	consensusState.wal = wal

	if err := consensusState.Start(); err != nil {
		return nil, errors.Wrap(err, "failed to start consensus state")
	}
	defer consensusState.Stop()

	select {
	case <-numBlocksWritten:
		wr.Flush()
		return b.Bytes(), nil
	case <-time.After(1 * time.Minute):
		wr.Flush()
		return b.Bytes(), fmt.Errorf("waited too long for teragrid to produce %d blocks (grep logs for `wal_generator`)", numBlocks)
	}
}

// f**ing long, but unique for each test
func makePathname() string {
	// get path
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// fmt.Println(p)
	sep := string(filepath.Separator)
	return strings.Replace(p, sep, "_", -1)
}

func randPort() int {
	// returns between base and base + spread
	base, spread := 20000, 20000
	return base + cmn.RandIntn(spread)
}

func makeAddrs() (string, string, string) {
	start := randPort()
	return fmt.Sprintf("tcp://0.0.0.0:%d", start),
		fmt.Sprintf("tcp://0.0.0.0:%d", start+1),
		fmt.Sprintf("tcp://0.0.0.0:%d", start+2)
}

// getConfig returns a config for test cases
func getConfig() *cfg.Config {
	pathname := makePathname()
	c := cfg.ResetTestRoot(fmt.Sprintf("%s_%d", pathname, cmn.RandInt()))

	// and we use random ports to run in parallel
	tm, rpc, grpc := makeAddrs()
	c.ChainConfigs[0].P2P.ListenAddress = tm
	c.ChainConfigs[0].RPC.ListenAddress = rpc
	c.ChainConfigs[0].RPC.GRPCListenAddress = grpc
	return c
}

// byteBufferWAL is a WAL which writes all msgs to a byte buffer. Writing stops
// when the heightToStop is reached. Client will be notified via
// signalWhenStopsTo channel.
type byteBufferWAL struct {
	enc               *WALEncoder
	stopped           bool
	heightToStop      int64
	signalWhenStopsTo chan<- struct{}

	logger log.Logger
}

// needed for determinism
var fixedTime, _ = time.Parse(time.RFC3339, "2017-01-02T15:04:05Z")

func newByteBufferWAL(logger log.Logger, enc *WALEncoder, nBlocks int64, signalStop chan<- struct{}) *byteBufferWAL {
	return &byteBufferWAL{
		enc:               enc,
		heightToStop:      nBlocks,
		signalWhenStopsTo: signalStop,
		logger:            logger,
	}
}

// Save writes message to the internal buffer except when heightToStop is
// reached, in which case it will signal the caller via signalWhenStopsTo and
// skip writing.
func (w *byteBufferWAL) Save(m WALMessage) {
	if w.stopped {
		w.logger.Debug("WAL already stopped. Not writing message", "msg", m)
		return
	}

	if endMsg, ok := m.(EndHeightMessage); ok {
		w.logger.Debug("WAL write end height message", "height", endMsg.Height, "stopHeight", w.heightToStop)
		if endMsg.Height == w.heightToStop {
			w.logger.Debug("Stopping WAL at height", "height", endMsg.Height)
			w.signalWhenStopsTo <- struct{}{}
			w.stopped = true
			return
		}
	}

	w.logger.Debug("WAL Write Message", "msg", m)
	err := w.enc.Encode(&TimedWALMessage{fixedTime, m})
	if err != nil {
		panic(fmt.Sprintf("failed to encode the msg %v", m))
	}
}

func (w *byteBufferWAL) Group() *auto.Group {
	panic("not implemented")
}
func (w *byteBufferWAL) SearchForEndHeight(height int64, options *WALSearchOptions) (gr *auto.GroupReader, found bool, err error) {
	return nil, false, nil
}

func (w *byteBufferWAL) Start() error { return nil }
func (w *byteBufferWAL) Stop() error  { return nil }
func (w *byteBufferWAL) Wait()        {}
