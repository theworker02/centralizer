# Node.js adapter (Tier 1)

| Capability | Status |
| --- | --- |
| Detect | Yes (JavaScript / TypeScript markers) |
| Call | Yes (stdio or localhost TCP shim, JS entry) |
| Stream | Yes (iterables / async iterables via STREAM_*) |
| Handles | Not implemented |
| Process | Yes |
| TypeScript execution | Only if the project already has a JS entry |

TypeScript sources raise detection confidence. Invocation still requires
a Node-loadable JavaScript entry (`package.json` `main` or `index.js`).
