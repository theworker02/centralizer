# Benchmarks

Do not publish a latency or throughput number without the following fields.

| Field | Example |
| --- | --- |
| Machine | hostname or cloud instance type |
| CPU | model and count |
| OS | `windows 10.0.26200` |
| Runtime version | CPython 3.13 / Node 22 / rustc 1.80 |
| Centralizer version | 0.1.0 |
| Test | `go test -bench=BenchmarkCIREncode` |
| Sample size | `b.N` or N calls |
| Transport | stdio / unix / native |
| Payload size | CIR map with 3 keys |

`centralizer bench <target>` prints planner scores for **viable** strategies. It is not a wall-clock measurement and it does not override policy.

## Built-in microbenchmarks

```bash
go test -bench=. -benchmem ./benchmarks ./pkg/cir ./internal/protocol
```

These measure CIR encode/decode and protocol framing on the machine that runs them. They are not a claim about cross-language call latency.

## Cross-language calls

Record real `Call` latency with `Service.Health().Latency` after a warmup. Compare only identical payload shapes. Native Go calls are a baseline, not a target other languages are expected to match.
