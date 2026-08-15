# Centralizer documentation site

Dark graphite / copper static site (Vite, React, TypeScript).

```bash
cd website
npm install
npm run dev
```

Build: `npm run build` (also `make website` from the repository root).

Hero, nav, and favicon use the official mark from `assets/` (`logo.svg`, `icon.svg`). Vite copies those files into `public/` on `buildStart`. The site depends on `@theworker02/centralizer-brand` (`file:../packages/centralizer-brand`) so the same files are available on npm. `package.json` sets `"logo": "../assets/logo.svg"`.

Vite `base` is `./`, so the same `dist/` works locally and on GitHub project Pages at https://theworker02.github.io/centralizer/. Enable Pages with **Settings → Pages → Source: GitHub Actions**. There is no custom domain / CNAME.
