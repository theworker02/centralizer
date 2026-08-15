# Telemetry

Centralizer ships process-local counters and an optional OpenTelemetry hook.
There is no required telemetry dependency.

## In-process metrics

`internal/telemetry.DefaultMetrics` records calls, errors, timeouts, bytes,
bridge restarts, streams, runtime startups, planner decisions, and a bounded
latency sample. `Snap()` exports JSON-friendly percentiles.

`centralizerd` exposes the same snapshot on loopback:

```text
GET http://127.0.0.1:4780/v1/metrics
```

The daemon binds localhost only. Do not expose it on a public interface.

## OpenTelemetry

The host application may scrape `telemetry.DefaultMetrics.Snap()` from its own
Meter or Tracer. Centralizer does not import the OpenTelemetry SDK so library
users keep a stdlib-only dependency graph.

Suggested mapping:

| Centralizer field | OTel instrument |
| --- | --- |
| `calls` | counter `centralizer.calls` |
| `errors` | counter `centralizer.errors` |
| `bridge_restarts` | counter `centralizer.bridge.restarts` |
| `runtime_startups` | counter `centralizer.runtime.startups` |
| `planner_decisions` | counter `centralizer.planner.decisions` |
| `latency_p50_ms` / `latency_p99_ms` | histogram `centralizer.call.duration` |

Attach spans around `Hub.Connect` with `centralizer.WithTracing(true)` for an
in-process span tree used by `centralizer trace`. That tracer is not an OTLP
exporter.

## Logs

Structured logs use `log/slog`. `centralizer.WithLogger` replaces the default
logger. `--verbose` on the CLI raises the level; `--json` emits machine-readable
lines when verbose.
