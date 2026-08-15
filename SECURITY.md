# Security policy

## Trust model

Discovered code is not trusted merely because Centralizer found it. A target may be an unreviewed project, a shared library, or a local service. Centralizer's job is to reach it under an explicit policy, not to attest that it is safe.

The host application is the trust boundary. Centralizer will start child processes, generate disposable shims into a private cache, and exchange protocol frames with those processes. Treat connected targets as you would any local executable you chose to run.

## Reporting a vulnerability

Please **do not** open a public issue for unpatched vulnerabilities.

Report privately via GitHub Security Advisories on this repository, or email the maintainers listed in the latest release notes.

Include:

- Centralizer version and commit
- Go version, OS, architecture
- Adapter and transport
- Impact (protocol, command injection, sandbox escape, path traversal, unsafe generated shims, resource exhaustion, adapter-specific bugs)
- A minimal reproduction if you can share one safely

We will acknowledge the report, work on a fix, and credit you if you want to be named.

## Scope

In scope:

- Protocol framing and length limits
- Command injection in adapter argv construction
- Path traversal in target resolution or cache writes
- Sandbox / isolation escapes in documented isolation modes
- Unsafe generated shim permissions
- Resource exhaustion that the supervisor should bound
- Handle validation bypasses
- `centralizerd` `/v1/metrics` if reachable beyond loopback
- Adapter vulnerabilities in this repository

Out of scope:

- Bugs in third-party language runtimes
- Targets that are themselves malicious once intentionally connected
- Denial of service by connecting to a target that is designed to hang, unless Centralizer fails to honor `context.Context`

## Supported versions

Security fixes target the latest `0.x` release until 1.0. The public API and protocol are not stable before 1.0.
