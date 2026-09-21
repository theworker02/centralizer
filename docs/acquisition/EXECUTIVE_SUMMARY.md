# Executive Summary — Centralizer

**Date:** 2026-09-21  
**Current license:** centralizer Source-Available Evaluation License (proprietary source-available)  
**Prior license (historical distributions):** Apache License, Version 2.0 (Apache-2.0)  
**Transition marker:** 41c8fbb / merge 31efeca (2026-09-20)

## What this is

Interop runtime that discovers, connects, supervises, and exposes software across languages via a unified protocol/hub.

## Problem addressed

Polyglot systems lack a supervised, local-first bridge with a documented protocol and adapter matrix.

## Maturity

Early (v0.1.x). Tier-1 call path for Go/Python/Node/Rust protocol speakers; many languages detect-only.

## Deployment model

`make build` → bin/centralizer, bin/centralizerd; optional website; adapters need host runtimes.

## Language / stack

Go 1.23+; website React/Vite; multi-language adapters · Version metadata: **0.1.2 (ldflags; tip post-relicense)**

## Licensing posture (factual)

- Current tree: proprietary / source-available terms in root `LICENSE` (see exact text).
- Historical distributions under **Apache License, Version 2.0 (Apache-2.0)** remain governed by those terms for copies received, where applicable.
- See [`LICENSE_TRANSITION_ANALYSIS.md`](./LICENSE_TRANSITION_ANALYSIS.md) and root `LICENSE_TRANSITION_NOTICE.md`.

## Ownership (asserted, not adjudicated)

Asserted holder: **theworker02 (https://github.com/theworker02)**.  
LICENSE: theworker02. NOTICE previously said 'The Centralizer Authors' + Apache (fixed this program). No CLA/DCO. Dependabot commits present.

**REQUIRES_LEGAL_REVIEW** before treating ownership as adjudicated or exclusive.

## What a buyer can expect

- Ability to evaluate and (after commercial license / acquisition) operate the project with documented handoff materials in this data room.
- Material third-party and historical-license limitations disclosed herein.
- No fabricated users, revenue, benchmarks, or exclusivity claims in this data room.

## Top diligence risks

- Apache→proprietary + proxy.golang.org v0.1.2
- Protocol IP packaging
- Adapter completeness honesty
