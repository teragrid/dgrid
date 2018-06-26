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

	defaultConfigFilePath  = filepath.Join(defaultConfigDir, defaultConfigFileName)
	defaultGenesisJSONPath = filepath.Join(defaultConfigDir, defaultGenesisJSONName)
	defaultPrivValPath     = filepath.Join(defaultConfigDir, defaultPrivValName)
	defaultNodeKeyPath     = filepath.Join(defaultConfigDir, defaultNodeKeyName)
	defaultAddrBookPath    = filepath.Join(defaultConfigDir, defaultAddrBookName)
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
		chain.SetRoot(chainDir)
	}
	return cfg
}

// DefaultConfig returns a default configuration for a teragrid node
func DefaultConfig() *Config {
	cfg := Config{
		RootDir:  "",
		LogLevel: DefaultPackageLogLevels(),
		ChainConfigs: []ChainConfig{
			*DefaultChainConfig(defaultChainName),
		},
	}
	return &cfg
}

// TestConfig returns a configuration that can be used for testing
func TestConfig() *Config {
	return DefaultConfig()
	//return &Config{
	//	DefaultConfig(),
	//}
}
