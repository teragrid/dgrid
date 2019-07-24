# Application BlockChain Interface (Asura)

Blockchains are systems for multi-master state machine replication.
**Asura** is an interface that defines the boundary between the replication engine (the blockchain),
and the state machine (the application).
Using a socket protocol, a consensus engine running in one process
can manage an application state running in another.

Previously, the Asura was referred to as TMSP.

The community has provided a number of addtional implementations, see the [Dgrid Ecosystem](https://teragrid.network/ecosystem)


## Installation & Usage

To get up and running quickly, see the [getting started guide](../docs/app-dev/getting-started.md) along with the [asura documentation](../docs/app-dev/asura.md) which will go through the examples found in the [examples](./example/) directory.

## Specification

A detailed description of the Asura methods and message types is contained in:

- [The main spec](../docs/spec/asura/asura.md)
- [A protobuf file](./types/types.proto)
- [A Go interface](./types/application.go)

## Protocol Buffers

To compile the protobuf file, run (from the root of the repo):

```
make protoc_abci
```

See `protoc --help` and [the Protocol Buffers site](https://developers.google.com/protocol-buffers)
for details on compiling for other languages. Note we also include a [GRPC](https://www.grpc.io/docs)
service definition.
