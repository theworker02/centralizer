# Examples

| Path | What it demonstrates |
| --- | --- |
| `go-python` | Discovery + CIR call into CPython |
| `go-node` | Discovery + CIR call into Node.js |
| `go-rust` | Protocol-speaking Rust engine |
| `go-c` / `go-cpp` | Detection only (invocation not implemented) |
| `go-wasm` | Detection notes for `.wasm` |
| `polyglot-hello` | Go host explains + calls Python and Node (no Rust) |
| `showcase` | Go orchestrator across Python, Rust, and Node |
| `streaming` | Python generator via STREAM_* |
| `recovery` | Health / error surface |
| `object-handles` | Python `HANDLE_CREATE` |

Run from the example directory or the repository root. Python and Node must be on `PATH`. Rust examples need `cargo`.

Buyer/evaluator path: `centralizer explain <target>` then `examples/polyglot-hello` — see [docs/EVALUATOR_GUIDE.md](../docs/EVALUATOR_GUIDE.md).
