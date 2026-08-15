<p align="center">
  <img src="assets/logo.svg" width="120" alt="Centralizer logo">
</p>

<h1 align="center">CENTRALIZER</h1>
<p align="center"><strong>One runtime. Every language.</strong></p>

<p align="center">
  Centralizer is a Go-based interoperability runtime<br>
  that discovers, connects, supervises, and exposes<br>
  software written across different programming<br>
  languages through a unified interface.
</p>

<p align="center">
  <a href="https://github.com/theworker02/centralizer/actions/workflows/test.yml"><img src="https://github.com/theworker02/centralizer/actions/workflows/test.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/theworker02/centralizer/pkg/centralizer"><img src="https://pkg.go.dev/badge/github.com/theworker02/centralizer/pkg/centralizer.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/theworker02/centralizer"><img src="https://goreportcard.com/badge/github.com/theworker02/centralizer" alt="Go Report Card"></a>
  <a href="https://github.com/theworker02/centralizer/security/code-scanning"><img src="https://github.com/theworker02/centralizer/actions/workflows/codeql.yml/badge.svg" alt="CodeQL"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://github.com/theworker02/centralizer/releases"><img src="https://img.shields.io/github/v/release/theworker02/centralizer?include_prereleases" alt="Latest Release"></a>
  <a href="https://theworker02.github.io/centralizer/"><img src="https://img.shields.io/badge/docs-GitHub%20Pages-c9844a" alt="Docs"></a>
</p>

```go
hub := centralizer.New()
service, err := hub.Connect(ctx, "./analytics")
if err != nil {
    panic(err)
}
result, err := service.Call(ctx, "calculate", centralizer.Args{
    "value": 42,
})
```

The calling application does not need to know which bridge Centralizer selected.

## What Centralizer is

Centralizer is a process-local orchestration layer. You point a Hub at a target directory or an in-process Go handler. The Hub:

1. Discovers language and runtime markers (`go.mod`, `pyproject.toml`, `package.json`, `Cargo.toml`, and the rest of the detect set).
2. Builds a capability graph (stdio, process, native, TCP, …).
3. Plans a bridge under an explicit policy.
4. Starts the selected adapter, converting values through CIR.
5. Supervises the live session: health, bounded recovery, circuit breaker.

It is not a language VM, a package manager, or a general FFI compiler. Language-specific work lives in small adapters and generated shims. Planning, protocol, supervision, and CIR stay in the Go core.

v0.1 is a working vertical slice: a Go host, Python / Node / Rust process targets that tests exercise, CIR, a deterministic planner, a supervisor, and a CLI. Later languages are detect-only until invocation is tested. See [ROADMAP.md](ROADMAP.md).

## Philosophy

Centralizer should never require the developer to manually solve an interoperability problem that Centralizer can reliably determine itself.

That sentence has limits. Automation that cannot be explained, reproduced, or bounded is not useful in a host process you have to debug. The design constraints are therefore:

- **Determinism.** Equivalent discovery input, capability graph, and policy produce the same ranking. Ties break on strategy name.
- **Explainability.** Every selection can be printed. `centralizer explain` and `Service.Explanation()` render the same planner report.
- **Honesty.** A capability is complete when tests demonstrate it. Detect-only adapters stay detect-only. Files existing under `adapters/` do not imply `Call`.
- **Bounds.** Recovery has a restart budget and exponential backoff. After the budget the target is quarantined. Infinite restart loops are a bug.
- **Policy over score.** A higher planner score cannot enable a denied transport, native execution, or subprocess.
- **One core.** Adapters detect and connect. They do not reimplement planning, supervision, or CIR.

When forced to choose, the project prefers stability over more languages, clear behavior over more abstraction, and debuggable automation over automatic magic. See [DESIGN.md](DESIGN.md).

## Architecture

```mermaid
flowchart TD
    A[Application] --> B[Centralizer API]
    B --> C[Capability Graph]
    C --> D[Bridge Planner]
    D --> E[CIR]
    E --> F[Bridge Supervisor]
    F --> G[Adapter]
    G --> H[Target Runtime]
```

Centralizer is a single Go module. Each package has one job.

| Package | Responsibility |
| --- | --- |
| `pkg/centralizer` | Public Hub / Service API |
| `pkg/cir` | Kind-tagged intermediate values |
| `pkg/schema` | Callable surface above CIR; explicit YAML load |
| `pkg/adapter` | Adapter interface, registry, catalog |
| `pkg/bridge` | Plan, Bridge, Stream contracts |
| `pkg/capability` | Capability graph |
| `pkg/health` | Supervisor snapshots |
| `pkg/manifest` | Optional `centralizer.yaml` |
| `pkg/lockfile` | Optional resolved-plan snapshot |
| `pkg/czerr` | Typed errors (`errors.Is` / `errors.As`) |
| `pkg/diagnostics` | `doctor` checks |
| `internal/discovery` | Target analysis and fingerprints |
| `internal/planner` | Deterministic strategy selection |
| `internal/supervisor` | Lifecycle, recovery, circuit breaker |
| `internal/session` | Protocol client and native bridge |
| `internal/protocol` | Framing and messages |
| `internal/transport` | stdio, TCP, Unix, named-pipe client, experimental SHM |
| `internal/security` | Paths, policy, environment filtering |
| `internal/telemetry` | slog, in-process metrics and traces |
| `internal/registry` | Connected service table |
| `internal/cache` | Shim and plan cache |
| `internal/lifecycle` | Handles and shutdown |
| `internal/shim` | Embedded, versioned shim templates |
| `adapters/*` | Language-specific detect / connect only |

```mermaid
flowchart LR
    subgraph public
      API[pkg/centralizer]
    end
    subgraph decide
      D[discovery]
      P[planner]
      C[capability]
    end
    subgraph run
      S[supervisor]
      T[session/transport]
      A[adapter]
    end
    API --> D --> C --> P --> S --> T --> A
```

### Connect path

1. Validate the target path (unless `native:`). Remote `http(s):` and `file:` URLs are rejected in v0.1.
2. Run every adapter's `Detect`. Keep confidence scores; do not pretend certainty.
3. Ask the winning adapter for capabilities. Host facts (OS sockets) do not imply the adapter can connect that way.
4. Plan under policy. Equivalent inputs produce the same ranking.
5. `Prepare` then `Connect`.
6. Load an explicit schema if `schema.yaml` or a manifest `schema:` field is present.
7. Wrap the bridge in a supervisor and register it.

### Ownership

The Hub owns services. A Service owns its supervisor. The supervisor owns the live Bridge. A stdio or TCP transport owns the child process and must kill it on `Close`. Handles never store foreign memory pointers. `Service.Close` drops locally tracked handles for that bridge (`DropBridge`) and reaps children.

See [ARCHITECTURE.md](ARCHITECTURE.md) for subsystem notes.

## Compatibility

Only mark a capability complete when tests demonstrate it. Detect-only languages stay detect-only.

| Runtime | Detect | Call | Stream | Handles | Native | Process | WASM |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Go | Yes | Yes | Planned | Yes | Yes | Yes | Planned |
| Python | Yes | Yes | Yes | Yes | N/A | Yes | N/A |
| Node.js | Yes | Yes | Yes | Planned | N/A | Yes | N/A |
| Rust | Yes | Yes* | Planned | Planned | Planned | Yes | Planned |
| C / C++ | Yes | No | No | No | Planned | Planned | Planned |
| WebAssembly | Yes | No | No | No | N/A | N/A | Planned |
| JVM / .NET / Ruby / PHP | Yes | No | No | No | No | Planned | No |
| Swift / Dart / Lua / Zig | Yes | No | No | No | No | Planned | No |

\* Rust invocation requires a target that speaks [Centralizer Protocol 1.x](PROTOCOL.md) on stdio (`cargo run -- --centralizer` or an existing binary). Centralizer does not compile arbitrary Rust into a shim in v0.1.

`centralizer adapters` prints name, tier, whether invocation is implemented, and claimed capabilities. Foundation adapters (C, C++, WASM, JVM, .NET, Ruby, PHP, Swift, Dart, Lua, Zig) detect only; `Connect` returns not implemented.

## CIR

[CIR](CIR.md) is the semantic boundary between languages. Adapters convert native values into CIR and back. They do not invent a second type system. Values are kind-tagged; they are not stored as an untyped `any` bag.

| Kind | Storage | Notes |
| --- | --- | --- |
| null | — | |
| boolean | `num` 0/1 | |
| int | `num` as int64 bits | overflow checked on convert |
| uint | `num` | |
| float | `num` as float64 bits | non-finite rejected by schema validate |
| decimal | `str` | exact decimal text |
| string | `str` | |
| bytes | `raw` | |
| array / tuple | `items` | |
| map / struct | `keys` + `items` | |
| enum | `str` + discriminant | |
| union | tag + inner | |
| optional | flag + inner | |
| result | ok flag + inner | |
| error | code + message | |
| timestamp | unix nanos | UTC |
| duration | nanos | |
| UUID | 16 bytes | |
| stream | id | |
| handle | id | never a foreign pointer |
| opaque | tag + payload | language-specific |

Optional `Meta` carries type name, language, and schema id.

Wire JSON is kind-tagged. Do not decode CIR by guessing JSON types.

```json
{"k":"int","i":42}
{"k":"map","m":[{"k":"value","v":{"k":"int","i":42}}]}
```

Go uses `cir.From` / `Value.Native`. The Python shim maps `int`→int, `float`→float, `str`→string, `bytes`→bytes, `dict`→map, `list`→array. The Node shim maps integer `number`→int, otherwise float; `Buffer`→bytes; object→map.

Schemas sit above CIR. Inferred schemas do not enforce argument types. An explicit `schema.yaml` (or manifest `schema:`) does, including unknown-function rejection.

## Bridge planner

The planner scores strategies on fixed integer weights (performance, reliability, isolation, startup, serialization, compatibility, security, portability, debuggability, availability). Scores are 0–100. Ties break on strategy name so equivalent inputs produce the same ranking.

Evaluated strategies:

| Strategy | Transport | Typical use |
| --- | --- | --- |
| in-process | `native` | Registered Go handlers (`native:name`) |
| Unix socket | `unix_socket` | Persistent process IPC on non-Windows hosts when the adapter claims the capability |
| named pipe | `named_pipe` | Windows local IPC; client dial exists, server side is experimental |
| stdio | `stdio` | Supervised child, NDJSON protocol — the portable default |
| TCP | `tcp` | Localhost length-prefixed frames when `prefer: [tcp]` and policy allows loopback |
| WASM | `wasm` | Planned invocation; detect-only in v0.1 |
| shared memory | `shared_memory` | Types exist; disabled (`ErrExperimental`) |

How selection works:

1. Discover markers.
2. Score language/runtime hypotheses. Ambiguity is reported, not hidden.
3. Build a capability graph.
4. Score strategies. Policy can reject a candidate; it cannot be overridden by `bench`.
5. Select the highest viable score. Record fallbacks.
6. Explain the decision.

```bash
centralizer explain ./examples/go-python/analytics
```

Manifest `prefer` may boost a listed strategy by a small integer. It cannot enable a policy-denied one. Policy always wins over a higher benchmark score.

`centralizer bench` reports planner scores for viable strategies. It does not start a timed shoot-out that rewrites policy. Published numbers must include machine, OS, versions, transport, and payload size. See [BENCHMARKS.md](BENCHMARKS.md).

## Supervisor, recovery, circuit breaker

External bridges are supervised:

`created → starting → healthy → degraded → recovering → unhealthy → quarantined → stopping → stopped`

On a recoverable error (`ErrBridgeFailed`, `ErrTransportFailure`, `ErrTimeout`) the supervisor:

1. Marks the service recovering.
2. Waits with exponential backoff (default 200 ms, capped at 5 s).
3. Rebuilds the bridge through the adapter factory.
4. If rebuild fails, tries the first recorded fallback, mapping strategy → transport correctly (`in_process` → `native`, not the raw strategy string).
5. Increments the restart counter. Default budget is 5 (`policy.max_restarts`).
6. Quarantines the target when the budget is exhausted.

The circuit breaker trips after consecutive failures (default 5), stays open for 5 s, then admits a half-open probe. Two successes close it. An open breaker returns `ErrCircuitOpen` so a broken dependency cannot livelock the host.

`WithAutoRecovery(false)` disables rebuilds. Native in-process handlers do not tear down a child process; process crash recovery is exercised in `tests/failure`.

Health snapshots include state, transport, latency, success rate, restarts, fallbacks, last error, and breaker state. `Service.Health()` and `Hub.Health()` expose them. `centralizer health <target>` connects, prints one snapshot, and closes.

## Protocol 1.x

Handshake negotiates `protocol: "1.0"`. Major version 1 is required. Unknown message types return `ERROR`. A 1.x peer may ignore unknown optional fields. A 2.x peer must not claim 1.x compatibility without a documented translation.

| Transport | Framing |
| --- | --- |
| stdio | NDJSON, one message per line |
| TCP / Unix | 4-byte big-endian length + JSON body |
| memory (tests) | same as TCP |

Maximum frame size: 16 MiB (`protocol.MaxFrameBytes`).

Envelope:

```json
{"v":1,"id":"corr-id","type":"CALL","payload":{}}
```

`id` is a correlation identifier. Responses reuse it.

| Type | Direction | Purpose |
| --- | --- | --- |
| HELLO | both | version and feature negotiation |
| CAPABILITIES | peer → host | optional capability list |
| DESCRIBE / DESCRIBE_OK | both | schema YAML |
| CALL | host → peer | function or method |
| RESULT | peer → host | CIR value |
| ERROR | either | structured error (`schema`, `conversion`, `handle`, `timeout`, `cancel`, `adapter`, `protocol`, `frame`) |
| STREAM_OPEN / STREAM_DATA / STREAM_CLOSE | both | streams |
| HANDLE_CREATE / HANDLE_RELEASE | both | opaque objects |
| GET / SET | host → peer | properties |
| HEARTBEAT | both | liveness |
| CANCEL | host → peer | abort `id` |
| SHUTDOWN | host → peer | graceful exit |
| OK | peer → host | empty success |

CALL arguments are CIR wire values, not raw JSON types:

```json
{"v":1,"id":"2","type":"CALL","payload":{"function":"calculate","args":{"value":{"k":"int","i":42}}}}
```

Python generators and Node iterables are streamed by the generated shims (`STREAM_OPEN` / `STREAM_DATA` / `STREAM_CLOSE`). Go native streaming is still planned.

Full text: [PROTOCOL.md](PROTOCOL.md), [protocol/v1.md](protocol/v1.md).

## Manifest, lock file, policy

Zero configuration is valid. A manifest exists for deterministic overrides.

`centralizer init` writes `centralizer.yaml` (refuses to overwrite unless `--force`):

```yaml
centralizer:
  version: 1
services:
  analytics:
    source: ./examples/go-python/analytics
    language: auto
    schema: schema.yaml
  engine:
    source: ./examples/go-rust/engine
    language: rust
    prefer:
      - stdio
policy:
  recovery: automatic
  isolation: process
  tracing: true
  native_execution: true
  network: localhost_only
  subprocesses: true
```

Load it in-process with `centralizer.LoadManifest` and `WithManifest`. Service entries can override source, language, entry, prefer list, and schema path.

Policy fields that the engine actually enforces:

| Field | Effect |
| --- | --- |
| `native_execution` | Allows or denies in-process Go handlers |
| `subprocesses` | Allows or denies child processes |
| `generated_code` | Allows or denies shim materialization |
| `network` | `localhost_only` (default), or `none` to reject TCP |
| `allowed_runtimes` | Adapter-name allow-list (empty = all) |
| `allowed_transports` | Transport-name allow-list (empty = all) |
| `max_restarts` | Supervisor budget (default 5) |

`centralizer lock <target> [path]` writes `centralizer.lock`: adapter, transport, strategy, fingerprint, scores, reasons. Connect does not require a lock file. `inspect` reads `centralizer.lock` or `<target>.lock` when present and reports `lock_matches` against the live plan.

## Adapters

Adapters implement detect / capabilities / prepare / connect. They must not implement planning, supervision, or CIR.

```go
type Adapter interface {
    Name() string
    Detect(context.Context, Target) (Detection, error)
    Capabilities(context.Context, Target) ([]capability.Capability, error)
    Prepare(context.Context, Target) error
    Connect(context.Context, Target, bridge.Plan) (bridge.Bridge, error)
}
```

Built-in adapters are registered by `centralizer.New()`. Additional adapters:

```go
hub := centralizer.New(centralizer.WithAdapter(myAdapter))
```

Do not use Go's plugin system as the extension mechanism. A later revision will speak the protocol to out-of-process third-party adapters.

Generated shims (Python, Node) are templates, versioned, hashed into the cache key, and written with restrictive permissions into a private cache. They are not copied into user trees. They implement protocol speak and module loading only.

See [ADAPTERS.md](ADAPTERS.md) and `adapters/*/README.md`.

## Installation

Requires Go 1.23+. The library dependency graph is the Go standard library plus `gopkg.in/yaml.v3`.

Library:

```bash
go get github.com/theworker02/centralizer/pkg/centralizer@latest
```

CLI:

```bash
go install github.com/theworker02/centralizer/cmd/centralizer@latest
```

Optional daemon:

```bash
go install github.com/theworker02/centralizer/cmd/centralizerd@latest
```

From a clone:

```bash
git clone https://github.com/theworker02/centralizer.git
cd centralizer
go test ./...
make build
```

`make build` writes `bin/centralizer` and `bin/centralizerd` with the module version via `-ldflags`.

Brand assets (SVG/PNG) for sites and npm consumers:

```bash
npm install @theworker02/centralizer-brand
```

The package is the official mark only. It is not a JavaScript runtime. Publishing to the npm registry is a separate step; this repository ships the package sources under `packages/centralizer-brand/`.

## Quickstart

```go
package main

import (
    "context"
    "fmt"

    "github.com/theworker02/centralizer/pkg/centralizer"
)

func main() {
    ctx := context.Background()
    hub := centralizer.New()
    analytics, err := hub.Connect(ctx, "./analytics")
    if err != nil {
        panic(err)
    }
    defer analytics.Close(ctx)

    result, err := analytics.Call(ctx, "calculate", centralizer.Args{
        "value": 42,
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(result)
}
```

Point Centralizer at `examples/go-python/analytics` to run this against a real Python module (CPython must be on `PATH`).

CLI equivalent:

```bash
centralizer detect ./examples/go-python/analytics
centralizer explain ./examples/go-python/analytics
centralizer call ./examples/go-python/analytics calculate value=21
```

`call` coerces `k=v` tokens: `true`/`false`, integers, floats, otherwise strings.

## Advanced API

```go
hub := centralizer.New(
    centralizer.WithAutoRecovery(true),
    centralizer.WithTracing(true),
    centralizer.WithTimeout(30*time.Second),
    centralizer.WithHandleTTL(time.Hour),
    centralizer.WithCacheDir(""), // default user cache
    centralizer.WithLogger(slog.Default()),
    centralizer.WithLanguage("auto"),
    centralizer.WithPrefer("stdio"),
    centralizer.WithManifest(m),
    centralizer.WithPolicy(m.Policy),
    centralizer.WithAdapter(myAdapter),
)

hub.RegisterNative(&session.Handler{
    Name: "math",
    Funcs: map[string]session.Func{
        "calculate": func(_ context.Context, args map[string]cir.Value) (cir.Value, error) {
            n, err := args["value"].Int()
            if err != nil {
                return cir.Value{}, err
            }
            return cir.Int(n * 2), nil
        },
    },
})

svc, err := hub.Connect(ctx, "native:math")
_ = svc.Language()
_ = svc.Runtime()
_ = svc.Transport()
_ = svc.Plan()
_ = svc.Health()
_ = svc.Explanation()
_ = svc.Capabilities()

result, err := svc.Call(ctx, "calculate", centralizer.Args{"value": 21})
result, err = svc.Invoke(ctx, bridge.Invocation{Function: "calculate", Args: cirArgs})

st, err := svc.Stream(ctx, "count_up", centralizer.Args{"n": 4})
sub, err := svc.Subscribe(ctx, "events")

h, err := svc.New(ctx, "Session", nil)
id, err := h.HandleID()
_ = svc.Get(ctx, id, "name")
_ = svc.Set(ctx, id, "name", "demo")
_ = svc.Release(ctx, id)

sc, err := svc.Describe(ctx)
text, plan, err := hub.Explain(ctx, "./analytics")
analysis, err := hub.Analyze(ctx, "./analytics")
lf, err := hub.LockPlan(ctx, "./analytics")
_ = hub.Adapters()
_ = hub.AdapterCatalog()
_ = hub.List()
_ = hub.Health()
_ = hub.Cache()
_ = hub.Close(ctx)
```

Public APIs do not panic for ordinary failures. `Call` converts `Args` through CIR and, when an explicit schema is loaded, validates function name and argument types. `WithTimeout` applies only when the context has no deadline. `WithHandleTTL` expires locally tracked handles; unknown ids are still forwarded to the remote.

Connect options can be passed per call: `hub.Connect(ctx, ref, centralizer.WithEntry("model.py"), centralizer.WithLanguage("python"))`.

## CLI

```text
centralizer detect <target>
centralizer inspect <target>
centralizer describe <target>
centralizer connect <target>
centralizer call <target> <function> [k=v...]
centralizer health <target>
centralizer list
centralizer graph <target>
centralizer explain <target>
centralizer bench <target>
centralizer trace <target>
centralizer doctor
centralizer cache list|clear
centralizer init [--force] [path]
centralizer lock <target> [path]
centralizer adapters
centralizer version
```

| Command | Behavior |
| --- | --- |
| `detect` | Score language/runtime hypotheses without connecting |
| `inspect` | Analysis + plan; includes lock comparison when a lock file exists |
| `describe` | Connect, print inferred or declared schema |
| `connect` | Establish a bridge and print health, then close |
| `call` | Connect, invoke, print the native result |
| `health` | Connect (or report that this process has no long-lived services) |
| `list` | Registered adapter names |
| `graph` | Capability graph summary |
| `explain` | Human-readable planner report |
| `bench` | Planner scores for viable strategies; does not override policy |
| `trace` | Connect with in-process spans; print elapsed / adapter / transport |
| `doctor` | Host toolchain, cache writability, Git, protocol version, adapters |
| `cache` | List or clear generated shim artifacts |
| `init` | Write a starter `centralizer.yaml` |
| `lock` | Write a resolved-plan `centralizer.lock` |
| `adapters` | Name, tier, invocation flag, claimed capabilities |
| `version` | SemVer and protocol (`centralizer 0.1.1 (protocol 1.0)`) |

`--json`, `--quiet` / `-q`, and `--verbose` / `-v` apply to every command. `--json` plus `--verbose` emits machine-readable slog lines.

Each CLI invocation constructs a new Hub. `list` and `health` without a long-lived process therefore cannot see services started by another command. Use the library or `centralizerd` if you need a process-wide table.

## Examples

| Path | What it demonstrates |
| --- | --- |
| `examples/go-python` | Discovery + CIR call into CPython (`calculate`) |
| `examples/go-node` | Discovery + CIR call into Node.js (`report`) |
| `examples/go-rust` | Protocol-speaking Rust engine (`multiply`) |
| `examples/go-c` / `examples/go-cpp` | Detection only (invocation not implemented) |
| `examples/go-wasm` | Detection notes for `.wasm` |
| `examples/showcase` | One Hub across Python, Rust, and Node |
| `examples/streaming` | Python generator via `STREAM_*` |
| `examples/recovery` | Health / error surface on a native handler |
| `examples/object-handles` | Python `HANDLE_CREATE` / `Release` |

Run from the example directory or the repository root. Python and Node must be on `PATH`. Rust examples need `cargo`. See [examples/README.md](examples/README.md).

## Security and trust

Discovered code is not trusted merely because Centralizer found it. A target may be an unreviewed project, a shared library, or a local service. Centralizer's job is to reach it under an explicit policy, not to attest that it is safe.

The host application is the trust boundary. Centralizer will start child processes, generate disposable shims into a private cache, and exchange protocol frames with those processes. Treat connected targets as you would any local executable you chose to run.

Controls that exist in this tree:

- Path validation (`ResolveTarget`): null bytes rejected; `http(s):` and `file:` rejected; target must exist after `Abs`+`Clean`.
- Environment filtering on child processes (`LD_PRELOAD`, `LD_LIBRARY_PATH`, `DYLD_*`, `PYTHONINSPECT`, `NODE_OPTIONS`).
- Payload size limit (16 MiB frames).
- Handle validation and optional TTL; handles are ids, not foreign pointers.
- Policy allow-lists for runtimes and transports.
- `centralizerd` binds loopback and rejects non-loopback peers.

Report unpatched vulnerabilities privately via GitHub Security Advisories. Do not open a public issue. See [SECURITY.md](SECURITY.md).

## Observability

### Doctor

`centralizer doctor` inspects the host: Go toolchain, Python, Node, Cargo, OS sockets, shared-memory warning (experimental / disabled), cache directory writability, Git on `PATH`, protocol version parse, and registered adapter names.

### Telemetry

`internal/telemetry.DefaultMetrics` records calls, errors, timeouts, bytes, bridge restarts, streams, runtime startups, planner decisions, and a bounded latency sample. `Snap()` exports JSON-friendly p50/p99.

There is no required OpenTelemetry dependency. A host may scrape `Snap()` into its own Meter. `WithTracing(true)` records an in-process span tree used by `centralizer trace`; that tracer is not an OTLP exporter.

Structured logs use `log/slog`. `WithLogger` replaces the default. See [docs/telemetry.md](docs/telemetry.md).

## Website

Documentation site sources live in [`website/`](website/) (Vite, React, TypeScript). Vite `base` is `./`, so the same `dist/` works locally and on GitHub project Pages.

Published site: [theworker02.github.io/centralizer](https://theworker02.github.io/centralizer/)

```bash
cd website
npm install
npm run dev
```

`make website` runs the production build. Enable Pages with **Settings → Pages → Source: GitHub Actions** (`.github/workflows/pages.yml`). There is no custom domain / CNAME.

The site copies `assets/logo.svg` and `assets/icon.svg` into `website/public` at build start so the hero, nav, and favicon cannot drift from the canonical mark.

## Brand assets

`assets/` is the single source of truth:

| File | Use |
| --- | --- |
| `assets/logo.svg` | Canonical mark (hexagonal C, copper hub, unified exit) |
| `assets/logo-dark.svg` | Dark-background variant |
| `assets/logo-light.svg` | Light-background variant |
| `assets/icon.svg` | Compact icon / favicon |
| `assets/icon-256.png`, `assets/icon-512.png` | Raster icons |
| `assets/github-social-preview.svg` | Vector source for the social preview |
| `assets/github-social-preview.png` | GitHub social preview (1280×640) |

README, the documentation site, and `@theworker02/centralizer-brand` all use these files. Do not introduce a one-off drawing. Upload `assets/github-social-preview.png` in repository settings for the social preview. See [docs/github-metadata.md](docs/github-metadata.md).

## centralizerd

`centralizerd` is optional. Ordinary library use does not start a daemon. `centralizer.New()` never launches it.

When run, it listens on loopback only (default `127.0.0.1:4780`) and exposes:

| Path | Purpose |
| --- | --- |
| `GET /healthz` | Version and protocol |
| `GET /v1/adapters` | Registered adapter names |
| `GET /v1/services` | In-process service table |
| `GET /v1/doctor` | Doctor report |
| `GET /v1/metrics` | `DefaultMetrics.Snap()` JSON |
| `POST /v1/register` | Connect a source into this process |

Do not bind it to a public interface. Non-loopback peers receive 403.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Adapter authors are welcome; the public `adapter.Adapter` interface is the extension point.

```bash
go test ./...
gofmt -w .
make lint
```

Use conventional commits (`feat`, `fix`, `docs`, `test`, `chore`, `refactor`). Do not mark a language or transport complete unless CI exercises it. Do not add dependencies unless the standard library is insufficient. Participants follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

## Roadmap

v0.1 is a working vertical slice: Go host, Python / Node / Rust process targets, CIR, planner, supervisor, and CLI. Streaming shims, lock files, explicit schemas, and localhost TCP are available in this tree. Later phases add C/WASM invocation, JVM/.NET, shared memory, and a multi-process daemon registry.

Do not treat later phases as implemented because files exist. See [ROADMAP.md](ROADMAP.md) and [CHANGELOG.md](CHANGELOG.md).
