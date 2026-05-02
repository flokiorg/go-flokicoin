# Flokicoin Core (go-flokicoin)

Lokichain is the public, permissionless blockchain that secures Flokicoin. Flokicoin Core (`go-flokicoin`) is the reference node software: it validates blocks and transactions, enforces consensus rules, and exposes network and wallet interfaces.

## Binaries

The project builds two primary binaries:
- `lokid`: The Flokicoin full node daemon.
- `lokid-cli`: A command-line interface to interact with `lokid`.

## Installation

### Requirements
- [Go](https://golang.org) 1.26.1 or newer.

### Build from Source
```bash
git clone https://github.com/flokiorg/go-flokicoin
cd go-flokicoin
go install -v . ./cmd/...
```

## Usage

Start the node:
```bash
lokid
```

Interacting with the node:
```bash
lokid-cli getblockchaininfo
```

Learn more: [Official Documentation](https://docs.flokicoin.org/lokichain)
