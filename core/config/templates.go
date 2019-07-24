package config

import "html/template"

// LeagueConfigFileTemplate returns the configutation template of a league
// based on the input league type, including Base or Reg
func LeagueConfigFileTemplate(leagueType int) *template.Template {
	var configTemplate *template.Template
	switch leagueType {
	case BaseLeagueType:
		if configTemplate, err = template.New("leagueConfigTemplate").Parse(
			baseLeagueConfigTemplate); err != nil {
			panic(err)
		}
	case RegLeagueType:
		if configTemplate, err = template.New("leagueConfigTemplate").Parse(
			regLeagueConfigTemplate); err != nil {
			panic(err)
		}
	}
	return configTemplate
}

// BaseLeagueConfigFileTemplate defines the configuration of the Base
// Note: any changes to the comments/variables/mapstructure
// must be reflected in the appropriate struct in config/config.go
const baseLeagueConfigFileTemplate = `# This is a TOML config file.
# It contains the configuration template of the Based League
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# ID of the Base League
league_id = "{{ .BaseLeagueConfig.LeagueId}}"

# TCP or UNIX socket address of the Asura application,
# or the name of an Asura application compiled in with the Dgrid binary
proxy_app = "{{ .BaseLeagueConfig.ProxyApp }}"

# A custom human readable name for this node
hostname = "{{ .BaseLeagueConfig.Hostname }}"

# If this node is many blocks behind the tip of the chain, FastSync
# allows them to catchup quickly by downloading blocks in parallel
# and verifying their commits
fast_sync = {{ .BaseLeagueConfig.FastSync }}

# Database backend: leveldb | memdb | cleveldb
db_backend = "{{ .BaseLeagueConfig.DBBackend }}"

# Database directory
db_dir = "{{ js .BaseLeagueConfig.DBPath }}"

# Output level for logging, including package level options
log_level = "{{ .BaseLeagueConfig.LogLevel }}"

# Output format: 'plain' (colored text) or 'json'
log_format = "{{ .BaseLeagueConfig.LogFormat }}"

##### additional base config options #####

# Path to the JSON file containing the initial validator set and other meta data
genesis_file = "{{ js .BaseLeagueConfig.Genesis }}"

# Path to the JSON file containing the private key to use as a validator in the consensus protocol
validator_key_file = "{{ js .BaseLeagueConfig.ValidatorKey }}"

# Path to the JSON file containing the last sign state of a validator
validator_state_file = "{{ js .BaseLeagueConfig.ValidatorState }}"

# TCP or UNIX socket address for Dgrid to listen on for
# connections from an external Validator process
validator_laddr = "{{ .BaseLeagueConfig.ValidatorListenAddr }}"

# Path to the JSON file containing the private key to use for node authentication in the p2p protocol
node_key_file = "{{ js .BaseLeagueConfig.NodeKey }}"

# Mechanism to connect to the Asura application: socket | grpc
asura = "{{ .BaseLeagueConfig.Asura }}"

# TCP or UNIX socket address for the profiling server to listen on
prof_laddr = "{{ .BaseLeagueConfig.ProfListenAddress }}"

# If true, query the Asura app on connecting to a new peer
# so the app can decide if we should keep the connection or not
filter_peers = {{ .BaseLeagueConfig.FilterPeers }}

##### advanced configuration options #####

##### rpc server configuration options #####
[rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "{{ .RPC.ListenAddress }}"

# A list of origins a cross-domain request can be executed from
# Default value '[]' disables cors support
# Use '["*"]' to allow any origin
cors_allowed_origins = [{{ range .RPC.CORSAllowedOrigins }}{{ printf "%q, " . }}{{end}}]

# A list of methods the client is allowed to use with cross-domain requests
cors_allowed_methods = [{{ range .RPC.CORSAllowedMethods }}{{ printf "%q, " . }}{{end}}]

# A list of non simple headers the client is allowed to use with cross-domain requests
cors_allowed_headers = [{{ range .RPC.CORSAllowedHeaders }}{{ printf "%q, " . }}{{end}}]

# TCP or UNIX socket address for the gRPC server to listen on
# NOTE: This server only supports /broadcast_tx_commit
grpc_laddr = "{{ .RPC.GRPCListenAddress }}"

# Maximum number of simultaneous connections.
# Does not include RPC (HTTP&WebSocket) connections. See max_open_connections
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
# Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
# 1024 - 40 - 10 - 50 = 924 = ~900
grpc_max_open_connections = {{ .RPC.GRPCMaxOpenConnections }}

# Activate unsafe RPC commands like /dial_seeds and /unsafe_flush_storage
unsafe = {{ .RPC.Unsafe }}

# Maximum number of simultaneous connections (including WebSocket).
# Does not include gRPC connections. See grpc_max_open_connections
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
# Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
# 1024 - 40 - 10 - 50 = 924 = ~900
max_open_connections = {{ .RPC.MaxOpenConnections }}

# Maximum number of unique clientIDs that can /subscribe
# If you're using /broadcast_tx_commit, set to the estimated maximum number
# of broadcast_tx_commit calls per block.
max_subscription_clients = {{ .RPC.MaxSubscriptionClients }}

# Maximum number of unique queries a given client can /subscribe to
# If you're using GRPC (or Local RPC client) and /broadcast_tx_commit, set to
# the estimated # maximum number of broadcast_tx_commit calls per block.
max_subscriptions_per_client = {{ .RPC.MaxSubscriptionsPerClient }}

# How long to wait for a tx to be committed during /broadcast_tx_commit.
# WARNING: Using a value larger than 10s will result in increasing the
# global HTTP write timeout, which applies to all connections and endpoints.
# See https://github.com/teragrid/dgrid/issues/3435
timeout_broadcast_tx_commit = "{{ .RPC.TimeoutBroadcastTxCommit }}"

# The name of a file containing certificate that is used to create the HTTPS server.
# If the certificate is signed by a certificate authority,
# the certFile should be the concatenation of the server's certificate, any intermediates,
# and the CA's certificate.
# NOTE: both tls_cert_file and tls_key_file must be present for Dgrid to create HTTPS server. Otherwise, HTTP server is run.
tls_cert_file = "{{ .RPC.TLSCertFile }}"

# The name of a file containing matching private key that is used to create the HTTPS server.
# NOTE: both tls_cert_file and tls_key_file must be present for Dgrid to create HTTPS server. Otherwise, HTTP server is run.
tls_key_file = "{{ .RPC.TLSKeyFile }}"

##### peer to peer configuration options #####
[p2p]

# Address to listen for incoming connections
laddr = "{{ .P2P.ListenAddress }}"

# Address to advertise to peers for them to dial
# If empty, will use the same port as the laddr,
# and will introspect on the listener or use UPnP
# to figure out the address.
external_address = "{{ .P2P.ExternalAddress }}"

# Comma separated list of seed nodes to connect to
seeds = "{{ .P2P.Seeds }}"

# Comma separated list of nodes to keep persistent connections to
persistent_peers = "{{ .P2P.PersistentPeers }}"

# UPNP port forwarding
upnp = {{ .P2P.UPNP }}

# Path to address book
addr_book_file = "{{ js .P2P.AddrBook }}"

# Set true for strict address routability rules
# Set false for private or local networks
addr_book_strict = {{ .P2P.AddrBookStrict }}

# Maximum number of inbound peers
max_num_inbound_peers = {{ .P2P.MaxNumInboundPeers }}

# Maximum number of outbound peers to connect to, excluding persistent peers
max_num_outbound_peers = {{ .P2P.MaxNumOutboundPeers }}

# Time to wait before flushing messages out on the connection
flush_throttle_timeout = "{{ .P2P.FlushThrottleTimeout }}"

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

# Toggle to disable guard against peers connecting from the same ip.
allow_duplicate_ip = {{ .P2P.AllowDuplicateIP }}

# Peer connection configuration.
handshake_timeout = "{{ .P2P.HandshakeTimeout }}"
dial_timeout = "{{ .P2P.DialTimeout }}"

##### league storage configuration options #####
[league_storage]

recheck = {{ .LeagueStorage.Recheck }}
broadcast = {{ .LeagueStorage.Broadcast }}
wal_dir = "{{ js .LeagueStorage.WalPath }}"

# Maximum number of transactions in the storage
size = {{ .LeagueStorage.Size }}

# Limit the total size of all txs in the storage.
# This only accounts for raw transactions (e.g. given 1MB transactions and
# max_txs_bytes=5MB, storage will only accept 5 transactions).
max_txs_bytes = {{ .LeagueStorage.MaxTxsBytes }}

# Size of the cache (used to filter transactions we saw earlier) in transactions
cache_size = {{ .LeagueStorage.CacheSize }}

##### FBA consensus configuration options #####
[fba_consensus]

wal_file = "{{ js .FBAConsensusConfig.WalPath }}"

timeout_propose = "{{ .FBAConsensusConfig.TimeoutPropose }}"
timeout_propose_delta = "{{ .FBAConsensusConfig.TimeoutProposeDelta }}"
timeout_prevote = "{{ .FBAConsensusConfig.TimeoutPrevote }}"
timeout_prevote_delta = "{{ .FBAConsensusConfig.TimeoutPrevoteDelta }}"
timeout_precommit = "{{ .FBAConsensusConfig.TimeoutPrecommit }}"
timeout_precommit_delta = "{{ .FBAConsensusConfig.TimeoutPrecommitDelta }}"
timeout_commit = "{{ .FBAConsensusConfig.TimeoutCommit }}"

# Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)
skip_timeout_commit = {{ .FBAConsensusConfig.SkipTimeoutCommit }}

# EmptyBlocks mode and possible interval between empty blocks
create_empty_blocks = {{ .FBAConsensusConfig.CreateEmptyBlocks }}
create_empty_blocks_interval = "{{ .FBAConsensusConfig.CreateEmptyBlocksInterval }}"

# Reactor sleep duration parameters
peer_gossip_sleep_duration = "{{ .FBAConsensusConfig.PeerGossipSleepDuration }}"
peer_query_maj23_sleep_duration = "{{ .FBAConsensusConfig.PeerQueryMaj23SleepDuration }}"

##### BFT consensus configuration options #####
[bft_consensus]

wal_file = "{{ js .BFTConsensusConfig.WalPath }}"

timeout_propose = "{{ .BFTConsensusConfig.TimeoutPropose }}"
timeout_propose_delta = "{{ .BFTConsensusConfig.TimeoutProposeDelta }}"
timeout_prevote = "{{ .BFTConsensusConfig.TimeoutPrevote }}"
timeout_prevote_delta = "{{ .BFTConsensusConfig.TimeoutPrevoteDelta }}"
timeout_precommit = "{{ .BFTConsensusConfig.TimeoutPrecommit }}"
timeout_precommit_delta = "{{ .BFTConsensusConfig.TimeoutPrecommitDelta }}"
timeout_commit = "{{ .BFTConsensusConfig.TimeoutCommit }}"

# Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)
skip_timeout_commit = {{ .BFTConsensusConfig.SkipTimeoutCommit }}

# EmptyBlocks mode and possible interval between empty blocks
create_empty_blocks = {{ .BFTConsensusConfig.CreateEmptyBlocks }}
create_empty_blocks_interval = "{{ .BFTConsensusConfig.CreateEmptyBlocksInterval }}"

# Reactor sleep duration parameters
peer_gossip_sleep_duration = "{{ .BFTConsensusConfig.PeerGossipSleepDuration }}"
peer_query_maj23_sleep_duration = "{{ .BFTConsensusConfig.PeerQueryMaj23SleepDuration }}"

##### transactions indexer configuration options #####
[tx_index]

# What indexer to use for transactions
#
# Options:
#   1) "null"
#   2) "kv" (default) - the simplest possible indexer, backed by key-value storage (defaults to levelDB; see DBBackend).
indexer = "{{ .TxIndex.Indexer }}"

# Comma-separated list of tags to index (by default the only tag is "tx.hash")
#
# You can also index transactions by height by adding "tx.height" tag here.
#
# It's recommended to index only a subset of tags due to possible memory
# bloat. This is, of course, depends on the indexer's DB and the volume of
# transactions.
index_tags = "{{ .TxIndex.IndexTags }}"

# When set to true, tells indexer to index all tags (predefined tags:
# "tx.hash", "tx.height" and all tags from DeliverTx responses).
#
# Note this may be not desirable (see the comment above). IndexTags has a
# precedence over IndexAllTags (i.e. when given both, IndexTags will be
# indexed).
index_all_tags = {{ .TxIndex.IndexAllTags }}

##### instrumentation configuration options #####
[instrumentation]

# When true, Prometheus metrics are served under /metrics on
# PrometheusListenAddr.
# Check out the documentation for the list of available metrics.
prometheus = {{ .Instrumentation.Prometheus }}

# Address to listen for Prometheus collector(s) connections
prometheus_listen_addr = "{{ .Instrumentation.PrometheusListenAddr }}"

# Maximum number of simultaneous connections.
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
max_open_connections = {{ .Instrumentation.MaxOpenConnections }}

# Instrumentation namespace
namespace = "{{ .Instrumentation.Namespace }}"
`

// ---------------------------------------------------------------------

// RegLeagueConfigFileTemplate defines the configuration of a reg league
// Note: any changes to the comments/variables/mapstructure
// must be reflected in the appropriate struct in config/config.go
const regLeagueConfigFileTemplate = `# This is a TOML config file.
# It contains the configuration template of the Based League
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# ID of the Base League
league_id = "{{ .BaseLeagueConfig.LeagueId}}"

# TCP or UNIX socket address of the Asura application,
# or the name of an Asura application compiled in with the Dgrid binary
proxy_app = "{{ .BaseLeagueConfig.ProxyApp }}"

# A custom human readable name for this node
hostname = "{{ .BaseLeagueConfig.Hostname }}"

# If this node is many blocks behind the tip of the chain, FastSync
# allows them to catchup quickly by downloading blocks in parallel
# and verifying their commits
fast_sync = {{ .BaseLeagueConfig.FastSync }}

# Database backend: leveldb | memdb | cleveldb
db_backend = "{{ .BaseLeagueConfig.DBBackend }}"

# Database directory
db_dir = "{{ js .BaseLeagueConfig.DBPath }}"

# Output level for logging, including package level options
log_level = "{{ .BaseLeagueConfig.LogLevel }}"

# Output format: 'plain' (colored text) or 'json'
log_format = "{{ .BaseLeagueConfig.LogFormat }}"

##### additional base config options #####

# Path to the JSON file containing the initial validator set and other meta data
genesis_file = "{{ js .BaseLeagueConfig.Genesis }}"

# Path to the JSON file containing the private key to use as a validator in the consensus protocol
validator_key_file = "{{ js .BaseLeagueConfig.ValidatorKey }}"

# Path to the JSON file containing the last sign state of a validator
validator_state_file = "{{ js .BaseLeagueConfig.ValidatorState }}"

# TCP or UNIX socket address for Dgrid to listen on for
# connections from an external Validator process
validator_laddr = "{{ .BaseLeagueConfig.ValidatorListenAddr }}"

# Path to the JSON file containing the private key to use for node authentication in the p2p protocol
node_key_file = "{{ js .BaseLeagueConfig.NodeKey }}"

# Mechanism to connect to the Asura application: socket | grpc
asura = "{{ .BaseLeagueConfig.Asura }}"

# TCP or UNIX socket address for the profiling server to listen on
prof_laddr = "{{ .BaseLeagueConfig.ProfListenAddress }}"

# If true, query the Asura app on connecting to a new peer
# so the app can decide if we should keep the connection or not
filter_peers = {{ .BaseLeagueConfig.FilterPeers }}

##### advanced configuration options #####

##### rpc server configuration options #####
[rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "{{ .RPC.ListenAddress }}"

# A list of origins a cross-domain request can be executed from
# Default value '[]' disables cors support
# Use '["*"]' to allow any origin
cors_allowed_origins = [{{ range .RPC.CORSAllowedOrigins }}{{ printf "%q, " . }}{{end}}]

# A list of methods the client is allowed to use with cross-domain requests
cors_allowed_methods = [{{ range .RPC.CORSAllowedMethods }}{{ printf "%q, " . }}{{end}}]

# A list of non simple headers the client is allowed to use with cross-domain requests
cors_allowed_headers = [{{ range .RPC.CORSAllowedHeaders }}{{ printf "%q, " . }}{{end}}]

# TCP or UNIX socket address for the gRPC server to listen on
# NOTE: This server only supports /broadcast_tx_commit
grpc_laddr = "{{ .RPC.GRPCListenAddress }}"

# Maximum number of simultaneous connections.
# Does not include RPC (HTTP&WebSocket) connections. See max_open_connections
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
# Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
# 1024 - 40 - 10 - 50 = 924 = ~900
grpc_max_open_connections = {{ .RPC.GRPCMaxOpenConnections }}

# Activate unsafe RPC commands like /dial_seeds and /unsafe_flush_storage
unsafe = {{ .RPC.Unsafe }}

# Maximum number of simultaneous connections (including WebSocket).
# Does not include gRPC connections. See grpc_max_open_connections
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
# Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
# 1024 - 40 - 10 - 50 = 924 = ~900
max_open_connections = {{ .RPC.MaxOpenConnections }}

# Maximum number of unique clientIDs that can /subscribe
# If you're using /broadcast_tx_commit, set to the estimated maximum number
# of broadcast_tx_commit calls per block.
max_subscription_clients = {{ .RPC.MaxSubscriptionClients }}

# Maximum number of unique queries a given client can /subscribe to
# If you're using GRPC (or Local RPC client) and /broadcast_tx_commit, set to
# the estimated # maximum number of broadcast_tx_commit calls per block.
max_subscriptions_per_client = {{ .RPC.MaxSubscriptionsPerClient }}

# How long to wait for a tx to be committed during /broadcast_tx_commit.
# WARNING: Using a value larger than 10s will result in increasing the
# global HTTP write timeout, which applies to all connections and endpoints.
# See https://github.com/teragrid/dgrid/issues/3435
timeout_broadcast_tx_commit = "{{ .RPC.TimeoutBroadcastTxCommit }}"

# The name of a file containing certificate that is used to create the HTTPS server.
# If the certificate is signed by a certificate authority,
# the certFile should be the concatenation of the server's certificate, any intermediates,
# and the CA's certificate.
# NOTE: both tls_cert_file and tls_key_file must be present for Dgrid to create HTTPS server. Otherwise, HTTP server is run.
tls_cert_file = "{{ .RPC.TLSCertFile }}"

# The name of a file containing matching private key that is used to create the HTTPS server.
# NOTE: both tls_cert_file and tls_key_file must be present for Dgrid to create HTTPS server. Otherwise, HTTP server is run.
tls_key_file = "{{ .RPC.TLSKeyFile }}"

##### peer to peer configuration options #####
[p2p]

# Address to listen for incoming connections
laddr = "{{ .P2P.ListenAddress }}"

# Address to advertise to peers for them to dial
# If empty, will use the same port as the laddr,
# and will introspect on the listener or use UPnP
# to figure out the address.
external_address = "{{ .P2P.ExternalAddress }}"

# Comma separated list of seed nodes to connect to
seeds = "{{ .P2P.Seeds }}"

# Comma separated list of nodes to keep persistent connections to
persistent_peers = "{{ .P2P.PersistentPeers }}"

# UPNP port forwarding
upnp = {{ .P2P.UPNP }}

# Path to address book
addr_book_file = "{{ js .P2P.AddrBook }}"

# Set true for strict address routability rules
# Set false for private or local networks
addr_book_strict = {{ .P2P.AddrBookStrict }}

# Maximum number of inbound peers
max_num_inbound_peers = {{ .P2P.MaxNumInboundPeers }}

# Maximum number of outbound peers to connect to, excluding persistent peers
max_num_outbound_peers = {{ .P2P.MaxNumOutboundPeers }}

# Time to wait before flushing messages out on the connection
flush_throttle_timeout = "{{ .P2P.FlushThrottleTimeout }}"

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

# Toggle to disable guard against peers connecting from the same ip.
allow_duplicate_ip = {{ .P2P.AllowDuplicateIP }}

# Peer connection configuration.
handshake_timeout = "{{ .P2P.HandshakeTimeout }}"
dial_timeout = "{{ .P2P.DialTimeout }}"

##### FBA consensus configuration options #####
[fba_consensus]

wal_file = "{{ js .FBAConsensusConfig.WalPath }}"

timeout_propose = "{{ .FBAConsensusConfig.TimeoutPropose }}"
timeout_propose_delta = "{{ .FBAConsensusConfig.TimeoutProposeDelta }}"
timeout_prevote = "{{ .FBAConsensusConfig.TimeoutPrevote }}"
timeout_prevote_delta = "{{ .FBAConsensusConfig.TimeoutPrevoteDelta }}"
timeout_precommit = "{{ .FBAConsensusConfig.TimeoutPrecommit }}"
timeout_precommit_delta = "{{ .FBAConsensusConfig.TimeoutPrecommitDelta }}"
timeout_commit = "{{ .FBAConsensusConfig.TimeoutCommit }}"

# Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)
skip_timeout_commit = {{ .FBAConsensusConfig.SkipTimeoutCommit }}

# EmptyBlocks mode and possible interval between empty blocks
create_empty_blocks = {{ .FBAConsensusConfig.CreateEmptyBlocks }}
create_empty_blocks_interval = "{{ .FBAConsensusConfig.CreateEmptyBlocksInterval }}"

# Reactor sleep duration parameters
peer_gossip_sleep_duration = "{{ .FBAConsensusConfig.PeerGossipSleepDuration }}"
peer_query_maj23_sleep_duration = "{{ .FBAConsensusConfig.PeerQueryMaj23SleepDuration }}"

##### BFT consensus configuration options #####
[bft_consensus]

wal_file = "{{ js .BFTConsensusConfig.WalPath }}"

timeout_propose = "{{ .BFTConsensusConfig.TimeoutPropose }}"
timeout_propose_delta = "{{ .BFTConsensusConfig.TimeoutProposeDelta }}"
timeout_prevote = "{{ .BFTConsensusConfig.TimeoutPrevote }}"
timeout_prevote_delta = "{{ .BFTConsensusConfig.TimeoutPrevoteDelta }}"
timeout_precommit = "{{ .BFTConsensusConfig.TimeoutPrecommit }}"
timeout_precommit_delta = "{{ .BFTConsensusConfig.TimeoutPrecommitDelta }}"
timeout_commit = "{{ .BFTConsensusConfig.TimeoutCommit }}"

# Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)
skip_timeout_commit = {{ .BFTConsensusConfig.SkipTimeoutCommit }}

# EmptyBlocks mode and possible interval between empty blocks
create_empty_blocks = {{ .BFTConsensusConfig.CreateEmptyBlocks }}
create_empty_blocks_interval = "{{ .BFTConsensusConfig.CreateEmptyBlocksInterval }}"

# Reactor sleep duration parameters
peer_gossip_sleep_duration = "{{ .BFTConsensusConfig.PeerGossipSleepDuration }}"
peer_query_maj23_sleep_duration = "{{ .BFTConsensusConfig.PeerQueryMaj23SleepDuration }}"

##### transactions indexer configuration options #####
[tx_index]

# What indexer to use for transactions
#
# Options:
#   1) "null"
#   2) "kv" (default) - the simplest possible indexer, backed by key-value storage (defaults to levelDB; see DBBackend).
indexer = "{{ .TxIndex.Indexer }}"

# Comma-separated list of tags to index (by default the only tag is "tx.hash")
#
# You can also index transactions by height by adding "tx.height" tag here.
#
# It's recommended to index only a subset of tags due to possible memory
# bloat. This is, of course, depends on the indexer's DB and the volume of
# transactions.
index_tags = "{{ .TxIndex.IndexTags }}"

# When set to true, tells indexer to index all tags (predefined tags:
# "tx.hash", "tx.height" and all tags from DeliverTx responses).
#
# Note this may be not desirable (see the comment above). IndexTags has a
# precedence over IndexAllTags (i.e. when given both, IndexTags will be
# indexed).
index_all_tags = {{ .TxIndex.IndexAllTags }}

##### instrumentation configuration options #####
[instrumentation]

# When true, Prometheus metrics are served under /metrics on
# PrometheusListenAddr.
# Check out the documentation for the list of available metrics.
prometheus = {{ .Instrumentation.Prometheus }}

# Address to listen for Prometheus collector(s) connections
prometheus_listen_addr = "{{ .Instrumentation.PrometheusListenAddr }}"

# Maximum number of simultaneous connections.
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
max_open_connections = {{ .Instrumentation.MaxOpenConnections }}

# Instrumentation namespace
namespace = "{{ .Instrumentation.Namespace }}"
`
