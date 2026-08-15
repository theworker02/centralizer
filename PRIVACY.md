# Privacy policy

Last updated: 2026-08-15

This document describes what Centralizer does and does not collect. It is written for the software in this repository (`github.com/theworker02/centralizer`), not for a hosted product.

Centralizer is a **local library and CLI**. It does not require an account. There is no Centralizer cloud, no sign-in, and no user database operated by the authors.

## Summary

| Question | Answer |
| --- | --- |
| Do I need an account? | No |
| Does Centralizer phone home? | No. Nothing is shipped to the authors by default |
| Is there analytics or telemetry to the project? | No remote analytics. Optional in-process counters stay on the machine |
| Where do doctor reports and the cache live? | On the local machine |
| What about `centralizerd`? | Optional, binds loopback only |
| What happens when I connect a target? | That code runs locally. You are responsible for it |
| Third parties? | GitHub, pkg.go.dev, and `proxy.golang.org` if you use those services |

## What Centralizer is

Centralizer discovers local software, plans a bridge, and (when an adapter implements invocation) starts or attaches to a target on the same machine. The host process is the trust boundary. Ordinary library use (`centralizer.New()`, `Hub.Connect`) does not open a network connection to the authors and does not start `centralizerd`.

## What is not collected

The distributed library, CLI, and optional daemon do **not**:

- Require registration, login, or an API key issued by the authors
- Send usage analytics, crash reports, or feature flags to the authors
- Embed a remote telemetry exporter (no OTLP, no SaaS metrics backend)
- Upload doctor output, planner explanations, lock files, or cache contents
- Scan the filesystem for the purpose of reporting it elsewhere

If a future release adds optional remote telemetry, it will be off by default and documented here and in [docs/telemetry.md](docs/telemetry.md).

## What stays on the machine

### Doctor

`centralizer doctor` inspects the local host: toolchain presence (Go, Python, Node, Cargo), OS socket support, cache writability, Git on `PATH`, protocol version, and registered adapter names. The report is printed to the terminal (or returned as JSON on stdout / on `centralizerd` `GET /v1/doctor`). It is not transmitted to the authors.

### Cache

Generated shims, plan snapshots, and related artifacts are written under the user cache directory (`os.UserCacheDir()` + `centralizer`), unless you set `WithCacheDir`. Typical locations:

- Windows: `%LocalAppData%\centralizer`
- macOS: `~/Library/Caches/centralizer`
- Linux: `$XDG_CACHE_HOME/centralizer` or `~/.cache/centralizer`

Cache files are created with restrictive permissions (directory `0700`, files `0600`). They are disposable. `centralizer cache list` and `centralizer cache clear` operate only on this local store.

### Logs and in-process metrics

Structured logs use `log/slog` and go where the host process sends them (usually stderr). `internal/telemetry.DefaultMetrics` records calls, errors, timeouts, bytes, restarts, and a bounded latency sample **in memory**. `Snap()` and `centralizerd` `GET /v1/metrics` export that snapshot locally. `WithTracing(true)` records an in-process span tree for `centralizer trace`. That tracer is not an OTLP exporter and does not leave the process unless you scrape it yourself.

### Manifest, lock, and schema files

`centralizer.yaml`, `centralizer.lock`, and `schema.yaml` are ordinary files you create or that the CLI writes next to your project. Centralizer does not upload them.

## Optional `centralizerd`

`centralizerd` is optional. `centralizer.New()` never launches it.

When you run it, it listens on **loopback only** (default `127.0.0.1:4780`). Non-loopback peers receive HTTP 403. Endpoints (`/healthz`, `/v1/adapters`, `/v1/services`, `/v1/doctor`, `/v1/metrics`, `/v1/register`) serve data from that process to localhost clients. Do not bind it to a public interface. If you do, you are exposing local service metadata and metrics; that is outside the intended deployment.

## Connecting a target

When you `Connect` or `centralizer call` a path, Centralizer may:

- Start a child process (Python, Node, Rust, …)
- Generate a disposable shim into the local cache
- Exchange protocol frames with that process
- Load modules or binaries from the target tree

That code runs **on your machine, under your user**. Centralizer does not attest that the target is safe. You are responsible for what you point it at, the same way you are responsible for any local executable you choose to run. See [SECURITY.md](SECURITY.md).

Environment variables that commonly inject code into child processes (`LD_PRELOAD`, `LD_LIBRARY_PATH`, `DYLD_*`, `PYTHONINSPECT`, `NODE_OPTIONS`) are filtered. That is a local hardening control, not a data-collection feature.

## Third-party services

Centralizer itself does not call these. You may, depending on how you obtain or host the project:

| Service | When it appears | What they may see |
| --- | --- | --- |
| [GitHub](https://github.com/theworker02/centralizer) | Clone, issues, Actions, Pages, releases | Git traffic, issue text, workflow logs, the public site |
| [GitHub Actions](https://docs.github.com/en/actions) | CI on this repository | Repository contents and workflow output on GitHub-hosted runners |
| [pkg.go.dev](https://pkg.go.dev/github.com/theworker02/centralizer) | Browsing or linking Go docs | Your request to Google’s module site |
| [proxy.golang.org](https://proxy.golang.org) | `go get`, `go install`, `GOPROXY` default | Module path and version you request |
| [sum.golang.org](https://sum.golang.org) | Default Go checksum database | Module path and version |

Those operators have their own privacy policies. Using GitHub or the public Go module proxy is optional; you can vendor the module or set `GOPROXY=off` / a private proxy.

The documentation site at [theworker02.github.io/centralizer](https://theworker02.github.io/centralizer/) is static files on GitHub Pages. This project does not add a separate analytics script. GitHub Pages may log requests under GitHub’s terms.

The brand package `@theworker02/centralizer-brand` is source in this repository. If it is later published to the npm registry, npm is a third party for that install path.

## Children and accounts

Centralizer is developer tooling. It does not target children, does not create profiles, and does not process payment information.

## Changes

Material changes to this policy will be recorded in [CHANGELOG.md](CHANGELOG.md) and the date at the top of this file.

## Contact

- General questions and documentation defects: [GitHub Issues](https://github.com/theworker02/centralizer/issues)
- Design discussion: [GitHub Discussions](https://github.com/theworker02/centralizer/discussions)
- Security vulnerabilities: **do not** open a public issue. Use [GitHub Security Advisories](https://github.com/theworker02/centralizer/security/advisories) as described in [SECURITY.md](SECURITY.md)

There is no separate privacy email. The issue tracker and security advisory flow are the contact methods for this project.
