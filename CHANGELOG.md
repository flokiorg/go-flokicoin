# Changelog

## [0.26.2]

### Fixed

- `lokid` and `lokid-cli` reported versions unrelated to what was released: the
  daemon said 0.25.12-alpha and the CLI said 0.23.2-beta, because each composed
  its version from hand-edited constants that were never bumped. The version is
  now taken from the release tag at build time, and the numeric components, the
  P2P user agent and the `/lokid:.../` RPC sub-version string are all derived
  from it, so there is nothing left to drift. (#13)
- The RPC sub-version string no longer ends in a stray hyphen when the version
  carries no pre-release label. (#13)

### Changed

- Built with Go 1.26.5. (#11)

## [0.26.1-alpha]

go-flokicoin v0.26.1-alpha adds CI (build/vet/test on every push and PR) and fixes 13 `go vet` findings surfaced by the first real vet run under CI.

### CI

- Added `.github/workflows/ci.yaml`: runs `go build`, `go vet`, and `go test` on push to `main` and on pull requests.

### Fixes

- **`rpcclient`**: fixed bare IPv6 literal handling in `ParseAddressString`.
- **`peer`**, **`mining/cpuminer`**: two call sites passed an already-formatted error string as the format string itself to `Errorf` — a real risk if the message ever contains a literal `%`. Switched to passing the string as an argument.
- **`mining`**: fixed a wrong format verb for `feePerKB` in a log line.
- **`chainjson`**: fixed a non-constant format string in the help-text writer.
- **`server`**: switched a filter-type log call from `Debug` to `Debugf` so its arguments are actually interpolated.
- **`txscript`**: fixed the format verb used for `ScriptInfo` in a test failure message.
- **`internal/musig2v040`**-equivalent (`crypto/schnorr/musig2`): stopped calling `t.Fatalf` from worker goroutines (not safe outside the test's own goroutine).
- **`blockchain`**: removed unreachable code left after a disabled rule-activation warning path; switched two disabled tests from a bare `return` to `t.Skip` so they're reported as skipped rather than silently doing nothing.
- **`integration`**: switched a disabled invalidate/reconsider block test from a bare `return` to `t.Skip`.
- **`wire`**: removed a duplicate `return` statement in an auxpow test helper.

All 13 are mechanical vet-driven fixes — no behavioral changes to consensus, signing, or networking logic.

Commit range: `0.26.0-alpha..0.26.1-alpha` (13 commits + 1 merge).

## [0.26.0-alpha]

go-flokicoin v0.26.0-alpha ports a batch of security hardening, protocol correctness, and RPC improvements from upstream btcd.

### Security & Robustness

- **BIP30 enforcement hardening**: ported upstream's stricter BIP30 (duplicate coinbase) enforcement logic.
- **Panic recovery**: peer message-handling goroutines now recover from panics instead of taking down the process.
- **Wire protocol hardening**: `ReadMessage` now enforces full payload consumption; witness item reads are bounded to the remaining decode slab, closing a class of malformed-message issues.
- **Address validation**: `addrv2` now rejects/reclassifies IPv4-mapped IPv6 addresses; RFC7343 ORCHIDv2 and zero-prefix IPv6 addresses are rejected as unroutable.

### RPC Client

- Fixed a shutdown double-resolve race and added `httpURL` caching.
- In-flight HTTP requests can now be canceled; several batch-request bugs fixed.
- Added `GetTxOutProof`, a dial-timeout fix, and faster JSON string parsing.

### Networking & Sync

- `netsync` now allows syncing with non-localhost peers on regtest/simnet.
- Fixed a peer add/done race between `peerHandler` and `syncManager`.

### Fixes

- Corrected a `Debugf` call in chain-tip reconsideration that was missing its argument (would have printed a blank tip count instead of the actual number of inactive tips being examined).

### Dependency Security

- Bumped `golang.org/x/crypto` from `v0.45.0` to `v0.52.0`, closing 7 upstream advisories in the `ssh`/`ssh/agent` subpackage (GHSA-rm3j-f69w-wqmq, GHSA-vgwf-h737-ff37, GHSA-f5wc-c3c7-36mc, GHSA-jppx-rxg9-jmrx, GHSA-89gr-r52h-f8rx, GHSA-5cgq-3rg8-m6cv, GHSA-x527-x647-q7gg). None of these code paths are reachable in go-flokicoin (only `pbkdf2`/`scrypt`/`ripemd160`/`sha3`/`nacl/secretbox` are used, never `ssh`), but the fix is a clean, zero-risk version bump.

Commit range: `0.25.13-alpha..0.26.0-alpha` (13 commits).

## [0.25.13-alpha]

### New Features

#### MuSig2 (Aggregated Nonce Signing)

- Added support for aggregated nonce signing flow in the MuSig2 implementation.
- Included comprehensive test coverage for the new signing flow.

### Bug Fixes

#### Network Parameters

- **TestNet4 P2P Port**: Corrected the TestNet4 default P2P port from the Bitcoin value (`48333`) to the Flokicoin value (`65212`).
- **Comment Cleanup**: Updated inline comments in `chaincfg/params.go` to reference Flokicoin rather than Bitcoin for the relevant network port entries.

### Dependency Updates

- **Go SDK**: Updated minimum Go version requirement to `1.26.1`.
- Routine `go mod tidy` cleanup.

## [0.25.12-alpha]

### Highlights

- Added aggregated nonce signing flow for MuSig2, including updated context handling and comprehensive tests.

## [0.25.11-alpha]

Focus is on fee estimation durability, RPC polish, and better accounting for mempool churn.

### Upgrade notes
- Fee estimator snapshots now live on disk as `fee_estimates.dat` under the data dir (atomic write/read). Ensure the daemon can write there on shutdown; stale files older than ~60h are ignored on startup.

### Highlights
  - **Fee estimation**
    - Tracks up to 1008 blocks of history and caps long-horizon queries instead of rejecting them, aligning with Bitcoin Core behavior (`mempool/estimatefee.go`).
    - Records mempool removal reasons so unconfirmed drops (conflicts, reorgs, evictions, rejects) clear estimator state and reduce stale observations (`mempool` plumbing, `netsync/manager.go`, `rpcserver.go`).
    - `estimatesmartfee` accepts `economical` or `conservative` modes (defaults conservative) and reports invalid modes cleanly; result now includes an explicit `errors` field (`rpcserver.go`).
  - **Persistence**
    - Fee estimator state is saved atomically to disk on shutdown and restored on startup when fresh; stale data surfaces an error and is skipped. Helpers and tests cover the binary format and max-age enforcement (`mempool/fee_persist*.go`, `server.go`).
### Notable commits (planned batch)
- mempool: track removal reasons in fee estimation; widen history to 1008 blocks and cap queries.
- mempool: persist fee estimator to disk with atomic writes and staleness checks.
- versioning: bump to 0.25.11-beta (binary reports alpha prerelease string).

### Testing
- Fee estimator persistence round-trip and staleness are covered by new unit tests. Broader regression/CI test suites have not been executed for this prerelease; run `go test ./...` in your environment if needed.

## [0.25.10-beta]

This version focuses on protocol/RPC refinements, improved sync heuristics, refreshed checkpoints, and a daemon rename. Nodes should upgrade to stay aligned with the network and to use the updated daemon/CLI binaries.

### Upgrade notes
- The daemon and CLI were renamed to `lokid` and `lokid-cli` (formerly `flokicoind`/`flokicoind-cli`). Update service files, scripts, and container entrypoints accordingly.

### Highlights
- **Consensus & chain rules**
  - Refreshed mainnet checkpoints through height 209,771 to speed up initial sync (`chaincfg/params.go`).
- **Networking & sync**
  - Improved outbound group tracking and ensured peers always answer `getblocks`, reducing stalls when nearly synced (`server.go`, `peer` tests).
  - Tightened download scheduling when the node is already current to avoid redundant block fetches (`netsync/manager.go`).
  - Refined block header/message serialization helpers for more robust IO paths (`wire/blockheader.go`, `wire/msgblock.go`).
- **RPC**
  - `getblockchaininfo` now computes verification progress from the best known peer headers, and `getblock` exposes richer details for blocks (`rpcserver.go`, `chainjson/chainsvrresults.go`).
  - `getrawtransaction` response types and help text were aligned with actual payloads; `createrawtransaction` now enforces stricter amount validation (`rpcserverhelp.go`, `rpcserver.go`).
  - Added block statistics aggregation backing `getblockstats`, including fee/size percentiles and UTXO deltas (`blockchain/stats/blockstats.go`, `rpcserver.go`).
  - `getinfo` now reports the correct P2P protocol version and fills previously missing fields (`rpcserver.go`).
- **Mempool & validation**
  - Switched value bounds to full `int64` range and guarded dust checks against overflow when computing `MaxLoki`/fee policy (`chainutil`, `mempool`).
  - Validation errors now emit properly cast `MaxLoki` values to avoid misleading messages (`blockchain`).
- **Tooling, tests, docs**
  - Added a deterministic block dataset generator (`make testexport`) and refreshed blockchain fixtures for CI stability (`cmd/testexport`, `blockchain/testdata`).
  - Strengthened difficulty/statistics test coverage and outbound group accounting tests (`blockchain`, `peer`).

### Notable commits (chronological)
- Tooling/fixtures: repo/test data refresh and `testexport` helper (`31eff16`, `0e5dcd3`).
- RPC/build: protocol/version reporting fixes and banner updates (`651a453`, `2e92497`).
- Policy safety: `MaxLoki` range and dust overflow guards (`9d9eef1`, `5b583c0`).
- RPC correctness: tighter `createrawtransaction`, aligned `getrawtransaction` types, block stats package (`4483bfa`, `ed3f11b`, `04bfa8b`).
- Daemon rename, new checkpoints, and sync/p2p/rpc refinements leading to 0.25.10-beta (`66a4658`, `037c55c`, `9f457dd`, `3388ec1`, `3d115e4`, `5eadc7b`, `91a004e`, `fbbef28`).

## [0.25.8-beta]

- This is a **pre-release** for testing and feedback.
- Developers and early adopters are encouraged to **report issues**.

## [0.25.7-beta]

This is a minor beta release of Flokicoin Core with a new per-block difficulty retarget algorithm (Digishield-style), AuxPoW header and consensus support, testnet feature updates, RPC corrections, and build/test improvements. This release schedules **`MAINNET ACTIVATION`** for both Digishield and AuxPoW via chain parameters. All mainnet nodes must upgrade before the activation heights below to avoid chain splits.

### Notable changes

- Consensus: Digishield difficulty retarget
  - Introduces a per-block exponential moving average retarget toward the target spacing with amplitude divisor 8.
  - Applies bounded damping each step (min 0.75x, max 1.5x of target spacing influence) to reduce volatility.
  - Supports min-difficulty on late blocks where enabled by chain parameters.
  - Integrates into validation and mining (next-required-bits, block template target) and expands unit/full-block tests.

- P2P/Wire: AuxPoW header support
  - Adds AuxPoW structures (coinbase, merkle branches, parent header) and serialization helpers.
  - Extends `BlockHeader` with AuxPoW flag and ChainID utilities plus optional AuxPoW payload encode/decode.
  - Updates `MsgBlock` decoding to parse AuxPoW payloads when present; adds coverage for read/write paths.
  - Consensus: AuxPoW is activated on mainnet at the height listed below.

- Testnet/regtest updates
  - Enables SegWit and Taproot features on testnet where configured.
  - Tunes `MinHighPriority` and increases the `MaxLoki` bound used in RPC amount validation.

- RPC fixes and improvements
  - Aligns reported P2P protocol version with `wire.ProtocolVersion` and fills `getinfo` fields (subversion, localservices, connections_in/out, localaddresses) from live server state.

- Tooling and tests
  - Adds `make testexport` target and a deterministic test-data exporter.
  - Unifies block dataset loader with mandatory network-magic validation; refactors call sites.
  - Adds an end-to-end difficulty validation over curated datasets and emits steady-state tuning metrics.

- Build and configuration
  - Drops local `replace` override for `flokicoin-neutrino` to avoid developer-only paths.
  - Prints a single startup banner with semantic version during config load; refactors ASCII-art constants.

### Compatibility

- Mainnet will activate Digishield and AuxPoW at fixed heights (below). These are consensus changes and constitute a hard fork at activation. All miners, validators, and services must upgrade before those heights.
- AuxPoW wire support remains backward compatible for non-AuxPoW headers until activation height.
- No database format changes are introduced.

### Activation schedule (mainnet)

- Digishield difficulty retarget: activates at height 115,000 (`DigishieldActivationHeight` in `chaincfg/params.go`).
- AuxPoW consensus: activates at height 115,840 (`AuxpowHeightEffective` in `chaincfg/params.go`).

These activations are parameterized in `repos/flokiorg/go-flokicoin/chaincfg/params.go` under `MainNetParams`.

### Upgrade notices

- Mainnet node operators: upgrade to 0.25.7-beta before heights 115,000 (Digishield) and 115,840 (AuxPoW) to remain on the canonical chain.
- Testnet/regtest operators should upgrade to participate in the updated difficulty and feature policies.
- Application developers should re-vendor/update to pick up RPC schema fixes and wire additions related to AuxPoW.

### RPC changes

- `getinfo` now reports accurate `subversion`, `localservices`, `localservicesnames`, `connections_in`, `connections_out`, and `localaddresses`.
- Reported max protocol version aligns with `wire.ProtocolVersion` (70016).

### Build system and tooling

- New `make testexport` target to export block datasets for deterministic tests.
- Removed local `replace` for `flokicoin-neutrino` to produce clean module graphs in builds.

### Tests and QA

- Unified block loader enforces network magic and uses buffered, full-length reads.
- Added E2E difficulty validation with metrics output for steady-state analysis.
- Curated, smaller test datasets improve determinism and CI time.

### Changelog since v0.25.6-beta

- b4f439e consensus: add Digishield difficulty retarget
- dec0948 wire: integrate AuxPoW header support
- 31eff16 chore: repo updates for rpc/db, tests, fixtures
- e2ac6e7 tests: validate difficulty dataset and expose tuning metrics
- 735410f tests: unify block loader to enforce network magic; refactor call sites
- 0e5dcd3 make: add testexport target for exporting block datasets
- 651a453 rpc: report correct p2p protocol and fill getinfo fields
- bea6a7a build: drop local replace for flokicoin-neutrino
- 2e92497 config: print startup banner with version; refactor art constants
- b24e870 wire: remove stale TODO comment in protocol.go
- 60460f8 enable segiwit/taproot for testnet
- f9bf3ae increase maxloki const
- 712355d update MinHighPriority const
- fa96bbd bump version 0.25.7-dev
- 4523b79 version: bump to 0.25.7-beta

### Credits

Thanks to everyone who directly contributed to this release:

- naliyi
- izimmerma
- mechion
- repins
- nanovex
- liarpo11
- arowqo
- thiniboy

## [0.25.6-beta]

- Update testnet3 and regression chain parameters
- Adjust difficulty logic for testnet4
- Restore default PoW limits for regression network
- Update version bits deployment thresholds and timings
- Remove legacy difficulty calculation fallback
- Clean up block version test formatting

## [0.25.5-beta]

- Added new blockchain checkpoints, including at block 0 and later heights.
- Enhanced PSBT handling: support for extended public keys (XPubs) and path validation.
- Added PSBT test coverage for BIP32 path serialization/deserialization.
- Improved validation of PSBT path minimum length.
- Improved block header timestamp validation.
- Updated mempool, txscript, and wallet to use maps.Copy for safer and cleaner map handling.
- Fixed typo: corrected PSBT constant from `XpubType` to `XPubType`.
- Fixed minor issues in PSBT encoding/decoding and comments.
- Renamed `flokicoin-cli` to `lokid-cli` across all binaries, config paths, and documentation.

## [0.25.1-beta]

- This is a **pre-release** for testing and feedback.
- Developers and early adopters are encouraged to **report issues**.
