# Acquisition Brief — Centralizer

**Date:** 2026-09-21  
**Contact:** GitHub [@theworker02](https://github.com/theworker02)  
**Status:** Diligence briefing only. **No acquisition has occurred** by virtue of this file.  
**Valuation:** Not stated.

---

## Strategic thesis

Polyglot systems keep accumulating ad-hoc bridges (FFI wrappers, bespoke stdio glue, one-off RPC). Centralizer is a **local-first interop runtime**: one Go hub discovers a target tree, plans a supervised bridge, and exposes a typed call surface (CIR + Protocol 1.x) so a host can invoke foreign code without pretending every language adapter is equally finished.

The acquisition value is the combination of:

1. A working vertical slice (Go host → Python / Node / Rust Call paths).
2. An honest adapter matrix (detect ≠ Call).
3. A documented protocol, planner, supervisor, and CLI that evaluators can exercise in the first hour.
4. A packaged diligence data room under `docs/acquisition/`.

This is early product (v0.1.x), not a mature platform claim.

---

## What the project does

Interop runtime that **discovers**, **connects**, **supervises**, and **exposes** software across languages via a unified protocol/hub.

Core loop:

1. `Detect` — score language/runtime hypotheses.
2. Plan — pick stdio / TCP / in-process / … with explainable scores.
3. `Connect` — materialize a bridge (generated shim or protocol speaker).
4. `Call` — invoke through CIR (when the adapter implements invocation).

Primary entry points: Go library (`pkg/centralizer`), CLI (`cmd/centralizer`), optional loopback daemon (`cmd/centralizerd`).

---

## Problem

Polyglot codebases lack a supervised, local-first bridge with:

- a documented wire protocol,
- deterministic bridge selection that can be printed (`centralizer explain`),
- and an explicit compatibility matrix that does not inflate detect-only adapters into Call claims.

---

## What ships today (v0.1.x)

| Capability | Status |
| --- | --- |
| Go Hub API (`Connect`, `Call`, handles, stream where shim supports it) | Ships |
| CLI: detect, inspect, explain, connect, call, doctor, adapters, lock, … | Ships |
| Python generated stdio/TCP shim | Call yes |
| Node generated stdio/TCP shim | Call yes |
| Go in-process handlers + protocol binaries | Call yes |
| Rust stdio Protocol 1.x speakers | Call yes (if binary speaks protocol) |
| Discovery for C, C++, WASM, JVM, .NET, Ruby, PHP, Swift, Dart, Lua, Zig | Detect only |
| `centralizerd` loopback registry | Optional / early |
| Diligence data room `docs/acquisition/` | Ships |

Build: `make build` → `bin/centralizer`, `bin/centralizerd`. Adapters need host runtimes on `PATH` (Python, Node, Cargo as applicable).

Evaluator demo path: `docs/acquisition/BUYER_DEMO.md`, `docs/EVALUATOR_GUIDE.md`, `examples/polyglot-hello`.

---

## Detect vs Call (honesty)

**Detection success is not Call readiness.**

- `centralizer detect` / discovery confidence scores mean “this tree looks like language X.”
- `Call implemented` is taken from the adapter catalog (`centralizer adapters`, `centralizer explain`).
- Foundation adapters (Lua, JVM, …) may Detect and still return `ErrNotImplemented` on `Connect` / `Call`.
- `centralizer explain <target>` prints detection score, chosen adapter, **whether Call is implemented**, planner selection, and next steps.
- `centralizer doctor` summarizes call-capable vs detect-only adapters on the host.

Do not mark detect-only adapters as Call-capable in marketing, diligence, or UI.

---

## Included / excluded (typical transaction framing)

**Included (typical):**

- This repository and original Go code authored for Centralizer.
- Protocol documentation and schema.
- Adapters and shim templates in-tree.
- Brand assets created for the project (`assets/`, brand package).
- Acquisition / diligence materials under `docs/acquisition/`.

**Excluded (typical):**

- Historical Apache-2.0 grants on prior tags / distributions (see license transition docs).
- Third-party dependency code (Go modules, runtimes, OS toolchains).
- Buyer infrastructure, secrets, cloud accounts.
- Unverified trademark registrations.
- Any implied user base, revenue, or exclusivity (none claimed here).

Exact transfer scope is a commercial/legal negotiation. See `docs/acquisition/TRANSFER_PLAN.md` and `TRANSFER_MANIFEST.md`.

---

## Maturity and differentiation

- **Maturity:** Early (v0.1.x). Suitable for evaluation and integration pilots; not a claim of production hardening across all adapters.
- **Deployment:** Library-in-process by default; optional localhost daemon; ephemeral CI/dev filesystem assumptions apply where relevant.
- **Differentiation:** Explainable planner + supervised stdio default + explicit detect-vs-call matrix + local-only privacy posture (no author telemetry).

---

## Diligence pointers

Start here:

| Document | Why |
| --- | --- |
| [`docs/acquisition/`](docs/acquisition/) | Full data room index |
| [`docs/acquisition/CENTRALIZER_DILIGENCE.md`](docs/acquisition/CENTRALIZER_DILIGENCE.md) | Project-specific diligence |
| [`docs/acquisition/EXECUTIVE_SUMMARY.md`](docs/acquisition/EXECUTIVE_SUMMARY.md) | Buyer overview |
| [`docs/acquisition/LICENSE_TRANSITION_ANALYSIS.md`](docs/acquisition/LICENSE_TRANSITION_ANALYSIS.md) | Apache → source-available transition |
| [`docs/acquisition/KNOWN_LIMITATIONS.md`](docs/acquisition/KNOWN_LIMITATIONS.md) | Honest limits |
| [`docs/acquisition/BUYER_DEMO.md`](docs/acquisition/BUYER_DEMO.md) | Reproducible demo (includes `explain`) |
| [`docs/EVALUATOR_GUIDE.md`](docs/EVALUATOR_GUIDE.md) | First-hour evaluator guide |
| [`LICENSE_TRANSITION_NOTICE.md`](LICENSE_TRANSITION_NOTICE.md) | Root transition notice |
| [`COMMERCIAL.md`](COMMERCIAL.md) | Commercial contact framing |

---

## Contact

GitHub [@theworker02](https://github.com/theworker02)

No valuation is offered in this brief or in the data room.
