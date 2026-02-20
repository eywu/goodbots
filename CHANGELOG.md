# Changelog

All notable changes to this project are documented in this file.

## [v0.0.3-cetoddle] - 2026-02-20

### Added
- `-c` CLI flag to configure concurrency (default `50`)
- `-mode` CLI flag (`goodbots` | `resolve`)
- New tests for:
  - DNS server round-robin behavior
  - context timeout/cancellation handling
  - processing final input line without trailing newline
  - concurrent output line integrity

### Changed
- Reused a single package-level DNS resolver (concurrency-safe)
- Replaced random resolver selection with deterministic atomic round-robin
- Added per-lookup 5s context timeouts for reverse and forward DNS calls
- Reworked output handling to channel-based single writer + buffered output
- Updated `go.mod` minimum Go version to `1.23.0`
- README updated for new CLI flags and behavior

### Fixed
- Last-line EOF handling bug (previously dropped final line without `\n`)
- Potential interleaving risk from concurrent writes to shared `io.Writer`
