# Adapters

An adapter discovers a target and, if implemented, opens a Bridge. It must not implement planning, supervision, or CIR.

```go
type Adapter interface {
    Name() string
    Detect(context.Context, Target) (Detection, error)
    Capabilities(context.Context, Target) ([]capability.Capability, error)
    Prepare(context.Context, Target) error
    Connect(context.Context, Target, bridge.Plan) (bridge.Bridge, error)
}
```

## Registration

Built-in adapters are registered by `centralizer.New()`. Additional adapters:

```go
hub := centralizer.New(centralizer.WithAdapter(myAdapter))
```

Do not use Go plugins as the sole extension mechanism. A future revision will speak the protocol to out-of-process third-party adapters.

## Lifecycle

1. `Detect` — return confidence 0 or `ErrUnsupportedTarget` if the tree is not yours.
2. `Capabilities` — only claim transports you can actually `Connect`.
3. `Prepare` — generate cache-resident shims if policy allows.
4. `Connect` — return a live `bridge.Bridge`.

## Author checklist

- Detection is independently testable.
- README lists current transports honestly.
- CIR conversion is documented.
- Child processes are owned by the transport and reaped on close.
- No copy of planner/protocol internals.

`centralizer adapters` prints name, tier, invocation, and claimed capabilities.

See `adapters/*/README.md` and `pkg/adapter`.
