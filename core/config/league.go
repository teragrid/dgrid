package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// RWRandomDropMode is a mode in which we randomly drop reads/writes, connections or sleep
	RWRandomDropMode = iota
	// SleepRandomMode is a mode in which we randomly sleep
	SleepRandomMode

	// LogFormatPlain is a format for colored text
	LogFormatPlain = "plain"
	// LogFormatJSON is a format for json output
	LogFormatJSON = "json"
)

const (
	// BaseLeagueType for initializing the default config for the Base
	BaseLeagueType = iota
	// RegularLeagueType for initializing the default for a league
	RegularLeagueType
)

// NOTE: Most of the structs & relevant comments + the
// default configuration options were used to manually
// generate the config.toml. Please reflect any changes
// made here in the defaultConfigTemplate constant in
// config/toml.go
// NOTE: pkg/cli must know to look in the config dir!
var (
	defaultConfigDir = "config"
	defaultDataDir   = "data"

	defaultConfigFile  = "config.toml"
	defaultGenesisFile = "genesis.json"

	// Set the default consensus protocol for the Base league
	defaultConsensu4BaseLeague = BFTConsensusProtocol

	// Set the default consensus protocol for the Reg league
	defaultConsensu4RegLeague = FBAConsensusProtocol

	// Each cell (logical node will be provided a unique private key)
	defaultCellKey     = "nodekey.json"
	defaultAddressBook = "addressbook.json"
)

// LeagueType offers Base or Regular
type LeagueType int

// DefaultLeagueConfig returns the default config of a resource config
func DefaultLeagueConfig(config Config) *Config {
	return config.Default()
}

// NewLeagueConfig returns a league config details specified by the league type
func NewLeagueConfig(leagueType LeagueType) (*Config, err) {
	switch leagueType {
	case BaseLeagueType:
		return DefaultLeagueConfig(BaseLeagueConfig{}), nil
	case RegularLeagueType:
		return DefaultLeagueConfig(RegularLeagueConfig{}), nil
	default:
		err = fmt.Errorf("Unknown consensus protocol %s", protocol)
		return nil, err
	}
	return nil, nil
}

// configLeagues is a local map of leagueID->leagueConfig
var configLeagues = struct {
	sync.RWMutex
	list map[string]*Config
}{list: make(map[string]*Config)}

// GetLeagueConfig returns the channel configuration of the league with league ID.
// Note that this call returns nil if league id has not been created.
func GetLeagueConfig(leagueID string) Config {
	configLeagues.RLock()
	defer configLeagues.RUnlock()
	if league, ok := configLeagues.list[leagueID]; ok {
		return league
	}
	return nil
}

// StoreLeagueConfig saves a new league configuration to the list
// for later reference
func StoreLeagueConfig(leagueID, config Config) bool {
	leagueConfig = GetLeagueConfig(leagueID)
	if leagueConfig == nil {
		configLeagues.Lock()
		defer configLeagues.Unlock()
		configLeagues[leagueID] = config
		return true
	}
	return false
}

// UpdateLeagueConfig updates a existing league configuration,
// return fail if it doesn't exist
func UpdateLeagueConfig(leagueID, config Config) bool {
	leagueConfig = GetLeagueConfig(leagueID)
	configLeagues.Lock()
	defer configLeagues.Unlock()
	if leagueConfig != nil {
		configLeagues[leagueID] = config
		return true
	}
	return false
}

// DefaultLogLevel returns a default log level of "error"
func DefaultLogLevel() string {
	return "error"
}

// DefaultPackageLogLevels returns a default log level setting so all packages
// log at "error", while the `state` and `main` packages log at "info"
func DefaultPackageLogLevels() string {
	return fmt.Sprintf("main:info,state:info,*:%s", DefaultLogLevel())
}

// -----------------------------------------------------------------------------
// BaseLeagueConfig
// -----------------------------------------------------------------------------

// BaseLeagueConfig defines the configuration of the Base
// for a Dgrid node
type BaseLeagueConfig struct {
	// Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`

	// Options for services
	RPC             *RPCConfig             `mapstructure:"rpc"`
	P2P             *P2PConfig             `mapstructure:"p2p"`
	LeagueStorage   *LeagueStorageConfig   `mapstructure:"league_storage"`
	Consensus       *FBAConsensusConfig    `mapstructure:"fba_consensus"`
	TxIndex         *TxIndexConfig         `mapstructure:"tx_index"`
	Instrumentation *InstrumentationConfig `mapstructure:"instrumentation"`
}

// Default returns a default configuration of the Base for a Dgrid node
func (cfg *BaseLeagueConfig) Default() *Config {
	return &BaseLeagueConfig{
		BaseConfig:      DefaultConfig(BaseConfig{ProxyApp: "tcp://127.0.0.1:26655"}),
		RPC:             DefaultConfig(RPCConfig{}),
		P2P:             DefaultConfig(P2PConfig{}),
		LeagueStorage:   DefaultConfig(LeagueStorage{}),
		Consensus:       NewConsensusConfig(FBAConsensusProtocol, defaultBaseLeagueConfigDir),
		TxIndex:         DefaultConfig(TxIndexConfig{}),
		Instrumentation: DefaultConfig(InstrumentationConfig{}),
	}
}

// Validate performs basic validation (checking param bounds, etc.)
// on a given league and returns an error if any check fails.
func (cfg *BaseLeagueConfig) Validate() error {
	if err := cfg.BaseConfig.Validate(); err != nil {
		return err
	}
	if err := cfg.RPC.Validate(); err != nil {
		return errors.Wrap(err, "Error in [rpc] section")
	}
	if err := cfg.P2P.Validate(); err != nil {
		return errors.Wrap(err, "Error in [p2p] section")
	}
	if err := cfg.LeagueStorage.Validate(); err != nil {
		return errors.Wrap(err, "Error in [league storage] section")
	}
	if err := cfg.Consensus.Validate(); err != nil {
		return errors.Wrap(err, "Error in [consensus] section")
	}
	return errors.Wrap(
		cfg.Instrumentation.Validate(),
		"Error in [instrumentation] section",
	)
}

// SetRoot sets the RootDir for all sub config structs in the Base
func (cfg *BaseLeagueConfig) SetRoot(root string) *Config {
	cfg.BaseConfig.RootDir = root
	cfg.RPC.RootDir = root
	cfg.P2P.RootDir = root
	cfg.LeagueStorage.RootDir = root
	cfg.Consensus.RootDir = root
	return cfg
}

// -----------------------------------------------------------------------------
// RegularLeagueConfig
// -----------------------------------------------------------------------------

// RegularLeagueConfig defines the configuration of a regular league
// for a Dgrid node
type RegularLeagueConfig struct {
	// Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`

	// Options for services
	RPC             *RPCConfig             `mapstructure:"rpc"`
	P2P             *P2PConfig             `mapstructure:"p2p"`
	LeagueStorage   *LeagueStorageConfig   `mapstructure:"league_storage"`
	Consensus       *BFTConsensusConfig    `mapstructure:"bft_consensus"`
	TxIndex         *TxIndexConfig         `mapstructure:"tx_index"`
	Instrumentation *InstrumentationConfig `mapstructure:"instrumentation"`
}

// Default returns the default configuration of a regular league for a Dgrid node
func (cfg *RegularLeagueConfig) Default() *Config {
	return &RegularLeagueConfig{
		BaseConfig:      DefaultConfig(BaseConfig{ProxyApp: "tcp://127.0.0.1:26658"}),
		RPC:             DefaultConfig(RPCConfig{}),
		P2P:             DefaultConfig(P2PConfig{}),
		LeagueStorage:   DefaultConfig(LeagueStorage{}),
		Consensus:       NewConsensusConfig(BFTConsensusProtocol, defaultRegLeagueConfigDir),
		TxIndex:         DefaultConfig(TxIndexConfig{}),
		Instrumentation: DefaultConfig(InstrumentationConfig{}),
	}
}

// Validate performs basic validation (checking param bounds, etc.)
// on a given league and returns an error if any check fails.
func (cfg *RegularLeagueConfig) Validate() error {
	if err := cfg.BaseConfig.Validate(); err != nil {
		return err
	}
	if err := cfg.RPC.Validate(); err != nil {
		return errors.Wrap(err, "Error in [rpc] section")
	}
	if err := cfg.P2P.Validate(); err != nil {
		return errors.Wrap(err, "Error in [p2p] section")
	}
	if err := cfg.LeagueStorage.Validate(); err != nil {
		return errors.Wrap(err, "Error in [league storage] section")
	}
	if err := cfg.Consensus.Validate(); err != nil {
		return errors.Wrap(err, "Error in [consensus] section")
	}
	return errors.Wrap(
		cfg.Instrumentation.Validate(),
		"Error in [instrumentation] section",
	)
}

// SetRoot sets the RootDir for all sub config structs in a reg league config
func (cfg *RegularLeagueConfig) SetRoot(root string) *Config {
	cfg.BaseConfig.RootDir = root
	cfg.RPC.RootDir = root
	cfg.P2P.RootDir = root
	cfg.LeagueStorage.RootDir = root
	cfg.Consensus.RootDir = root
	return cfg
}

//-----------------------------------------------------------------------------
// BaseConfig
//-----------------------------------------------------------------------------

// BaseConfig defines the base configuration for a Dgrid node
type BaseConfig struct {
	// leagueID is unexposed and immutable but here for convenience
	leagueID string

	// The root directory for all data.
	// This should be set in viper so it can unmarshal into this struct
	RootDir string `mapstructure:"home"`

	// The home directory for configuration of a league
	ConfigDir string `mapstructure:config_dir`

	// TCP or UNIX socket address of the Asura application,
	// or the name of an Asura application compiled in with the Dgrid binary
	ProxyApp string `mapstructure:"proxy_app"`

	// A custom human readable name for this node
	Hostname string `mapstructure:"hostname"`

	// If this node is many blocks behind the tip of the chain, FastSync
	// allows them to catchup quickly by downloading blocks in parallel
	// and verifying their commits
	FastSync bool `mapstructure:"fast_sync"`

	// Database backend: leveldb | memdb | cleveldb
	DBBackend string `mapstructure:"db_backend"`

	// Database directory
	DBPath string `mapstructure:"db_dir"`

	// Output level for logging
	LogLevel string `mapstructure:"log_level"`

	// Output format: 'plain' (colored text) or 'json'
	LogFormat string `mapstructure:"log_format"`

	// Path to the JSON file containing the initial validator set and other meta data
	Genesis string `mapstructure:"genesis_file"`

	// TCP or UNIX socket address for Dgrid to listen on for
	// connections from an external Validator process
	ValidatorListenAddr string `mapstructure:"validator_laddr"`

	// A JSON file containing the private key to use for p2p authenticated encryption
	CellKey string `mapstructure:"node_key_file"`

	// Mechanism to connect to the Asura application: socket | grpc
	Asura string `mapstructure:"asura"`

	// TCP or UNIX socket address for the profiling server to listen on
	ProfListenAddress string `mapstructure:"prof_laddr"`

	// If true, query the Asura app on connecting to a new peer
	// so the app can decide if we should keep the connection or not
	FilterPeers bool `mapstructure:"filter_peers"` // false
}

// Default returns a default base configuration for a Dgrid node
func (cfg *BaseConfig) Default() *BaseConfig {
	return &BaseConfig{
		ConfigDir:         defaultConfigDir,
		Genesis:           filepath.Join(defaultConfigDir, defaultGenesisFile),
		CellKey:           filepath.Join(defaultConfigDir, defaultCellKey),
		Hostname:          defaultHostname,
		ProxyApp:          cfg.ProxyApp,
		Asura:             "socket",
		LogLevel:          DefaultPackageLogLevels(),
		LogFormat:         LogFormatPlain,
		ProfListenAddress: "",
		FastSync:          true,
		FilterPeers:       false,
		DBBackend:         "leveldb",
		DBPath:            "data",
	}
}

// Validate performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *BaseConfig) Validate() error {
	switch cfg.LogFormat {
	case LogFormatPlain, LogFormatJSON:
	default:
		return errors.New("unknown log_format (must be 'plain' or 'json')")
	}
	return nil
}

// LeagueID returns the id of a league
func (cfg BaseConfig) LeagueID() string {
	return cfg.leagueID
}

// GenesisFile returns the full path to the genesis.json file
func (cfg BaseConfig) GenesisFile() string {
	return Rootify(cfg.Genesis, cfg.RootDir)
}

// CellKeyFile returns the full path to the node_key.json file
func (cfg BaseConfig) CellKeyFile() string {
	return Rootify(cfg.CellKey, cfg.RootDir)
}

// DBDir returns the full path to the database directory
func (cfg BaseConfig) DBDir() string {
	return Rootify(cfg.DBPath, cfg.RootDir)
}

// FuzzConnConfig is a FuzzedConnection configuration.
type FuzzConnConfig struct {
	Mode         int
	MaxDelay     time.Duration
	ProbDropRW   float64
	ProbDropConn float64
	ProbSleep    float64
}

// DefaultFuzzConnConfig returns the default config.
func DefaultFuzzConnConfig() *FuzzConnConfig {
	return &FuzzConnConfig{
		Mode:         FuzzModeDrop,
		MaxDelay:     3 * time.Second,
		ProbDropRW:   0.2,
		ProbDropConn: 0.00,
		ProbSleep:    0.00,
	}
}

//-----------------------------------------------------------------------------
// Hostname

var defaultHostname = getDefaultHostname()

// getDefaultHostname() returns a default hostname. If runtime
// fails to get the host name, "anonymous" will be returned.
func getDefaultHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "anonymous"
	}
	return hostname
}
