# Python adapter (Tier 1)

| Capability | Status |
| --- | --- |
| Detect | Yes |
| Call | Yes (stdio or localhost TCP shim) |
| Stream | Yes (generators / iterables via STREAM_*) |
| Handles | Yes (`HANDLE_CREATE` / `HANDLE_RELEASE`) |
| Native embed | Not implemented |
| Process | Yes |
| WASM | N/A |

The adapter writes a disposable protocol shim into the Centralizer cache
and launches the installed CPython interpreter. User projects are not
modified.
