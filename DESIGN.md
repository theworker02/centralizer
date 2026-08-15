# Design

## Principles

When forced to choose:

- stability over more languages
- clear behavior over more abstraction
- debuggable automation over automatic magic

Automation must remain deterministic, observable, and bounded.

## Why Go

The host runtime is Go so that orchestration, planning, protocol, and supervision live in one place. Language shims stay small. They must not reimplement Centralizer.

## CIR storage

`cir.Value` is kind-tagged. Scalars use `num` / `str` / `raw`. Composites use `items` and `keys`. This avoids boxing every integer in an `any`. Reflection is used only at the Go-native conversion boundary.

## Protocol

Version 1 uses JSON payloads. Stdio uses NDJSON so Python and Node shims stay tiny. Sockets use a 4-byte big-endian length prefix and the same JSON body. Major versions must match; minor versions may differ. Maximum frame size is 16 MiB.

## Planning

Scores are integers 0–100 with fixed weights. Ties break on strategy name. Manifest `prefer` may boost a strategy; it cannot enable a policy-denied one. `centralizer bench` reports scores and does not rewrite policy.

## Recovery

Retries are budgeted. Backoff is exponential and capped. After the budget, the target is quarantined. The circuit breaker opens after consecutive failures so a broken dependency cannot livelock the host.

## Shims

Generated shims are templates, versioned, hashed into the cache key, and written with restrictive permissions. They are not copied into user trees. They implement protocol speak and module loading only.

## Shared memory

The types exist. The transport is experimental and returns `ErrExperimental` unless explicitly enabled. Do not use it in production in v0.1.

## Daemon

`centralizerd` binds loopback only. Library `New()` never starts it.
