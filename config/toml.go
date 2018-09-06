package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/viper"
	cmn "github.com/teragrid/teralibs/common"
)

var configTemplate *template.Template

func init() {
	var err error
	if configTemplate, err = template.New("configFileTemplate").Parse(defaultConfigTemplate); err != nil {
		fmt.Println("Error configTemplate")
		panic(err)
	}
}

/****** these are for production settings ***********/

// EnsureRoot creates the root, config, and data directories if they don't exist,
// and panics if it fails.
func EnsureRoot(rootDir string, config *Config) {
	if err := cmn.EnsureDir(rootDir, 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}
	for _, chain := range config.ChainConfigs {
		chainDir := chain.ChainID()
		if err := cmn.EnsureDir(filepath.Join(rootDir, chainDir), 0700); err != nil {
			cmn.PanicSanity(err.Error())
		}
		if err := cmn.EnsureDir(filepath.Join(rootDir, chainDir, defaultConfigDir), 0700); err != nil {
			cmn.PanicSanity(err.Error())
		}
		if err := cmn.EnsureDir(filepath.Join(rootDir, chainDir, defaultDataDir), 0700); err != nil {
			cmn.PanicSanity(err.Error())
		}
	}

	//	configFilePath := rootDir //filepath.Join(rootDir, defaultChainName, defaultConfigFilePath)
	//	fmt.Println("EnsureRoot  " + filepath.Join(configFilePath, "config.json"))
	// Write default config file if missing.
	//	if !cmn.FileExists(filepath.Join(configFilePath, "config.json")) {
	//		fmt.Println("EnsureRoot (WRITE) " + filepath.Join(configFilePath, "config.json"))
	//		writeDefaultConfigFile(configFilePath, config)
	//	}
	WriteConfigFile(rootDir, config)
}

// XXX: this func should probably be called by cmd/teragrid/commands/init.go
// alongside the writing of the genesis.json and priv_validator.json
func writeDefaultConfigFile(configFilePath string, config *Config) {
	WriteConfigFile(configFilePath, config)
}

// WriteConfigFile renders config using the template and writes it to configFilePath.
func WriteConfigFile(configFilePath string, config *Config) {
	var chains []string
	chains = make([]string, len(config.ChainConfigs))
	for idx, chain := range config.ChainConfigs {
		chains[idx] = chain.ChainID()

		var buffer bytes.Buffer

		if err := configTemplate.Execute(&buffer, chain); err != nil {
			fmt.Println("WriteConfigFile Panic " + configFilePath)
			panic(err)
		} else {
			if !cmn.FileExists(filepath.Join(configFilePath, chain.ChainID(), defaultConfigFilePath)) {
				cmn.MustWriteFile(filepath.Join(configFilePath, chain.ChainID(), defaultConfigFilePath), buffer.Bytes(), 0644)
			}
		}
	}
	if true || !cmn.FileExists(filepath.Join(configFilePath, "config.json")) {
		var runtime_viper = viper.New()
		runtime_viper.SetConfigType("json")
		runtime_viper.SetConfigFile(filepath.Join(configFilePath, "config.json"))
		runtime_viper.SetDefault("LogLevel", config.LogLevel)
		runtime_viper.SetDefault("Chains", chains)
		runtime_viper.WriteConfig()
	}
}

// Note: any changes to the comments/variables/mapstructure
// must be reflected in the appropriate struct in config/config.go
const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# TCP or UNIX socket address of the Asura application,
# or the name of an Asura application compiled in with the Teragrid binary
proxy_app = "{{ .BaseConfig.ProxyApp }}"

# A custom human readable name for this node
moniker = "{{ .BaseConfig.Moniker }}"

# If this node is many blocks behind the tip of the chain, FastSync
# allows them to catchup quickly by downloading blocks in parallel
# and verifying their commits
fast_sync = {{ .BaseConfig.FastSync }}

# Database backend: leveldb | memdb
db_backend = "{{ .BaseConfig.DBBackend }}"

# Database directory
db_path = "{{ js .BaseConfig.DBPath }}"

# Output level for logging, including package level options
log_level = "{{ .BaseConfig.LogLevel }}"

##### additional base config options #####

# Path to the JSON file containing the initial validator set and other meta data
genesis_file = "{{ js .BaseConfig.Genesis }}"

# Path to the JSON file containing the private key to use as a validator in the consensus protocol
priv_validator_file = "{{ js .BaseConfig.PrivValidator }}"

# Path to the JSON file containing the private key to use for node authentication in the p2p protocol
node_key_file = "{{ js .BaseConfig.NodeKey}}"

# Mechanism to connect to the Asura application: socket | grpc
asura = "{{ .BaseConfig.Asura }}"

# TCP or UNIX socket address for the profiling server to listen on
prof_laddr = "{{ .BaseConfig.ProfListenAddress }}"

# If true, query the Asura app on connecting to a new peer
# so the app can decide if we should keep the connection or not
filter_peers = {{ .BaseConfig.FilterPeers }}

##### advanced configuration options #####

##### rpc server configuration options #####
[rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "{{ .RPC.ListenAddress }}"

# TCP or UNIX socket address for the gRPC server to listen on
# NOTE: This server only supports /broadcast_tx_commit
grpc_laddr = "{{ .RPC.GRPCListenAddress }}"

# Activate unsafe RPC commands like /dial_seeds and /unsafe_flush_mempool
unsafe = {{ .RPC.Unsafe }}

##### peer to peer configuration options #####
[p2p]

# Address to listen for incoming connections
laddr = "{{ .P2P.ListenAddress }}"

# Comma separated list of seed nodes to connect to
seeds = "{{ .P2P.Seeds }}"

# Comma separated list of nodes to keep persistent connections to
# Do not add private peers to this list if you don't want them advertised
persistent_peers = "{{ .P2P.PersistentPeers }}"

# Path to address book
addr_book_file = "{{ js .P2P.AddrBook }}"

# Set true for strict address routability rules
addr_book_strict = {{ .P2P.AddrBookStrict }}

# Time to wait before flushing messages out on the connection, in ms
flush_throttle_timeout = {{ .P2P.FlushThrottleTimeout }}

# Maximum number of peers to connect to
max_num_peers = {{ .P2P.MaxNumPeers }}

# Maximum size of a message packet payload, in bytes
max_packet_msg_payload_size = {{ .P2P.MaxPacketMsgPayloadSize }}

# Rate at which packets can be sent, in bytes/second
send_rate = {{ .P2P.SendRate }}

# Rate at which packets can be received, in bytes/second
recv_rate = {{ .P2P.RecvRate }}

# Set true to enable the peer-exchange reactor
pex = {{ .P2P.PexReactor }}

# Seed mode, in which node constantly crawls the network and looks for
# peers. If another node asks it for addresses, it responds and disconnects.
#
# Does not work if the peer-exchange reactor is disabled.
seed_mode = {{ .P2P.SeedMode }}

# Comma separated list of peer IDs to keep private (will not be gossiped to other peers)
private_peer_ids = "{{ .P2P.PrivatePeerIDs }}"

##### mempool configuration options #####
[mempool]

recheck = {{ .Mempool.Recheck }}
recheck_empty = {{ .Mempool.RecheckEmpty }}
broadcast = {{ .Mempool.Broadcast }}
wal_dir = "{{ js .Mempool.WalPath }}"

# size of the mempool
size = {{ .Mempool.Size }}

# size of the cache (used to filter transactions we saw earlier)
cache_size = {{ .Mempool.CacheSize }}

##### consensus configuration options #####
[consensus]

wal_file = "{{ js .Consensus.WalPath }}"

# All timeouts are in milliseconds
timeout_propose = {{ .Consensus.TimeoutPropose }}
timeout_propose_delta = {{ .Consensus.TimeoutProposeDelta }}
timeout_prevote = {{ .Consensus.TimeoutPrevote }}
timeout_prevote_delta = {{ .Consensus.TimeoutPrevoteDelta }}
timeout_precommit = {{ .Consensus.TimeoutPrecommit }}
timeout_precommit_delta = {{ .Consensus.TimeoutPrecommitDelta }}
timeout_commit = {{ .Consensus.TimeoutCommit }}

# Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)
skip_timeout_commit = {{ .Consensus.SkipTimeoutCommit }}

# BlockSize
max_block_size_txs = {{ .Consensus.MaxBlockSizeTxs }}
max_block_size_bytes = {{ .Consensus.MaxBlockSizeBytes }}

# EmptyBlocks mode and possible interval between empty blocks in seconds
create_empty_blocks = {{ .Consensus.CreateEmptyBlocks }}
create_empty_blocks_interval = {{ .Consensus.CreateEmptyBlocksInterval }}

# Reactor sleep duration parameters are in milliseconds
peer_gossip_sleep_duration = {{ .Consensus.PeerGossipSleepDuration }}
peer_query_maj23_sleep_duration = {{ .Consensus.PeerQueryMaj23SleepDuration }}

##### transactions indexer configuration options #####
[tx_index]

# What indexer to use for transactions
#
# Options:
#   1) "null" (default)
#   2) "kv" - the simplest possible indexer, backed by key-value storage (defaults to levelDB; see DBBackend).
indexer = "{{ .TxIndex.Indexer }}"

# Comma-separated list of tags to index (by default the only tag is tx hash)
#
# It's recommended to index only a subset of tags due to possible memory
# bloat. This is, of course, depends on the indexer's DB and the volume of
# transactions.
index_tags = "{{ .TxIndex.IndexTags }}"

# When set to true, tells indexer to index all tags. Note this may be not
# desirable (see the comment above). IndexTags has a precedence over
# IndexAllTags (i.e. when given both, IndexTags will be indexed).
index_all_tags = {{ .TxIndex.IndexAllTags }}

`

/****** these are for test settings ***********/

func ResetTestRoot(testName string) *Config {
	rootDir := os.ExpandEnv("$HOME/.teragrid_test")
	rootDir = filepath.Join(rootDir, testName)
	// Remove ~/.teragrid_test_bak
	if cmn.FileExists(rootDir + "_bak") {
		if err := os.RemoveAll(rootDir + "_bak"); err != nil {
			cmn.PanicSanity(err.Error())
		}
	}
	// Move ~/.teragrid_test to ~/.teragrid_test_bak
	if cmn.FileExists(rootDir) {
		if err := os.Rename(rootDir, rootDir+"_bak"); err != nil {
			cmn.PanicSanity(err.Error())
		}
	}
	// Create new dir
	if err := cmn.EnsureDir(rootDir, 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}
	if err := cmn.EnsureDir(filepath.Join(rootDir, defaultChainName), 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}
	if err := cmn.EnsureDir(filepath.Join(rootDir, defaultChainName, defaultConfigDir), 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}
	if err := cmn.EnsureDir(filepath.Join(rootDir, defaultChainName, defaultDataDir), 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}

	config := DefaultConfig()
	//baseConfig := DefaultBaseConfig(defaultChainName)
	baseConfig := config.ChainConfigs[0]
	configFilePath := filepath.Join(rootDir, defaultChainName, defaultConfigFilePath)
	genesisFilePath := filepath.Join(rootDir, defaultChainName, baseConfig.Genesis)
	privFilePath := filepath.Join(rootDir, defaultChainName, baseConfig.PrivValidator)

	// Write default config file if missing.
	if !cmn.FileExists(configFilePath) {
		fmt.Println("EnsureDir_writeDefaultCondigFile XXXX " + configFilePath)
		writeDefaultConfigFile(rootDir, config)
	}
	if !cmn.FileExists(genesisFilePath) {
		fmt.Println("ResetTestRoot genesisFilePath XXXX " + genesisFilePath)
		cmn.MustWriteFile(genesisFilePath, []byte(testGenesis), 0644)
	}
	fmt.Println("ResetTestRoot privFilePath XXXX " + privFilePath)
	// we always overwrite the priv val
	cmn.MustWriteFile(privFilePath, []byte(testPrivValidator), 0644)

	configX := TestConfig().SetRoot(rootDir)
	return configX
}

var testGenesis = `{
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "chain_id": "teragrid_test",
  "validators": [
    {
      "pub_key": {
        "type": "AC26791624DE60",
        "value":"AT/+aaL1eB0477Mud9JMm8Sh8BIvOYlPGC9KkIUmFaE="
      },
      "power": 10,
      "name": ""
    }
  ],
  "app_hash": ""
}`

var testPrivValidator = `{
  "address": "849CB2C877F87A20925F35D00AE6688342D25B47",
  "pub_key": {
    "type": "AC26791624DE60",
    "value": "AT/+aaL1eB0477Mud9JMm8Sh8BIvOYlPGC9KkIUmFaE="
  },
  "priv_key": {
    "type": "954568A3288910",
    "value": "EVkqJO/jIXp3rkASXfh9YnyToYXRXhBr6g9cQVxPFnQBP/5povV4HTjvsy530kybxKHwEi85iU8YL0qQhSYVoQ=="
  },
  "last_height": 0,
  "last_round": 0,
  "last_step": 0
}`