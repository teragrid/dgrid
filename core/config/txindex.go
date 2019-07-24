package config

// TxIndexConfig

// TxIndexConfig defines the configuration for the transaction indexer,
// including tags to index.
type TxIndexConfig struct {
	// What indexer to use for transactions
	//
	// Options:
	//   1) "null"
	//   2) "kv" (default) - the simplest possible indexer, backed by key-value storage (defaults to levelDB; see DBBackend).
	Indexer string `mapstructure:"indexer"`

	// Comma-separated list of tags to index (by default the only tag is "tx.hash")
	//
	// You can also index transactions by height by adding "tx.height" tag here.
	//
	// It's recommended to index only a subset of tags due to possible memory
	// bloat. This is, of course, depends on the indexer's DB and the volume of
	// transactions.
	IndexTags string `mapstructure:"index_tags"`

	// When set to true, tells indexer to index all tags (predefined tags:
	// "tx.hash", "tx.height" and all tags from DeliverTx responses).
	//
	// Note this may be not desirable (see the comment above). IndexTags has a
	// precedence over IndexAllTags (i.e. when given both, IndexTags will be
	// indexed).
	IndexAllTags bool `mapstructure:"index_all_tags"`
}

// DefaultTxIndexConfig returns a default configuration for the transaction indexer.
func DefaultTxIndexConfig() *TxIndexConfig {
	return &TxIndexConfig{
		Indexer:      "kv",
		IndexTags:    "",
		IndexAllTags: false,
	}
}
