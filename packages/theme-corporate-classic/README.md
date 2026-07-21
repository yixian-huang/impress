# `@inkless/theme-corporate-classic`

Inkless **corporate-classic** theme — consulting / company site layout, chrome, tokens, and seven-page IA (`home`, `about`, `advantages`, `core-services`, `cases`, `experts`, `contact`).

- **Theme id**: `corporate-classic` (stable; do not rename without a data migration)
- **Contract**: v1 (`@inkless/theme-host`)
- **Hardcoded page components** live in the host app (`frontend/src/pages/*`); attach loaders via `createCorporateClassicTheme({ loaders })` or use the host built-in re-export.

## Host consumption

```ts
import { createCorporateClassicTheme } from "@inkless/theme-corporate-classic";

export const corporateClassicTheme = createCorporateClassicTheme({
  loaders: {
    home: () => import("@/pages/home/page"),
    about: () => import("@/pages/about/page"),
    // ...
  },
});
```

## Workspace

Monorepo package under `packages/theme-corporate-classic` (same pattern as blog-first / product-first before independent GitHub repos).
