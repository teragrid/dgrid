## 0.0.1 (May 22, 2018)

BREAKING CHANGES:

- Strict SemVer starting now!


FEATURES:

- TMSP and RPC support TCP and UNIX sockets
- Addition config options including block size and consensus parameters
- New WAL mode `cswal_light`; logs only the validator's own votes
- New RPC endpoints:
	- for starting/stopping profilers, and for updating config
	- `/broadcast_tx_commit`, returns when tx is included in a block, else an error
	- `/unsafe_flush_mempool`, empties the mempool


IMPROVEMENTS:

- Various optimizations
- Remove bad or invalidated transactions from the mempool cache (allows later duplicates)
- More elaborate testing using CircleCI including benchmarking throughput on 4 digitalocean droplets

BUG FIXES:

- Various fixes to WAL and replay logic
- Various race conditions

## PreHistory


