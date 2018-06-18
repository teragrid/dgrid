package config

//-----------------------------------------------------------------------------
// ShardConfig

// ShardConfig defines the base configuration for a teragrid quorum
type ShardConfig struct {
	ShardID   string
	Validator bool

	// Path to the JSON file containing the private key to use as a validator in the consensus protocol
	PrivValidator string `mapstructure:"priv_validator_file"`
}

// Config defines the top level configuration for a teragrid node
type ChainConfig struct {
	// Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`
	Shard      []ShardConfig
	// Options for services
	RPC       *RPCConfig       `mapstructure:"rpc"`
	P2P       *P2PConfig       `mapstructure:"p2p"`
	Mempool   *MempoolConfig   `mapstructure:"mempool"`
	Consensus *ConsensusConfig `mapstructure:"consensus"`
	TxIndex   *TxIndexConfig   `mapstructure:"tx_index"`
}

// DefaultConfig returns a default configuration for a teragrid node
func DefaultChainConfig() *ChainConfig {
	return &ChainConfig{
		BaseConfig: DefaultBaseConfig(),
		RPC:        DefaultRPCConfig(),
		P2P:        DefaultP2PConfig(),
		Mempool:    DefaultMempoolConfig(),
		Consensus:  DefaultConsensusConfig(),
		TxIndex:    DefaultTxIndexConfig(),
	}
}

// TestConfig returns a configuration that can be used for testing
func TestChainConfig() *ChainConfig {
	return &ChainConfig{
		BaseConfig: TestBaseConfig(),
		RPC:        TestRPCConfig(),
		P2P:        TestP2PConfig(),
		Mempool:    TestMempoolConfig(),
		Consensus:  TestConsensusConfig(),
		TxIndex:    TestTxIndexConfig(),
	}
}

// SetRoot sets the RootDir for all Config structs
func (cfg *ChainConfig) SetRoot(root string) *ChainConfig {
	cfg.BaseConfig.RootDir = root
	cfg.RPC.RootDir = root
	cfg.P2P.RootDir = root
	cfg.Mempool.RootDir = root
	cfg.Consensus.RootDir = root
	return cfg
}
