# teragrid

## Overview

This is a quick start guide. If you have a vague idea about how teragrid
works and want to get started right away, continue. Otherwise, [review the
documentation](http://teragrid.readthedocs.io/en/master/).

## Install

### Quick Install

On a fresh Ubuntu 16.04 machine can be done with [this script](https://git.io/vNLfY), like so:

```
curl -L https://git.io/vxWlX | bash
source ~/.profile
```

WARNING: do not run the above on your local machine.

The script is also used to facilitate cluster deployment below.

### Manual Install

Requires:
- `go` minimum version 1.9
- `$GOPATH` environment variable must be set
- `$GOPATH/bin` must be on your `$PATH` (see https://github.com/teragrid/teragrid/wiki/Setting-GOPATH)

To install teragrid, run:

```
go get github.com/teragrid/teragrid
cd $GOPATH/src/github.com/teragrid/teragrid
make get_tools && make get_vendor_deps
make install
```

Note that `go get` may return an error but it can be ignored.

Confirm installation:

```
$ teragrid version
0.18.0-XXXXXXX
```

## Initialization

Running:

```
teragrid init
```

will create the required files for a single, local node.

These files are found in `$HOME/.teragrid`:

```
$ ls $HOME/.teragrid

config.toml  data  genesis.json  priv_validator.json
```

For a single, local node, no further configuration is required.
Configuring a cluster is covered further below.

## Local Node

Start teragrid with a simple in-process application:

```
teragrid node --proxy_app=kvstore
```

and blocks will start to stream in:

```
I[01-06|01:45:15.592] Executed block                               module=state height=1 validTxs=0 invalidTxs=0
I[01-06|01:45:15.624] Committed state                              module=state height=1 txs=0 appHash=
```

Check the status with:

```
curl -s localhost:46657/status
```

### Sending Transactions

With the kvstore app running, we can send transactions:

```
curl -s 'localhost:46657/broadcast_tx_commit?tx="abcd"'
```

and check that it worked with:

```
curl -s 'localhost:46657/asura_query?data="abcd"'
```

We can send transactions with a key and value too:

```
curl -s 'localhost:46657/broadcast_tx_commit?tx="name=satoshi"'
```

and query the key:

```
curl -s 'localhost:46657/asura_query?data="name"'
```

where the value is returned in hex.

## Cluster of Nodes

First create four Ubuntu cloud machines. The following was tested on Digital
Ocean Ubuntu 16.04 x64 (3GB/1CPU, 20GB SSD). We'll refer to their respective IP
addresses below as IP1, IP2, IP3, IP4.

Then, `ssh` into each machine, and execute [this script](https://git.io/vNLfY):

```
curl -L https://git.io/vNLfY | bash
source ~/.profile
```

This will install `go` and other dependencies, get the teragrid source code, then compile the `teragrid` binary.

Next, `cd` into `docs/examples`. Each command below should be run from each node, in sequence:

```
teragrid node --home ./node1 --proxy_app=kvstore --p2p.persistent_peers="3a558bd6f8c97453aa6c2372bb800e8b6ed8e6db@IP1:46656,ccf30d873fddda10a495f42687c8f33472a6569f@IP2:46656,9a4c3de5d6788a76c6ee3cd9ff41e3b45b4cfd14@IP3:46656,58e6f2ab297b3ceae107ba4c8c2898da5c009ff4@IP4:46656"
teragrid node --home ./node2 --proxy_app=kvstore --p2p.persistent_peers="3a558bd6f8c97453aa6c2372bb800e8b6ed8e6db@IP1:46656,ccf30d873fddda10a495f42687c8f33472a6569f@IP2:46656,9a4c3de5d6788a76c6ee3cd9ff41e3b45b4cfd14@IP3:46656,58e6f2ab297b3ceae107ba4c8c2898da5c009ff4@IP4:46656"
teragrid node --home ./node3 --proxy_app=kvstore --p2p.persistent_peers="3a558bd6f8c97453aa6c2372bb800e8b6ed8e6db@IP1:46656,ccf30d873fddda10a495f42687c8f33472a6569f@IP2:46656,9a4c3de5d6788a76c6ee3cd9ff41e3b45b4cfd14@IP3:46656,58e6f2ab297b3ceae107ba4c8c2898da5c009ff4@IP4:46656"
teragrid node --home ./node4 --proxy_app=kvstore --p2p.persistent_peers="3a558bd6f8c97453aa6c2372bb800e8b6ed8e6db@IP1:46656,ccf30d873fddda10a495f42687c8f33472a6569f@IP2:46656,9a4c3de5d6788a76c6ee3cd9ff41e3b45b4cfd14@IP3:46656,58e6f2ab297b3ceae107ba4c8c2898da5c009ff4@IP4:46656"
```

Note that after the third node is started, blocks will start to stream in
because >2/3 of validators (defined in the `genesis.json`) have come online.
Seeds can also be specified in the `config.toml`. See [this
PR](https://github.com/teragrid/teragrid/pull/792) for more information
about configuration options.

Transactions can then be sent as covered in the single, local node example above.
