# Changelog

All notable changes are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Centralizer follows [SemVer](https://semver.org/). The public API is not stable before 1.0.

## [Unreleased]

### Changed

- Official mark is a hand-drawn hexagonal C: off-white structure, copper hub, three inbound radii, one unified exit (`assets/`, `@theworker02/centralizer-brand`, website)

## [0.1.1] - 2026-08-15

### Added

- Official brand package `@theworker02/centralizer-brand` under `packages/centralizer-brand/` (SVG/PNG from `assets/`, not published unless requested)
- Website Vite plugin copies `assets/logo.svg` and `assets/icon.svg` into `website/public` at build start
- `bridge.TransportName` maps planner strategies to transport labels
- Tests for `TransportName` and `splitRef` drive-letter / entry parsing
- Expanded README: architecture, CIR, planner, supervisor, protocol, manifest/lock/policy, full CLI, advanced API, security, telemetry, brand, `centralizerd`

### Changed

- Version identity is `0.1.1` (`internal/version`, `Makefile`)
- Documentation site `package.json` records `logo` and a `file:` dependency on the brand package
- Doctor on Windows reports the named-pipe client foundation instead of “planned”
- `splitRef` treats a Windows drive prefix (`C:`) as a path, not a dead-code branch beside `i > 1`

### Fixed

- Supervisor fallback reconnect used `string(strategy)` as the transport, so `in_process` would be retried as transport `in_process` instead of `native`
- Removed the unused `filepath.Separator` no-op in `centralizerd`
- Removed the identical Windows/non-Windows return in `FilterEnv`

### Security

- No change to the trust model. Path validation, env filtering, loopback daemon, and frame limits remain as in 0.1.0.

## [0.1.0] - 2026-08-14

### Added

- Go hub API: `Connect`, `Call`, `Invoke`, `Get`, `Set`, `New`, `Release`, `Stream`, `Subscribe`, `Describe`
- CIR kind-tagged values with JSON wire encoding, conversion, and validation
- Centralizer Protocol 1.0 (NDJSON on stdio, length-prefixed frames on sockets)
- Discovery with confidence scores for Go, Python, Node, Rust, C, C++, WASM, JVM, .NET, Ruby, PHP, Swift, Dart, Lua, Zig
- Deterministic bridge planner with explainable output
- Supervisor with bounded recovery and circuit breaker
- Python and Node generated shims (cache-resident, disposable)
- Go in-process native handlers
- Rust stdio protocol speakers (`cargo run -- --centralizer`)
- CLI: detect, inspect, describe, connect, call, health, list, graph, explain, bench, trace, doctor, cache, version
- Optional `centralizerd` loopback registry
- Manifest and policy engine
- Unit, integration, failure-injection, and fuzz tests
- Documentation site sources and Apache 2.0 license
- GitHub Pages workflow (`.github/workflows/pages.yml`) publishing `website/dist` to https://theworker02.github.io/centralizer/
- Optional `centralizer.lock` snapshot of a resolved plan (`centralizer lock`, `inspect` can read it)
- `centralizer init` writes a starter `centralizer.yaml`
- `centralizer adapters` lists tier and claimed capabilities
- Python and Node shim `STREAM_OPEN` / `STREAM_DATA` / `STREAM_CLOSE` for generators and iterables
- Localhost TCP framed transport when the planner selects `tcp` (`prefer: [tcp]`)
- Explicit `schema.yaml` / manifest `schema:` load with call validation
- Handle TTL and `DropBridge` on `Service.Close`
- `centralizerd` `GET /v1/metrics` JSON snapshot; OpenTelemetry hook notes in `docs/telemetry.md`
- Windows named-pipe transport foundation (dials an existing pipe; server side remains experimental)
- Doctor checks for Git, cache writability, and protocol version

### Changed

- Documentation site expanded with architecture flow, honest compatibility matrix, and changelog
- README docs badge; Vite `base` remains `./` for local and project Pages

### Security

- Path validation, environment filtering, localhost-only daemon, payload size limits, handle validation
