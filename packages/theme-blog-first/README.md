# @inkless/theme-blog-first

Inkless **blog-first** theme: personal-blog home, header/footer chrome, and reading-room tokens.

## Consumed by host (built-in)

Inkless frontend depends on this workspace package and calls:

```ts
import { blogFirstTheme } from "@inkless/theme-blog-first";
themeManager.registerBuiltIn(blogFirstTheme);
```

Theme source imports host APIs only from `@inkless/theme-host` (Vite alias → `frontend/src/theme-host`).

## Remote install (UMD)

```bash
pnpm -C packages/theme-blog-first build
# → dist/theme.umd.js  (and theme.es.js)
```

Host must expose (see `frontend/src/plugins/externals.ts`):

- `window.__INKLESS_SHARED__` — React peers + `host`
- `window.InklessThemeHost` — same as `__INKLESS_SHARED__.host`
- `window.__INKLESS_THEME_REGISTER__(theme)` — registration callback during load

Then admin sets `installed_themes.externalUrl` to the UMD URL.

## Layout

```
src/
  index.ts          # ThemePlugin export
  register.ts       # UMD auto-register entry
  chrome/           # BlogHeader / BlogFooter / brand rules
  pages/home.tsx    # Theme home presentation
inkless.theme.json  # Package manifest for installers
```

## Contract

See monorepo `docs/theme-contract.md`.
