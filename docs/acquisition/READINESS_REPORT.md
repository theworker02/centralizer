# Acquisition Readiness Report — Centralizer

**Date:** 2026-09-21  
**No numeric score.** Statuses reflect evidence available in-repo and this program.

| Section | Status | Notes |
|---------|--------|-------|
| BUILD | READY | Verified in TEST_EVIDENCE.md (this program) |
| TESTS | READY | Verified in TEST_EVIDENCE.md (this program) |
| SECURITY | READY_WITH_DISCLOSURE | No SECRET_FOUND. Localhost-default daemon; process/shim launch surfaces documented. |
| DOCUMENTATION | READY_WITH_DISCLOSURE | Data room created this program |
| IP OWNERSHIP | REQUIRES_LEGAL_REVIEW | LICENSE: theworker02. NOTICE previously said 'The Centralizer Authors' + Apache (fixed this program)… |
| LICENSE CLARITY | READY_WITH_DISCLOSURE | Current LICENSE clear; history documented; ETW revocation language corrected if applicable |
| DEPENDENCIES | READY_WITH_DISCLOSURE | Go: gopkg.in/yaml.v3 (MIT+Apache). Website: React/Vite MIT; typescript Apache-2.0; caniuse-lite CC-B… |
| THIRD-PARTY ASSETS | READY_WITH_DISCLOSURE / REQUIRES_LEGAL_REVIEW | See diligence |
| DATA RIGHTS | REQUIRES_LEGAL_REVIEW | Especially federated/operator/vendor data |
| REPRODUCIBILITY | READY_WITH_DISCLOSURE | BUYER_DEMO provided |
| TRANSFERABILITY | READY_WITH_DISCLOSURE | See TRANSFER_MANIFEST |
| OPERATIONS | READY_WITH_DISCLOSURE | Handoff + ops docs |
| BUYER DEMO | READY_WITH_DISCLOSURE | Commands verified where stack runnable; see TEST_EVIDENCE |
| KNOWN LIABILITIES | READY_WITH_DISCLOSURE | See DISCLOSURE_SCHEDULE |

## Blockers

### Before outreach
- Stale Apache README/NOTICE/website copy (fixed this program)
- CHANGELOG missing relicense note

### Before diligence
- Apache→proprietary + proxy.golang.org v0.1.2
- Protocol IP packaging
- Adapter completeness honesty

### Before signing
- Formal IP assignment
- Trademark status of Centralizer mark

### Before closing
- Go module path ownership
- GitHub Pages site transfer
- SPA/APA
