# Evaluator Guide — Centralizer (first hour)

**Audience:** acquirers, technical diligence engineers, and integration evaluators.  
**Date:** 2026-09-21  
**Contact:** [@theworker02](https://github.com/theworker02)

This guide gets you from clone to an honest read of what Call works today—without treating detect-only adapters as finished.

## Minute 0–10: build and host check

```bash
git clone https://github.com/theworker02/centralizer.git && cd centralizer
make build
./bin/centralizer version
./bin/centralizer doctor
./bin/centralizer adapters
```

What to notice:

- `doctor` reports Go / Python / Node / Cargo presence, cache writability, and a **call-capable vs detect-only** split.
- `adapters` lists `call` as `yes` or `no` per adapter. Trust this column over folder names under `adapters/`.

## Minute 10–25: explain before you call

```bash
./bin/centralizer explain ./examples/go-python/analytics
./bin/centralizer explain ./examples/go-node/reporter
./bin/centralizer --json explain ./examples/go-python/analytics | head
```

Each report includes:

| Field | Meaning |
| --- | --- |
| Detection score | How strongly the tree matched a language |
| Adapter | Chosen adapter name |
| Call implemented | Catalog truth — **not** inferred from Detect |
| Bridge plan | Strategy / transport / score when planning succeeds |
| Next steps | Concrete CLI / example commands |

If Call is `no`, stop. Detection confidence alone is not a ship claim.

## Minute 25–40: live Call path

Requires `python3`/`python` and `node` on `PATH`:

```bash
./bin/centralizer call ./examples/go-python/analytics calculate value=21
./bin/centralizer call ./examples/go-node/reporter report value=21
go run ./examples/polyglot-hello
```

`polyglot-hello` is the smallest multi-language demo (Python + Node, no Rust). For Rust as well, see `examples/showcase` (needs `cargo`).

## Minute 40–55: negative control (detect-only)

```bash
echo "print('hi')" > /tmp/cz-hello.lua
./bin/centralizer detect /tmp/cz-hello.lua
./bin/centralizer explain /tmp/cz-hello.lua
```

You should see Lua detection with **`Call implemented: no (detect-only)`**.  
`Connect` / `Call` against that target must fail with not-implemented — that is correct behavior.

## Minute 55–60: diligence pointers

| Doc | Use |
| --- | --- |
| [ACQUISITION.md](../ACQUISITION.md) | Strategic brief (no valuation) |
| [acquisition/BUYER_DEMO.md](acquisition/BUYER_DEMO.md) | Short command script |
| [acquisition/CENTRALIZER_DILIGENCE.md](acquisition/CENTRALIZER_DILIGENCE.md) | Project diligence |
| [acquisition/KNOWN_LIMITATIONS.md](acquisition/KNOWN_LIMITATIONS.md) | Limits |
| [acquisition/LICENSE_TRANSITION_ANALYSIS.md](acquisition/LICENSE_TRANSITION_ANALYSIS.md) | License transition |
| [ADAPTERS.md](../ADAPTERS.md) / README compatibility matrix | Capability claims |

## Decision heuristics

- **Proceed with integration pilot** if Python/Node (and optionally Rust protocol) Call paths work on your machines and the license transition is acceptable to counsel.
- **Do not** price or roadmap “all languages under `adapters/`” as Call-complete.
- **Do** use `explain` + `adapters` as the shared vocabulary between product and diligence.

No valuation is provided in this guide.
