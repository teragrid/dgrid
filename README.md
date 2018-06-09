# Teragrid

[Byzantine-Fault Tolerant](https://en.wikipedia.org/wiki/Byzantine_fault_tolerance)
[State Machine Replication](https://en.wikipedia.org/wiki/State_machine_replication).
Or [Blockchain](https://en.wikipedia.org/wiki/Blockchain_(database)) for short.

[![version](https://img.shields.io/github/tag/teragrid/teragrid.svg)](https://github.com/teragrid/teragrid/releases/latest)
[![API Reference](
https://teragrid.network/api/docs
)](https://godoc.org/github.com/teragrid/teragrid)
[![Go version](https://img.shields.io/badge/go-1.9.2-blue.svg)](https://github.com/moovweb/gvm)
[![Rocket.Chat](https://demo.rocket.chat/images/join-chat.svg)](https://cosmos.rocket.chat/)
[![license](https://img.shields.io/github/license/teragrid/teragrid.svg)](https://github.com/teragrid/teragrid/blob/master/LICENSE)
[![](https://tokei.rs/b1/github/teragrid/teragrid?category=lines)](https://github.com/teragrid/teragrid)


Branch    | Tests | Coverage
----------|-------|----------
master    | [![CircleCI](https://circleci.com/gh/teragrid/teragrid/tree/master.svg?style=shield)](https://circleci.com/gh/teragrid/teragrid/tree/master) | [![codecov](https://codecov.io/gh/teragrid/teragrid/branch/master/graph/badge.svg)](https://codecov.io/gh/teragrid/teragrid)
develop   | [![CircleCI](https://circleci.com/gh/teragrid/teragrid/tree/develop.svg?style=shield)](https://circleci.com/gh/teragrid/teragrid/tree/develop) | [![codecov](https://codecov.io/gh/teragrid/teragrid/branch/develop/graph/badge.svg)](https://codecov.io/gh/teragrid/teragrid)

_NOTE: This is alpha software. Please contact us if you intend to run it in production._

Teragrid Core is Byzantine Fault Tolerant (BFT) middleware that takes a state transition machine - written in any programming language -
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

### Tools
* [Deployment, Benchmarking, and Monitoring](http://teragrid.readthedocs.io/projects/tools/en/develop/index.html#teragrid-tools)

### Applications

* [TeraReal](http://github.com/teragrid/terareal); A real estate platform on Teragrid
* [TeraCoin](http://github.com/teragrid/teracoin); Teragrid cryptocurrency
* [TeraWallet](http://github.com/teragrid/terawallet); Teragrid wallet
* [TeraTalk](http://github.com/teragrid/teratalk); A Secure Message Platform Integrated in Teragrid
* [Many more](https://teragrid.readthedocs.io/en/master/ecosystem.html)

### More

* [Original Whitepaper](https://teragrid.network/static/docs/teragrid-whitepaper.pdf)
* [Teragrid Blog](https://blog.teragrid.network/teragrid/home)

## Contributing

Yay open source! Please see our [contributing guidelines](CONTRIBUTING.md).

## Versioning

### SemVer

Teragrid uses [SemVer](http://semver.org/) to determine when and how the version changes.
According to SemVer, anything in the public API can change at any time before version 1.0.0

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

### Upgrades

In an effort to avoid accumulating technical debt prior to 1.0.0,
we do not guarantee that breaking changes (ie. bumps in the MINOR version)
will work with existing teragrid blockchains. In these cases you will
have to start a new blockchain, or write something custom to get the old
data into the new chain.

However, any bump in the PATCH version should be compatible with existing histories
(if not please open an [issue](https://github.com/teragrid/teragrid/issues)).

## Code of Conduct

Please read, understand and adhere to our [code of conduct](CODE_OF_CONDUCT.md).
