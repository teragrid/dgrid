package config

import "errors"

// LeagueStorageConfig

// LeagueStorageConfig defines the configuration options for Dgrid League Storage
type LeagueStorageConfig struct {
	RootDir     string `mapstructure:"home"`
	Recheck     bool   `mapstructure:"recheck"`
	Broadcast   bool   `mapstructure:"broadcast"`
	WalPath     string `mapstructure:"wal_dir"`
	Size        int    `mapstructure:"size"`
	MaxTxsBytes int64  `mapstructure:"max_txs_bytes"`
	CacheSize   int    `mapstructure:"cache_size"`
}

// Default returns a default configuration for Dgrid League Storage
func (cfg *LeagueStorageConfig) Default() *Config {
	return &LeagueStorageConfig{
		Recheck:   true,
		Broadcast: true,
		WalPath:   "",
		// Each signature verification takes .5ms, Size reduced until we implement
		// Asura Recheck
		Size:        5000,
		MaxTxsBytes: 1024 * 1024 * 1024, // 1GB
		CacheSize:   10000,
	}
}

// Validate performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *LeagueStorageConfig) Validate() error {
	if cfg.Size < 0 {
		return errors.New("size can't be negative")
	}
	if cfg.MaxTxsBytes < 0 {
		return errors.New("max_txs_bytes can't be negative")
	}
	if cfg.CacheSize < 0 {
		return errors.New("cache_size can't be negative")
	}
	return nil
}

// WalDir returns the full path to the League Storage's write-ahead log
func (cfg *LeagueStorageConfig) WalDir() string {
	return Rootify(cfg.WalPath, cfg.RootDir)
}

// WalEnabled returns true if the WAL is enabled.
func (cfg *LeagueStorageConfig) WalEnabled() bool {
	return cfg.WalPath != ""
}
