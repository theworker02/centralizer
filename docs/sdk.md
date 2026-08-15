# Adapter SDK

See [ADAPTERS.md](../ADAPTERS.md) for the interface, registration, lifecycle, CIR conversion, and testing expectations.

Helper packages:

- `pkg/adapter` — `Adapter`, `Registry`, `DetectAll`
- `pkg/cir` — value conversion
- `internal/shim.ConnectStdio` — protocol handshake over a child process
- `internal/session.NewNative` — in-process Go handlers
