package cell

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	bc "github.com/teragrid/dgrid/core/blockchain"
	"github.com/teragrid/dgrid/core/blockchain/p2p"
	"github.com/teragrid/dgrid/core/blockchain/p2p/pex"
	cfg "github.com/teragrid/dgrid/core/config"
	cs "github.com/teragrid/dgrid/core/consensus"
	"github.com/teragrid/dgrid/core/types"
	ttime "github.com/teragrid/dgrid/core/types/time"
	"github.com/teragrid/dgrid/evidence"
	cmn "github.com/teragrid/dgrid/pkg/common"
	"github.com/teragrid/dgrid/pkg/crypto/ed25519"
	dbm "github.com/teragrid/dgrid/pkg/db"
	"github.com/teragrid/dgrid/pkg/log"
	tpubsub "github.com/teragrid/dgrid/pkg/pubsub"
	"github.com/teragrid/dgrid/proxy"
	rpccore "github.com/teragrid/dgrid/rpc/core"
	ctypes "github.com/teragrid/dgrid/rpc/core/types"
	grpccore "github.com/teragrid/dgrid/rpc/grpc"
	rpcserver "github.com/teragrid/dgrid/rpc/lib/server"
	sm "github.com/teragrid/dgrid/state"
	"github.com/teragrid/dgrid/state/txindex"
	"github.com/teragrid/dgrid/state/txindex/kv"
	"github.com/teragrid/dgrid/state/txindex/null"
	storage "github.com/teragrid/dgrid/storage"
	"github.com/teragrid/dgrid/third_party/amino"
	"github.com/teragrid/dgrid/version"
	"github.com/teragrid/dgridcore/consensus/validator"
)

//------------------------------------------------------------------------------

// DBContext specifies config information for loading a new DB.
type DBContext struct {
	ID     string
	Config *cfg.Config
}

// DBProvider takes a DBContext and returns an instantiated DB.
type DBProvider func(*DBContext) (dbm.DB, error)

// DefaultDBProvider returns a database using the DBBackend and DBDir
// specified in the ctx.Config.
func DefaultDBProvider(ctx *DBContext) (dbm.DB, error) {
	dbType := dbm.DBBackendType(ctx.Config.DBBackend)
	return dbm.NewDB(ctx.ID, dbType, ctx.Config.DBDir()), nil
}

// GenesisDocProvider returns a GenesisDoc.
// It allows the GenesisDoc to be pulled from sources other than the
// filesystem, for instance from a distributed key-value store cluster.
type GenesisDocProvider func() (*types.GenesisDoc, error)

// DefaultGenesisDocProviderFunc returns a GenesisDocProvider that loads
// the GenesisDoc from the config.GenesisFile() on the filesystem.
func DefaultGenesisDocProviderFunc(config *cfg.Config) GenesisDocProvider {
	return func() (*types.GenesisDoc, error) {
		return types.GenesisDocFromFile(config.GenesisFile())
	}
}

// CellProvider takes a config and a logger and returns a ready to go Cell.
type CellProvider func(*cfg.Config, log.Logger) (*Cell, error)

// DefaultNewCell returns a Dgrid cell with default settings for the
// Validator, ClientCreator, GenesisDoc, and DBProvider.
// It implements CellProvider.
func DefaultNewCell(config *cfg.Config, logger log.Logger) (*Cell, error) {
	// Generate cell PrivKey
	cellKey, err := p2p.LoadOrGenCellKey(config.CellKeyFile())
	if err != nil {
		return nil, err
	}

	// Convert old Validator if it exists.
	oldPrivVal := config.OldValidatorFile()
	newPrivValKey := config.ValidatorKeyFile()
	newPrivValState := config.ValidatorStateFile()
	if _, err := os.Stat(oldPrivVal); !os.IsNotExist(err) {
		oldPV, err := validator.LoadOldFilePV(oldPrivVal)
		if err != nil {
			return nil, fmt.Errorf("Error reading OldValidator from %v: %v\n", oldPrivVal, err)
		}
		logger.Info("Upgrading Validator file",
			"old", oldPrivVal,
			"newKey", newPrivValKey,
			"newState", newPrivValState,
		)
		oldPV.Upgrade(newPrivValKey, newPrivValState)
	}

	return NewCell(config,
		validator.LoadOrGenFilePV(newPrivValKey, newPrivValState),
		cellKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.Asura, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		logger,
	)
}

// MetricsProvider returns a consensus, p2p and storage Metrics.
type MetricsProvider func(leagueID string) (*cs.Metrics, *p2p.Metrics, *storage.Metrics, *sm.Metrics)

// DefaultMetricsProvider returns Metrics build using Prometheus client library
// if Prometheus is enabled. Otherwise, it returns no-op Metrics.
func DefaultMetricsProvider(config *cfg.InstrumentationConfig) MetricsProvider {
	return func(leagueID string) (*cs.Metrics, *p2p.Metrics, *storage.Metrics, *sm.Metrics) {
		if config.Prometheus {
			return cs.PrometheusMetrics(config.Namespace, "chain_id", leagueID),
				p2p.PrometheusMetrics(config.Namespace, "chain_id", leagueID),
				storage.PrometheusMetrics(config.Namespace, "chain_id", leagueID),
				sm.PrometheusMetrics(config.Namespace, "chain_id", leagueID)
		}
		return cs.NopMetrics(), p2p.NopMetrics(), storage.NopMetrics(), sm.NopMetrics()
	}
}

//------------------------------------------------------------------------------

// Cell is the highest level interface to a full Dgrid cell.
// It includes all configuration information and running services.
type Cell struct {
	cmn.BaseService

	// config
	config     *cfg.Config
	genesisDoc *types.GenesisDoc // initial validator set
	validator  types.Validator   // local cell's validator key

	// network
	transport   *p2p.MultiplexTransport
	sw          *p2p.Switch  // p2p connections
	addrBook    pex.AddrBook // known peers
	cellInfo    p2p.CellInfo
	cellKey     *p2p.CellKey // our cell privkey
	isListening bool

	// services
	eventBus         *types.EventBus // pub/sub for services
	stateDB          dbm.DB
	blockStore       *bc.BlockStore          // store the blockchain to disk
	bcReactor        *bc.BlockchainReactor   // for fast-syncing
	storageReactor   *storage.StorageReactor // for gossipping transactions
	consensusState   *cs.ConsensusState      // latest consensus state
	consensusReactor *cs.ConsensusReactor    // for participating in the consensus
	evidencePool     *evidence.EvidencePool  // tracking evidence
	proxyApp         proxy.AppConns          // connection to the application
	rpcListeners     []net.Listener          // rpc servers
	txIndexer        txindex.TxIndexer
	indexerService   *txindex.IndexerService
	prometheusSrv    *http.Server
}

// NewCell returns a new, ready to go, Dgrid Cell.
func NewCell(config *cfg.Config,
	validator types.Validator,
	cellKey *p2p.CellKey,
	clientCreator proxy.ClientCreator,
	genesisDocProvider GenesisDocProvider,
	dbProvider DBProvider,
	metricsProvider MetricsProvider,
	logger log.Logger) (*Cell, error) {

	// Get BlockStore
	blockStoreDB, err := dbProvider(&DBContext{"blockstore", config})
	if err != nil {
		return nil, err
	}
	blockStore := bc.NewBlockStore(blockStoreDB)

	// Get State
	stateDB, err := dbProvider(&DBContext{"state", config})
	if err != nil {
		return nil, err
	}

	// Get genesis doc
	// TODO: move to state package?
	genDoc, err := loadGenesisDoc(stateDB)
	if err != nil {
		genDoc, err = genesisDocProvider()
		if err != nil {
			return nil, err
		}
		// save genesis doc to prevent a certain class of user errors (e.g. when it
		// was changed, accidentally or not). Also good for audit trail.
		saveGenesisDoc(stateDB, genDoc)
	}

	state, err := sm.LoadStateFromDBOrGenesisDoc(stateDB, genDoc)
	if err != nil {
		return nil, err
	}

	// Create the proxyApp and establish connections to the Asura app (consensus, storage, query).
	proxyApp := proxy.NewAppConns(clientCreator)
	proxyApp.SetLogger(logger.With("module", "proxy"))
	if err := proxyApp.Start(); err != nil {
		return nil, fmt.Errorf("Error starting proxy app connections: %v", err)
	}

	// EventBus and IndexerService must be started before the handshake because
	// we might need to index the txs of the replayed block as this might not have happened
	// when the cell stopped last time (i.e. the cell stopped after it saved the block
	// but before it indexed the txs, or, endblocker panicked)
	eventBus := types.NewEventBus()
	eventBus.SetLogger(logger.With("module", "events"))

	err = eventBus.Start()
	if err != nil {
		return nil, err
	}

	// Transaction indexing
	var txIndexer txindex.TxIndexer
	switch config.TxIndex.Indexer {
	case "kv":
		store, err := dbProvider(&DBContext{"tx_index", config})
		if err != nil {
			return nil, err
		}
		if config.TxIndex.IndexTags != "" {
			txIndexer = kv.NewTxIndex(store, kv.IndexTags(splitAndTrimEmpty(config.TxIndex.IndexTags, ",", " ")))
		} else if config.TxIndex.IndexAllTags {
			txIndexer = kv.NewTxIndex(store, kv.IndexAllTags())
		} else {
			txIndexer = kv.NewTxIndex(store)
		}
	default:
		txIndexer = &null.TxIndex{}
	}

	indexerService := txindex.NewIndexerService(txIndexer, eventBus)
	indexerService.SetLogger(logger.With("module", "txindex"))

	err = indexerService.Start()
	if err != nil {
		return nil, err
	}

	// Create the handshaker, which calls RequestInfo, sets the AppVersion on the state,
	// and replays any blocks as necessary to sync teragrid with the app.
	consensusLogger := logger.With("module", "consensus")
	handshaker := cs.NewHandshaker(stateDB, state, blockStore, genDoc)
	handshaker.SetLogger(consensusLogger)
	handshaker.SetEventBus(eventBus)
	if err := handshaker.Handshake(proxyApp); err != nil {
		return nil, fmt.Errorf("Error during handshake: %v", err)
	}

	// Reload the state. It will have the Version.Consensus.App set by the
	// Handshake, and may have other modifications as well (ie. depending on
	// what happened during block replay).
	state = sm.LoadState(stateDB)

	// Log the version info.
	logger.Info("Version info",
		"software", version.TMCoreSemVer,
		"block", version.BlockProtocol,
		"p2p", version.P2PProtocol,
	)

	// If the state and software differ in block version, at least log it.
	if state.Version.Consensus.Block != version.BlockProtocol {
		logger.Info("Software and state have different block protocols",
			"software", version.BlockProtocol,
			"state", state.Version.Consensus.Block,
		)
	}

	if config.ValidatorListenAddr != "" {
		// If an address is provided, listen on the socket for a connection from an
		// external signing process.
		// FIXME: we should start services inside OnStart
		validator, err = createAndStartValidatorSocketClient(config.ValidatorListenAddr, logger)
		if err != nil {
			return nil, errors.Wrap(err, "Error with private validator socket client")
		}
	}

	// Decide whether to fast-sync or not
	// We don't fast-sync when the only validator is us.
	fastSync := config.FastSync
	if state.Validators.Size() == 1 {
		addr, _ := state.Validators.GetByIndex(0)
		privValAddr := validator.GetPubKey().Address()
		if bytes.Equal(privValAddr, addr) {
			fastSync = false
		}
	}

	pubKey := validator.GetPubKey()
	addr := pubKey.Address()
	// Log whether this cell is a validator or an observer
	if state.Validators.HasAddress(addr) {
		consensusLogger.Info("This cell is a validator", "addr", addr, "pubKey", pubKey)
	} else {
		consensusLogger.Info("This cell is not a validator", "addr", addr, "pubKey", pubKey)
	}

	csMetrics, p2pMetrics, memplMetrics, smMetrics := metricsProvider(genDoc.LeagueID)

	// Make StorageReactor
	storage := storage.NewStorage(
		config.Storage,
		proxyApp.Storage(),
		state.LastBlockHeight,
		storage.WithMetrics(memplMetrics),
		storage.WithPreCheck(sm.TxPreCheck(state)),
		storage.WithPostCheck(sm.TxPostCheck(state)),
	)
	storageLogger := logger.With("module", "storage")
	storage.SetLogger(storageLogger)
	if config.Storage.WalEnabled() {
		storage.InitWAL() // no need to have the storage wal during tests
	}
	storageReactor := storage.NewStorageReactor(config.Storage, storage)
	storageReactor.SetLogger(storageLogger)

	if config.Consensus.WaitForTxs() {
		storage.EnableTxsAvailable()
	}

	// Make Evidence Reactor
	evidenceDB, err := dbProvider(&DBContext{"evidence", config})
	if err != nil {
		return nil, err
	}
	evidenceLogger := logger.With("module", "evidence")
	evidencePool := evidence.NewEvidencePool(stateDB, evidenceDB)
	evidencePool.SetLogger(evidenceLogger)
	evidenceReactor := evidence.NewEvidenceReactor(evidencePool)
	evidenceReactor.SetLogger(evidenceLogger)

	blockExecLogger := logger.With("module", "state")
	// make block executor for consensus and blockchain reactors to execute blocks
	blockExec := sm.NewBlockExecutor(
		stateDB,
		blockExecLogger,
		proxyApp.Consensus(),
		storage,
		evidencePool,
		sm.BlockExecutorWithMetrics(smMetrics),
	)

	// Make BlockchainReactor
	bcReactor := bc.NewBlockchainReactor(state.Copy(), blockExec, blockStore, fastSync)
	bcReactor.SetLogger(logger.With("module", "blockchain"))

	// Make ConsensusReactor
	consensusState := cs.NewConsensusState(
		config.Consensus,
		state.Copy(),
		blockExec,
		blockStore,
		storage,
		evidencePool,
		cs.StateMetrics(csMetrics),
	)
	consensusState.SetLogger(consensusLogger)
	if validator != nil {
		consensusState.SetValidator(validator)
	}
	consensusReactor := cs.NewConsensusReactor(consensusState, fastSync, cs.ReactorMetrics(csMetrics))
	consensusReactor.SetLogger(consensusLogger)

	// services which will be publishing and/or subscribing for messages (events)
	// consensusReactor will set it on consensusState and blockExecutor
	consensusReactor.SetEventBus(eventBus)

	p2pLogger := logger.With("module", "p2p")
	cellInfo, err := makeCellInfo(
		config,
		cellKey.ID(),
		txIndexer,
		genDoc.LeagueID,
		p2p.NewProtocolVersion(
			version.P2PProtocol, // global
			state.Version.Consensus.Block,
			state.Version.Consensus.App,
		),
	)
	if err != nil {
		return nil, err
	}

	// Setup Transport.
	var (
		mConnConfig = p2p.MConnConfig(config.P2P)
		transport   = p2p.NewMultiplexTransport(cellInfo, *cellKey, mConnConfig)
		connFilters = []p2p.ConnFilterFunc{}
		peerFilters = []p2p.PeerFilterFunc{}
	)

	if !config.P2P.AllowDuplicateIP {
		connFilters = append(connFilters, p2p.ConnDuplicateIPFilter())
	}

	// Filter peers by addr or pubkey with an Asura query.
	// If the query return code is OK, add peer.
	if config.FilterPeers {
		connFilters = append(
			connFilters,
			// Asura query for address filtering.
			func(_ p2p.ConnSet, c net.Conn, _ []net.IP) error {
				res, err := proxyApp.Query().QuerySync(asura.RequestQuery{
					Path: fmt.Sprintf("/p2p/filter/addr/%s", c.RemoteAddr().String()),
				})
				if err != nil {
					return err
				}
				if res.IsErr() {
					return fmt.Errorf("Error querying asura app: %v", res)
				}

				return nil
			},
		)

		peerFilters = append(
			peerFilters,
			// Asura query for ID filtering.
			func(_ p2p.IPeerSet, p p2p.Peer) error {
				res, err := proxyApp.Query().QuerySync(asura.RequestQuery{
					Path: fmt.Sprintf("/p2p/filter/id/%s", p.ID()),
				})
				if err != nil {
					return err
				}
				if res.IsErr() {
					return fmt.Errorf("Error querying asura app: %v", res)
				}

				return nil
			},
		)
	}

	p2p.MultiplexTransportConnFilters(connFilters...)(transport)

	// Setup Switch.
	sw := p2p.NewSwitch(
		config.P2P,
		transport,
		p2p.WithMetrics(p2pMetrics),
		p2p.SwitchPeerFilters(peerFilters...),
	)
	sw.SetLogger(p2pLogger)
	sw.AddReactor("MEMPOOL", storageReactor)
	sw.AddReactor("BLOCKCHAIN", bcReactor)
	sw.AddReactor("CONSENSUS", consensusReactor)
	sw.AddReactor("EVIDENCE", evidenceReactor)
	sw.SetCellInfo(cellInfo)
	sw.SetCellKey(cellKey)

	p2pLogger.Info("P2P Cell ID", "ID", cellKey.ID(), "file", config.CellKeyFile())

	// Optionally, start the pex reactor
	//
	// TODO:
	//
	// We need to set Seeds and PersistentPeers on the switch,
	// since it needs to be able to use these (and their DNS names)
	// even if the PEX is off. We can include the DNS name in the NetAddress,
	// but it would still be nice to have a clear list of the current "PersistentPeers"
	// somewhere that we can return with net_info.
	//
	// If PEX is on, it should handle dialing the seeds. Otherwise the switch does it.
	// Note we currently use the addrBook regardless at least for AddOurAddress
	addrBook := pex.NewAddrBook(config.P2P.AddrBookFile(), config.P2P.AddrBookStrict)

	// Add ourselves to addrbook to prevent dialing ourselves
	addrBook.AddOurAddress(sw.NetAddress())

	addrBook.SetLogger(p2pLogger.With("book", config.P2P.AddrBookFile()))
	if config.P2P.PexReactor {
		// TODO persistent peers ? so we can have their DNS addrs saved
		pexReactor := pex.NewPEXReactor(addrBook,
			&pex.PEXReactorConfig{
				Seeds:    splitAndTrimEmpty(config.P2P.Seeds, ",", " "),
				SeedMode: config.P2P.SeedMode,
				// See consensus/reactor.go: blocksToContributeToBecomeGoodPeer 10000
				// blocks assuming 10s blocks ~ 28 hours.
				// TODO (melekes): make it dynamic based on the actual block latencies
				// from the live network.
				// https://github.com/teragrid/dgrid/issues/3523
				SeedDisconnectWaitPeriod: 28 * time.Hour,
			})
		pexReactor.SetLogger(logger.With("module", "pex"))
		sw.AddReactor("PEX", pexReactor)
	}

	sw.SetAddrBook(addrBook)

	// run the profile server
	profileHost := config.ProfListenAddress
	if profileHost != "" {
		go func() {
			logger.Error("Profile server", "err", http.ListenAndServe(profileHost, nil))
		}()
	}

	cell := &Cell{
		config:     config,
		genesisDoc: genDoc,
		validator:  validator,

		transport: transport,
		sw:        sw,
		addrBook:  addrBook,
		cellInfo:  cellInfo,
		cellKey:   cellKey,

		stateDB:          stateDB,
		blockStore:       blockStore,
		bcReactor:        bcReactor,
		storageReactor:   storageReactor,
		consensusState:   consensusState,
		consensusReactor: consensusReactor,
		evidencePool:     evidencePool,
		proxyApp:         proxyApp,
		txIndexer:        txIndexer,
		indexerService:   indexerService,
		eventBus:         eventBus,
	}
	cell.BaseService = *cmn.NewBaseService(logger, "Cell", cell)
	return cell, nil
}

// OnStart starts the Cell. It implements cmn.Service.
func (n *Cell) OnStart() error {
	now := ttime.Now()
	genTime := n.genesisDoc.GenesisTime
	if genTime.After(now) {
		n.Logger.Info("Genesis time is in the future. Sleeping until then...", "genTime", genTime)
		time.Sleep(genTime.Sub(now))
	}

	// Add private IDs to addrbook to block those peers being added
	n.addrBook.AddPrivateIDs(splitAndTrimEmpty(n.config.P2P.PrivatePeerIDs, ",", " "))

	// Start the RPC server before the P2P server
	// so we can eg. receive txs for the first block
	if n.config.RPC.ListenAddress != "" {
		listeners, err := n.startRPC()
		if err != nil {
			return err
		}
		n.rpcListeners = listeners
	}

	if n.config.Instrumentation.Prometheus &&
		n.config.Instrumentation.PrometheusListenAddr != "" {
		n.prometheusSrv = n.startPrometheusServer(n.config.Instrumentation.PrometheusListenAddr)
	}

	// Start the transport.
	addr, err := p2p.NewNetAddressStringWithOptionalID(n.config.P2P.ListenAddress)
	if err != nil {
		return err
	}
	if err := n.transport.Listen(*addr); err != nil {
		return err
	}

	n.isListening = true

	// Start the switch (the P2P server).
	err = n.sw.Start()
	if err != nil {
		return err
	}

	// Always connect to persistent peers
	if n.config.P2P.PersistentPeers != "" {
		err = n.sw.DialPeersAsync(n.addrBook, splitAndTrimEmpty(n.config.P2P.PersistentPeers, ",", " "), true)
		if err != nil {
			return err
		}
	}

	return nil
}

// OnStop stops the Cell. It implements cmn.Service.
func (n *Cell) OnStop() {
	n.BaseService.OnStop()

	n.Logger.Info("Stopping Cell")

	// first stop the non-reactor services
	n.eventBus.Stop()
	n.indexerService.Stop()

	// now stop the reactors
	// TODO: gracefully disconnect from peers.
	n.sw.Stop()

	// stop storage WAL
	if n.config.Storage.WalEnabled() {
		n.storageReactor.Storage.CloseWAL()
	}

	if err := n.transport.Close(); err != nil {
		n.Logger.Error("Error closing transport", "err", err)
	}

	n.isListening = false

	// finally stop the listeners / external services
	for _, l := range n.rpcListeners {
		n.Logger.Info("Closing rpc listener", "listener", l)
		if err := l.Close(); err != nil {
			n.Logger.Error("Error closing listener", "listener", l, "err", err)
		}
	}

	if pvsc, ok := n.validator.(cmn.Service); ok {
		pvsc.Stop()
	}

	if n.prometheusSrv != nil {
		if err := n.prometheusSrv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			n.Logger.Error("Prometheus HTTP server Shutdown", "err", err)
		}
	}
}

// ConfigureRPC sets all variables in rpccore so they will serve
// rpc calls from this cell
func (n *Cell) ConfigureRPC() {
	rpccore.SetStateDB(n.stateDB)
	rpccore.SetBlockStore(n.blockStore)
	rpccore.SetConsensusState(n.consensusState)
	rpccore.SetStorage(n.storageReactor.Storage)
	rpccore.SetEvidencePool(n.evidencePool)
	rpccore.SetP2PPeers(n.sw)
	rpccore.SetP2PTransport(n)
	pubKey := n.validator.GetPubKey()
	rpccore.SetPubKey(pubKey)
	rpccore.SetGenesisDoc(n.genesisDoc)
	rpccore.SetAddrBook(n.addrBook)
	rpccore.SetProxyAppQuery(n.proxyApp.Query())
	rpccore.SetTxIndexer(n.txIndexer)
	rpccore.SetConsensusReactor(n.consensusReactor)
	rpccore.SetEventBus(n.eventBus)
	rpccore.SetLogger(n.Logger.With("module", "rpc"))
	rpccore.SetConfig(*n.config.RPC)
}

func (n *Cell) startRPC() ([]net.Listener, error) {
	n.ConfigureRPC()
	listenAddrs := splitAndTrimEmpty(n.config.RPC.ListenAddress, ",", " ")
	coreCodec := amino.NewCodec()
	ctypes.RegisterAmino(coreCodec)

	if n.config.RPC.Unsafe {
		rpccore.AddUnsafeRoutes()
	}

	// we may expose the rpc over both a unix and tcp socket
	listeners := make([]net.Listener, len(listenAddrs))
	for i, listenAddr := range listenAddrs {
		mux := http.NewServeMux()
		rpcLogger := n.Logger.With("module", "rpc-server")
		wmLogger := rpcLogger.With("protocol", "websocket")
		wm := rpcserver.NewWebsocketManager(rpccore.Routes, coreCodec,
			rpcserver.OnDisconnect(func(remoteAddr string) {
				err := n.eventBus.UnsubscribeAll(context.Background(), remoteAddr)
				if err != nil && err != tpubsub.ErrSubscriptionNotFound {
					wmLogger.Error("Failed to unsubscribe addr from events", "addr", remoteAddr, "err", err)
				}
			}))
		wm.SetLogger(wmLogger)
		mux.HandleFunc("/websocket", wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(mux, rpccore.Routes, coreCodec, rpcLogger)

		config := rpcserver.DefaultConfig()
		config.MaxOpenConnections = n.config.RPC.MaxOpenConnections
		// If necessary adjust global WriteTimeout to ensure it's greater than
		// TimeoutBroadcastTxCommit.
		if config.WriteTimeout <= n.config.RPC.TimeoutBroadcastTxCommit {
			config.WriteTimeout = n.config.RPC.TimeoutBroadcastTxCommit + 1*time.Second
		}

		listener, err := rpcserver.Listen(
			listenAddr,
			config,
		)
		if err != nil {
			return nil, err
		}

		var rootHandler http.Handler = mux
		if n.config.RPC.IsCorsEnabled() {
			corsMiddleware := cors.New(cors.Options{
				AllowedOrigins: n.config.RPC.CORSAllowedOrigins,
				AllowedMethods: n.config.RPC.CORSAllowedMethods,
				AllowedHeaders: n.config.RPC.CORSAllowedHeaders,
			})
			rootHandler = corsMiddleware.Handler(mux)
		}
		if n.config.RPC.IsTLSEnabled() {
			go rpcserver.StartHTTPAndTLSServer(
				listener,
				rootHandler,
				n.config.RPC.CertFile(),
				n.config.RPC.KeyFile(),
				rpcLogger,
				config,
			)
		} else {
			go rpcserver.StartHTTPServer(
				listener,
				rootHandler,
				rpcLogger,
				config,
			)
		}

		listeners[i] = listener
	}

	// we expose a simplified api over grpc for convenience to app devs
	grpcListenAddr := n.config.RPC.GRPCListenAddress
	if grpcListenAddr != "" {
		config := rpcserver.DefaultConfig()
		config.MaxOpenConnections = n.config.RPC.MaxOpenConnections
		listener, err := rpcserver.Listen(grpcListenAddr, config)
		if err != nil {
			return nil, err
		}
		go grpccore.StartGRPCServer(listener)
		listeners = append(listeners, listener)
	}

	return listeners, nil
}

// startPrometheusServer starts a Prometheus HTTP server, listening for metrics
// collectors on addr.
func (n *Cell) startPrometheusServer(addr string) *http.Server {
	srv := &http.Server{
		Addr: addr,
		Handler: promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer, promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{MaxRequestsInFlight: n.config.Instrumentation.MaxOpenConnections},
			),
		),
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			n.Logger.Error("Prometheus HTTP server ListenAndServe", "err", err)
		}
	}()
	return srv
}

// Switch returns the Cell's Switch.
func (n *Cell) Switch() *p2p.Switch {
	return n.sw
}

// BlockStore returns the Cell's BlockStore.
func (n *Cell) BlockStore() *bc.BlockStore {
	return n.blockStore
}

// ConsensusState returns the Cell's ConsensusState.
func (n *Cell) ConsensusState() *cs.ConsensusState {
	return n.consensusState
}

// ConsensusReactor returns the Cell's ConsensusReactor.
func (n *Cell) ConsensusReactor() *cs.ConsensusReactor {
	return n.consensusReactor
}

// StorageReactor returns the Cell's StorageReactor.
func (n *Cell) StorageReactor() *storage.StorageReactor {
	return n.storageReactor
}

// EvidencePool returns the Cell's EvidencePool.
func (n *Cell) EvidencePool() *evidence.EvidencePool {
	return n.evidencePool
}

// EventBus returns the Cell's EventBus.
func (n *Cell) EventBus() *types.EventBus {
	return n.eventBus
}

// Validator returns the Cell's Validator.
// XXX: for convenience only!
func (n *Cell) Validator() types.Validator {
	return n.validator
}

// GenesisDoc returns the Cell's GenesisDoc.
func (n *Cell) GenesisDoc() *types.GenesisDoc {
	return n.genesisDoc
}

// ProxyApp returns the Cell's AppConns, representing its connections to the Asura application.
func (n *Cell) ProxyApp() proxy.AppConns {
	return n.proxyApp
}

// Config returns the Cell's config.
func (n *Cell) Config() *cfg.Config {
	return n.config
}

//------------------------------------------------------------------------------

func (n *Cell) Listeners() []string {
	return []string{
		fmt.Sprintf("Listener(@%v)", n.config.P2P.ExternalAddress),
	}
}

func (n *Cell) IsListening() bool {
	return n.isListening
}

// CellInfo returns the Cell's Info from the Switch.
func (n *Cell) CellInfo() p2p.CellInfo {
	return n.cellInfo
}

func makeCellInfo(
	config *cfg.Config,
	cellID p2p.ID,
	txIndexer txindex.TxIndexer,
	leagueID string,
	protocolVersion p2p.ProtocolVersion,
) (p2p.CellInfo, error) {
	txIndexerStatus := "on"
	if _, ok := txIndexer.(*null.TxIndex); ok {
		txIndexerStatus = "off"
	}
	cellInfo := p2p.DefaultCellInfo{
		ProtocolVersion: protocolVersion,
		ID_:             cellID,
		Network:         leagueID,
		Version:         version.TMCoreSemVer,
		Channels: []byte{
			bc.BlockchainChannel,
			cs.StateChannel, cs.DataChannel, cs.VoteChannel, cs.VoteSetBitsChannel,
			storage.StorageChannel,
			evidence.EvidenceChannel,
		},
		Moniker: config.Moniker,
		Other: p2p.DefaultCellInfoOther{
			TxIndex:    txIndexerStatus,
			RPCAddress: config.RPC.ListenAddress,
		},
	}

	if config.P2P.PexReactor {
		cellInfo.Channels = append(cellInfo.Channels, pex.PexChannel)
	}

	lAddr := config.P2P.ExternalAddress

	if lAddr == "" {
		lAddr = config.P2P.ListenAddress
	}

	cellInfo.ListenAddr = lAddr

	err := cellInfo.Validate()
	return cellInfo, err
}

//------------------------------------------------------------------------------

var (
	genesisDocKey = []byte("genesisDoc")
)

// panics if failed to unmarshal bytes
func loadGenesisDoc(db dbm.DB) (*types.GenesisDoc, error) {
	bytes := db.Get(genesisDocKey)
	if len(bytes) == 0 {
		return nil, errors.New("Genesis doc not found")
	}
	var genDoc *types.GenesisDoc
	err := cdc.UnmarshalJSON(bytes, &genDoc)
	if err != nil {
		cmn.PanicCrisis(fmt.Sprintf("Failed to load genesis doc due to unmarshaling error: %v (bytes: %X)", err, bytes))
	}
	return genDoc, nil
}

// panics if failed to marshal the given genesis document
func saveGenesisDoc(db dbm.DB, genDoc *types.GenesisDoc) {
	bytes, err := cdc.MarshalJSON(genDoc)
	if err != nil {
		cmn.PanicCrisis(fmt.Sprintf("Failed to save genesis doc due to marshaling error: %v", err))
	}
	db.SetSync(genesisDocKey, bytes)
}

func createAndStartValidatorSocketClient(
	listenAddr string,
	logger log.Logger,
) (types.Validator, error) {
	var listener net.Listener

	protocol, address := cmn.ProtocolAndAddress(listenAddr)
	ln, err := net.Listen(protocol, address)
	if err != nil {
		return nil, err
	}
	switch protocol {
	case "unix":
		listener = validator.NewUnixListener(ln)
	case "tcp":
		// TODO: persist this key so external signer
		// can actually authenticate us
		listener = validator.NewTCPListener(ln, ed25519.GenPrivKey())
	default:
		return nil, fmt.Errorf(
			"Wrong listen address: expected either 'tcp' or 'unix' protocols, got %s",
			protocol,
		)
	}

	pvsc := validator.NewSignerValidatorEndpoint(logger.With("module", "validator"), listener)
	if err := pvsc.Start(); err != nil {
		return nil, errors.Wrap(err, "failed to start private validator")
	}

	return pvsc, nil
}

// splitAndTrimEmpty slices s into all subslices separated by sep and returns a
// slice of the string s with all leading and trailing Unicode code points
// contained in cutset removed. If sep is empty, SplitAndTrim splits after each
// UTF-8 sequence. First part is equivalent to strings.SplitN with a count of
// -1.  also filter out empty strings, only return non-empty strings.
func splitAndTrimEmpty(s, sep, cutset string) []string {
	if s == "" {
		return []string{}
	}

	spl := strings.Split(s, sep)
	nonEmptyStrings := make([]string, 0, len(spl))
	for i := 0; i < len(spl); i++ {
		element := strings.Trim(spl[i], cutset)
		if element != "" {
			nonEmptyStrings = append(nonEmptyStrings, element)
		}
	}
	return nonEmptyStrings
}
