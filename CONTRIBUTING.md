# Contributing

Thank you for helping build Centralizer. The project values stability, clear behavior, and debuggable automation over additional language coverage.

## Development

```bash
git clone https://github.com/theworker02/centralizer.git
cd centralizer
go test ./...
make lint
```

Use `gofmt`. Do not add dependencies unless the standard library is insufficient. Public APIs must not panic for ordinary failures.

## Pull requests

- Target `main`. This repository does not use `master`.
- Keep changes focused.
- Include tests for new behavior.
- Update documentation and the compatibility matrix when a capability becomes real.
- Do not mark a language or transport complete unless CI exercises it.
- Use conventional commits: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:`.

## Adapter contributions

New adapters should:

1. Implement `adapter.Adapter`.
2. Live under `adapters/<name>/` with a README stating current transports.
3. Avoid duplicating planner, protocol, or CIR logic.
4. Return `ErrUnsupportedTarget` when the tree is not theirs.
5. Ship a fixture under `examples/` or `tests/` if invocation is claimed.

Discuss large adapters in an issue first (`adapter_request` template).

## Discussions

Use GitHub Issues for defects and adapter requests. Use Discussions for design questions. Do not publish unpatched security issues; see [SECURITY.md](SECURITY.md).

## Code of conduct

Participants follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
