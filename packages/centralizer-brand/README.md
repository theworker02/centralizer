<p align="center">
  <img src="logo.svg" width="120" alt="Centralizer logo">
</p>

<h1 align="center">@theworker02/centralizer-brand</h1>
<p align="center"><strong>One runtime. Every language.</strong></p>

Official Centralizer brand assets. The marks in this package are copies of the repository `assets/` directory, which is the single source of truth.

This package is not a JavaScript runtime and does not implement Centralizer Protocol. It exists so sites, READMEs, and npm consumers use the same hub + converging paths / copper node — not a one-off drawing.

## Files

| File | Use |
| --- | --- |
| `logo.svg` | Canonical mark |
| `logo-dark.svg` | Dark-background variant |
| `logo-light.svg` | Light-background variant |
| `icon.svg` | Compact icon / favicon |
| `icon-256.png`, `icon-512.png` | Raster icons |
| `github-social-preview.png` | GitHub social preview (1280×640) |

`package.json` sets `"logo": "logo.svg"`.

## Usage

```js
import { logo, icon } from "@theworker02/centralizer-brand";
```

```js
import logoUrl from "@theworker02/centralizer-brand/logo.svg";
```

After changing files in `assets/`, run `node scripts/sync-brand.mjs` from the repository root.

## License

Apache-2.0. See the repository [LICENSE](https://github.com/theworker02/centralizer/blob/main/LICENSE) and [NOTICE](https://github.com/theworker02/centralizer/blob/main/NOTICE).
