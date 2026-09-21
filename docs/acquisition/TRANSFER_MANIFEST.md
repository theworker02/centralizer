# Transfer Manifest — Centralizer

**Date:** 2026-09-21

| Asset | Category | Notes |
|-------|----------|-------|
| repository | TRANSFERABLE | github.com/theworker02/centralizer |
| Go module path | TRANSFERABLE_WITH_CONSENT | Import path tied to GitHub org/user |
| protocol v1 docs | TRANSFERABLE | Project-defined; historical Apache copies remain Apache |
| adapters | TRANSFERABLE | Original; host language runtimes are third-party |
| website | TRANSFERABLE | Subject to npm dep licenses |
| brand assets | TRANSFERABLE | Registration UNKNOWN |
| secrets | NONTRANSFERABLE | Rotate |

## Credentials migration checklist (no secrets committed)

- [ ] Inventory GitHub secrets / Actions secrets
- [ ] Inventory cloud API tokens (Cloudflare, etc.)
- [ ] Inventory package registry tokens
- [ ] Inventory signing keys
- [ ] Rotate all of the above at closing — **ROTATE_IMMEDIATELY** if any exposure suspected
- [ ] Buyer creates replacement secrets in buyer-controlled accounts

**NEVER commit credentials.**
