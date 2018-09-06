# Teragrid

[Byzantine-Fault Tolerant](https://en.wikipedia.org/wiki/Byzantine_fault_tolerance)
([Federated Byzantine Agreement](https://www.stellar.org/papers/stellar-consensus-protocol.pdf))
[State Machine Replication](https://en.wikipedia.org/wiki/State_machine_replication).
Or [Blockchain](https://en.wikipedia.org/wiki/Blockchain_(database)) for short.

[![version](https://img.shields.io/github/tag/teragrid/teragrid.svg)](https://github.com/teragrid/teragrid/releases/latest)
[![Go version](https://img.shields.io/badge/go-1.9.2-blue.svg)](https://github.com/moovweb/gvm)
[![license](https://img.shields.io/github/license/teragrid/teragrid.svg)](https://github.com/teragrid/teragrid/blob/master/LICENSE)
[![](https://tokei.rs/b1/github/teragrid/teragrid?category=lines)](https://github.com/teragrid/teragrid)


_NOTE: This is alpha software. Please contact us if you intend to run it in production._

Teragrid Core is Multi-Consensus middleware that takes a state transition machine - written in any programming language -
and securely replicates it on many machines.

For more information, from introduction to install to application development, [Read The Docs](https://teragrid.readthedocs.io/en/master/).

## Minimum requirements

Requirement|Notes
---|---
Go version | Go1.9 or higher

## Install

To download pre-built binaries, see our [downloads page](https://teragrid.network/downloads).

To install from source, you should be able to:

`go get -u github.com/teragrid/teragrid/cmd/teragrid`

For more details (or if it fails), [read the docs](https://teragrid.readthedocs.io/en/master/install.html).

## Resources

### Teragrid Core

All resources involving the use of, building application on, or developing for, teragrid, can be found at [Read The Docs](https://teragrid.readthedocs.io/en/master/). Additional information about some - and eventually all - of the sub-projects below, can be found at Read The Docs.

### Sub-projects

* [Asura](http://github.com/teragrid/asura), the Application Blockchain Interface
* [Parkhill](http://github.com/teragrid/parkhill), a D-Apps MVC Framwork for rapid development and adoptation 

### Applications

* [TeraReal](http://github.com/teragrid/terareal); A real estate platform on Teragrid
* [TeraCoin](http://github.com/teragrid/teracoin); Teragrid cryptocurrency
* [TeraWallet](http://github.com/teragrid/terawallet); Teragrid wallet
* [TeraTalk](http://github.com/teragrid/teratalk); A Secure Message Platform Integrated in Teragrid
* [Many more](https://github.com/teragrid/teradocs/wiki)

### More

* [Original Whitepaper](https://github.com/teragrid/teradocs/wiki/Teragrid-White-Paper)
* [Teragrid Blog](https://medium.com/teragrid-network)

## Contributing

Yay open source! Please see our [contributing guidelines](CONTRIBUTING.md).

## Versioning

### SemVer

Teragrid uses [SemVer](http://semver.org/) to determine when and how the version changes.
According to SemVer, anything in the public API can change at any time before version 1.1.0

To provide some stability to Teragrid users in these 0.X.X days, the MINOR version is used
to signal breaking changes across a subset of the total public API. This subset includes all
interfaces exposed to other processes (cli, rpc, p2p, etc.), as well as parts of the following packages:

- types
- rpc/client
- config
- node

Exported objects in these packages that are not covered by the versioning scheme
are explicitly marked by `// UNSTABLE` in their go doc comment and may change at any time.
Functions, types, and values in any other package may also change at any time.


## Code of Conduct

Please read, understand and adhere to our [code of conduct](CODE_OF_CONDUCT.md).
