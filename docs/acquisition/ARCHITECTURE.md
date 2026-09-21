# Architecture — Centralizer

See also repository root architecture docs where present (`ARCHITECTURE.md`, `docs/`, `README.md`).

## Stack

Go 1.23+; website React/Vite; multi-language adapters

## Deployment

`make build` → bin/centralizer, bin/centralizerd; optional website; adapters need host runtimes.

## Summary

Interop runtime that discovers, connects, supervises, and exposes software across languages via a unified protocol/hub.

## Boundaries

- Third-party runtimes, cloud providers, and SDKs are **dependencies**, not owned assets.
- Project-specific diligence: [`CENTRALIZER_DILIGENCE.md`](./CENTRALIZER_DILIGENCE.md).

Buyer should walk architecture with the handoff plan ([HANDOFF_PLAN.md](./HANDOFF_PLAN.md)).
