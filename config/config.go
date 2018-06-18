package config

import (
	//	"fmt"
	//	"os"
	"path/filepath"
	//	"time"
)

// NOTE: Most of the structs & relevant comments + the
// default configuration options were used to manually
// generate the config.toml. Please reflect any changes
// made here in the defaultConfigTemplate constant in
// config/toml.go
// NOTE: teralibs/cli must know to look in the config dir!
var (
	DefaultteragridDir = ".teragrid"
	defaultChainName   = "default"
	defaultConfigDir   = "config"
	defaultDataDir     = "data"

	defaultConfigFileName  = "config.toml"
	defaultGenesisJSONName = "genesis.json"

	defaultPrivValName  = "priv_validator.json"
	defaultNodeKeyName  = "node_key.json"
	defaultAddrBookName = "addrbook.json"

	defaultConfigFilePath  = filepath.Join(defaultChainName, defaultConfigDir, defaultConfigFileName)
	defaultGenesisJSONPath = filepath.Join(defaultChainName, defaultConfigDir, defaultGenesisJSONName)
	defaultPrivValPath     = filepath.Join(defaultChainName, defaultConfigDir, defaultPrivValName)
	defaultNodeKeyPath     = filepath.Join(defaultChainName, defaultConfigDir, defaultNodeKeyName)
	defaultAddrBookPath    = filepath.Join(defaultChainName, defaultConfigDir, defaultAddrBookName)
)

type Config struct {
	// The root directory for all data.
	// This should be set in viper so it can unmarshal into this struct
	RootDir string `mapstructure:"home"`
	// Output level for logging
	LogLevel     string `mapstructure:"log_level"`
	ChainConfigs []ChainConfig
}

// SetRoot sets the RootDir for all Config structs
func (cfg *Config) SetRoot(root string) *Config {
	cfg.RootDir = root
	for _, chain := range cfg.ChainConfigs {
		chainDir := root + chain.ChainID()
		chain.BaseConfig.RootDir = chainDir
		chain.RPC.RootDir = chainDir
		chain.P2P.RootDir = chainDir
		chain.Mempool.RootDir = chainDir
		chain.Consensus.RootDir = chainDir
	}
	return cfg
}

// DefaultConfig returns a default configuration for a teragrid node
func DefaultConfig() *Config {
	cfg := Config{
		RootDir:  "",
		LogLevel: DefaultPackageLogLevels(),
	}
	cfg.ChainConfigs = make([]ChainConfig, 1)
	var chainConfig ChainConfig
	chainConfig = *DefaultChainConfig()
	cfg.ChainConfigs[0] = chainConfig
	return &cfg
}

// TestConfig returns a configuration that can be used for testing
func TestConfig() *Config {
	return DefaultConfig()
	//	return &Config{
	//		DefaultConfig(),
	//	}
}
