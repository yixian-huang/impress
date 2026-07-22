# Editorial Firm Theme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a new built-in theme `editorial-firm` — four dynamic pages, magazine/atelier chrome, and an `ef-*` section library — without changing `corporate-classic` or production 印迹 sites.

**Architecture:** Monorepo package `@inkless/theme-editorial-firm` owns tokens, chrome, sections, and seed config data. Host registers the theme, merges sections via existing `useSectionRegistry`, and seeds theme pages from `pages.json`. Unified page published configs (sections) are seeded on first activate when empty. Contract stays v1.

**Tech Stack:** React 19, TypeScript, Vite/pnpm workspace, Tailwind CSS variables, Gin/GORM built-in themes, Vitest.

**Spec:** `docs/superpowers/specs/2026-07-22-editorial-firm-theme-design.md`

## Global Constraints

- Theme id must be exactly `editorial-firm` (stable; used in DB / bootstrap).
- Section types must use prefix `ef-` (no collision with host `hero`, `text-image`, …).
- Do not change `DEFAULT_FALLBACK_THEME_ID` / `DefaultFallbackThemeID` (remain `corporate-classic`).
- Do not migrate or rewrite 印迹 / Blotting content; do not force-activate this theme.
- Theme package must not deep-import `@/…` from the host; only `@inkless/theme-host` peers.
- Neutral bilingual seed copy only — no 印迹 / Blotting brand strings.
- Quality gate after each task that touches product code: relevant tests green; full `pnpm lint && pnpm type-check` before final handoff.
- Commits: one focused commit per completed task (or logical sub-slice).

---

## File map (create / modify)

### Create

| Path | Responsibility |
|------|----------------|
| `packages/theme-editorial-firm/package.json` | Package manifest + inkless metadata |
| `packages/theme-editorial-firm/inkless.theme.json` | Theme declaration for external path |
| `packages/theme-editorial-firm/tsconfig.json` | Package TS config (mirror corporate-classic) |
| `packages/theme-editorial-firm/README.md` | Author docs |
| `packages/theme-editorial-firm/src/index.ts` | `createEditorialFirmTheme`, tokens, exports |
| `packages/theme-editorial-firm/src/register.ts` | UMD-style register helper (optional parity) |
| `packages/theme-editorial-firm/src/tokens.ts` | `ink-editorial` + `noir-gallery` tokens |
| `packages/theme-editorial-firm/src/chrome/EditorialHeader.tsx` | Sticky header chrome |
| `packages/theme-editorial-firm/src/chrome/EditorialFooter.tsx` | Minimal footer chrome |
| `packages/theme-editorial-firm/src/sections/*.tsx` | Eight `ef-*` section components |
| `packages/theme-editorial-firm/src/sections/schemas.ts` | Field schemas for admin (re-exported to host merge) |
| `packages/theme-editorial-firm/src/seed/pageConfigs.ts` | Default four-page section configs (TS data) |
| `frontend/src/plugins/themes/editorial-firm/index.ts` | Host built-in entry |
| `frontend/src/plugins/themes/editorial-firm/contractAlignment.test.ts` | Contract + page meta tests |
| `backend/internal/builtinthemes/editorial_firm_seeds.json` | Optional embedded seed for UnifiedPage configs |
| `docs` touch: package README only (spec already exists) |

### Modify

| Path | Change |
|------|--------|
| `frontend/src/plugins/builtinThemes.ts` | Add `EDITORIAL_FIRM` |
| `frontend/src/plugins/ThemeManagerContext.tsx` | `registerBuiltIn(editorialFirmTheme)` |
| `frontend/src/plugins/builtinThemePages.test.ts` | Assert frontend/backend page metas |
| `frontend/package.json` | `"@inkless/theme-editorial-firm": "workspace:*"` |
| `frontend/tailwind.config.ts` | Scan package `src/**` for utilities |
| `frontend/src/theme/sectionSchemas.ts` | Merge `ef-*` field schemas (admin PropertiesPanel) |
| `package.json` (root) | Optional `build:theme-editorial-firm` script |
| `backend/internal/builtinthemes/constants.go` | `EditorialFirm = "editorial-firm"` |
| `backend/internal/builtinthemes/pages.json` | Four dynamic pages |
| `backend/internal/seed/seed.go` | `SeedInstalledThemes` entry for editorial-firm (inactive) |
| `backend/internal/handler/installed_theme/handler.go` | After `SeedThemePages`, seed empty unified page configs when theme is editorial-firm |
| `backend/internal/service/…` | Small helper to apply seed configs (see Task 6) |

---

### Task 1: Package skeleton + tokens + theme factory (no chrome UI yet)

**Files:**
- Create: `packages/theme-editorial-firm/package.json`
- Create: `packages/theme-editorial-firm/tsconfig.json`
- Create: `packages/theme-editorial-firm/inkless.theme.json`
- Create: `packages/theme-editorial-firm/src/tokens.ts`
- Create: `packages/theme-editorial-firm/src/index.ts`
- Create: `packages/theme-editorial-firm/src/register.ts`
- Create: `packages/theme-editorial-firm/README.md`
- Modify: `frontend/package.json` (workspace dep)
- Modify: `frontend/src/plugins/builtinThemes.ts`
- Modify: `package.json` (optional filter script)

**Interfaces:**
- Produces:
  - `EDITORIAL_FIRM_THEME_ID = "editorial-firm"`
  - `EDITORIAL_FIRM_CONTRACT_VERSION = "1"`
  - `editorialFirmTokens: ThemeTokens` (ink-editorial)
  - `noirGalleryTokens: ThemeTokens`
  - `createEditorialFirmTheme(): ThemePlugin` (pages metadata, tokens, empty sections map for now; chrome placeholders ok as null until Task 2)
  - `EDITORIAL_PAGE_KEYS = ["home","about","services","contact"]`

- [ ] **Step 1: Scaffold package.json and tsconfig**

Mirror `packages/theme-corporate-classic/package.json` with:

```json
{
  "name": "@inkless/theme-editorial-firm",
  "version": "0.1.0",
  "private": false,
  "type": "module",
  "main": "./src/index.ts",
  "exports": {
    ".": "./src/index.ts",
    "./register": "./src/register.ts",
    "./package.json": "./package.json",
    "./inkless.theme.json": "./inkless.theme.json"
  },
  "inkless": {
    "type": "theme",
    "id": "editorial-firm",
    "contractVersion": "1",
    "hostPackage": "@inkless/theme-host"
  },
  "peerDependencies": {
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "react-i18next": "^15.0.0",
    "react-router-dom": "^7.0.0"
  }
}
```

Copy `tsconfig.json` from corporate-classic package; adjust `include` to `["src"]`.

- [ ] **Step 2: Implement tokens**

```ts
// packages/theme-editorial-firm/src/tokens.ts
import type { ThemeTokens } from "@inkless/theme-host";

export const editorialFirmTokens: ThemeTokens = {
  colors: {
    primary: "#111111",
    primaryDark: "#000000",
    accent: "#C45C26",
    accentHover: "#A34A1C",
    surface: "#FAF8F5",
    surfaceAlt: "#F0EBE3",
    onPrimary: "#FFFFFF",
    onSurface: "#1A1A1A",
    onSurfaceMuted: "#5C5C5C",
    border: "#E5DFD6",
  },
  fonts: {
    heading:
      "\"Iowan Old Style\", \"Palatino Linotype\", Palatino, \"Songti SC\", \"Noto Serif SC\", serif",
    sans: 'system-ui, -apple-system, "PingFang SC", "Noto Sans SC", sans-serif',
  },
  layout: {
    maxWidth: "1280px",
    borderRadius: "0.125rem",
    contentPadding: "1.25rem",
    sectionSpacing: "6rem",
    contentGap: "2.5rem",
  },
};

export const noirGalleryTokens: ThemeTokens = {
  colors: {
    primary: "#F5F5F5",
    primaryDark: "#FFFFFF",
    accent: "#E8B86D",
    accentHover: "#D4A017",
    surface: "#0A0A0A",
    surfaceAlt: "#141414",
    onPrimary: "#0A0A0A",
    onSurface: "#F0F0F0",
    onSurfaceMuted: "#A3A3A3",
    border: "#2A2A2A",
  },
  fonts: editorialFirmTokens.fonts,
  layout: editorialFirmTokens.layout,
};
```

If `ThemeTokens` is not exported from `@inkless/theme-host` in this workspace, import the same type path corporate-classic uses (check `packages/theme-corporate-classic` — it imports from `@inkless/theme-host`). Host must already re-export or peer-resolve; follow corporate-classic pattern exactly.

- [ ] **Step 3: Implement `createEditorialFirmTheme` with four dynamic pages**

```ts
// packages/theme-editorial-firm/src/index.ts (core shape)
export const EDITORIAL_FIRM_THEME_ID = "editorial-firm";
export const EDITORIAL_FIRM_CONTRACT_VERSION = "1";
export const EDITORIAL_PAGE_CONTENT_KEYS = ["home", "about", "services", "contact"] as const;

const PAGE_NAV = {
  home: { slug: "home", label: "Home", labelZh: "首页", order: 0 },
  about: { slug: "about", label: "About", labelZh: "关于", order: 1 },
  services: { slug: "services", label: "Services", labelZh: "服务", order: 2 },
  contact: { slug: "contact", label: "Contact", labelZh: "联系", order: 3 },
} as const;

export function createEditorialFirmTheme(): ThemePlugin {
  return {
    manifest: {
      id: EDITORIAL_FIRM_THEME_ID,
      name: "Editorial Firm",
      nameZh: "编辑机构",
      description:
        "Magazine-style firm site: home, about, services, contact — section-driven",
      descriptionZh: "杂志气质机构官网：首页、关于、服务、联系 — 区块配置驱动",
      author: "Inkless CMS",
      version: "0.1.0",
      type: "theme",
      preview: "linear-gradient(135deg, #111111 0%, #C45C26 100%)",
      tags: ["corporate", "editorial", "bilingual", "dynamic"],
    },
    contractVersion: EDITORIAL_FIRM_CONTRACT_VERSION,
    defaultTokens: editorialFirmTokens,
    tokenPresets: [
      {
        id: "ink-editorial",
        name: "Ink Editorial",
        nameZh: "墨色编辑",
        preview: "linear-gradient(135deg, #111111 0%, #C45C26 100%)",
        tokens: editorialFirmTokens,
      },
      {
        id: "noir-gallery",
        name: "Noir Gallery",
        nameZh: "黑白画廊",
        preview: "linear-gradient(135deg, #0A0A0A 0%, #E8B86D 100%)",
        tokens: noirGalleryTokens,
      },
    ],
    pages: EDITORIAL_PAGE_CONTENT_KEYS.map((key) => ({
      slug: PAGE_NAV[key].slug,
      renderMode: "dynamic" as const,
      contentKey: key,
      nav: {
        label: PAGE_NAV[key].label,
        labelZh: PAGE_NAV[key].labelZh,
        order: PAGE_NAV[key].order,
        showInHeader: true,
        showInFooter: true,
      },
    })),
    defaultLayout: {
      type: "default",
      contentProfile: "wide",
      header: { style: "sticky" },
      footer: { style: "minimal" },
    },
    sections: {},
    sectionMetas: [],
    // layoutChrome filled in Task 2
  };
}

export const editorialFirmTheme = createEditorialFirmTheme();
```

- [ ] **Step 4: Wire workspace dependency**

In `frontend/package.json` dependencies:

```json
"@inkless/theme-editorial-firm": "workspace:*"
```

Run: `pnpm install` from repo root.  
Expected: link resolves under `frontend/node_modules/@inkless/theme-editorial-firm`.

- [ ] **Step 5: Add builtin id constant**

```ts
// frontend/src/plugins/builtinThemes.ts
export const BUILTIN_THEME_IDS = {
  CORPORATE_CLASSIC: "corporate-classic",
  BLOG_FIRST: "blog-first",
  PRODUCT_FIRST: "product-first",
  MINIMAL_STARTER: "minimal-starter",
  EDITORIAL_FIRM: "editorial-firm",
} as const;
```

- [ ] **Step 6: Smoke import test (package)**

Add `packages/theme-editorial-firm` type-check:

```bash
pnpm --filter @inkless/theme-editorial-firm type-check
```

Expected: PASS (or fix types until pass).

- [ ] **Step 7: Commit**

```bash
git add packages/theme-editorial-firm frontend/package.json pnpm-lock.yaml frontend/src/plugins/builtinThemes.ts package.json
git commit -m "feat(theme): scaffold editorial-firm package and tokens"
```

---

### Task 2: Editorial chrome (Header + Footer)

**Files:**
- Create: `packages/theme-editorial-firm/src/chrome/EditorialHeader.tsx`
- Create: `packages/theme-editorial-firm/src/chrome/EditorialFooter.tsx`
- Modify: `packages/theme-editorial-firm/src/index.ts` (`layoutChrome`)
- Modify: `frontend/tailwind.config.ts` (content scan)

**Interfaces:**
- Consumes: `useBranding`, `useThemePages`, `useGlobalConfig`, `HeaderChromeProps` / `FooterChromeProps` from `@inkless/theme-host` (same as CorporateHeader/Footer).
- Produces: `EditorialHeader`, `EditorialFooter` assigned to `layoutChrome`.

- [ ] **Step 1: Extend Tailwind content paths**

```ts
// frontend/tailwind.config.ts content array — add:
"./node_modules/@inkless/theme-editorial-firm/src/**/*.{js,ts,jsx,tsx}",
// and for workspace source if linked differently:
"../packages/theme-editorial-firm/src/**/*.{js,ts,jsx,tsx}",
```

Without this, utility classes in the theme will be purged and layout will collapse.

- [ ] **Step 2: Implement EditorialHeader**

Requirements:
- Sticky; on scroll add solid `bg-surface/95 backdrop-blur` border-b.
- Left: logo (`branding.logo.light`) or uppercase `branding.siteName` wordmark (`font-heading tracking-wide`).
- Center/right: links from `useThemePages().headerNavItems` (label + path).
- Locale toggle only if host already exposes one via shared chrome helpers — if no shared toggle component, skip language UI (site still works; i18n via existing app chrome if any). Prefer parity with CorporateHeader if it has locale.
- Mobile: button toggles nav panel; links close panel on click.
- Coerce any accidental object labels with `typeof label === "string" ? label : ""`.

Use host hooks only:

```ts
import {
  useBranding,
  useThemePages,
  type HeaderChromeProps,
} from "@inkless/theme-host";
```

- [ ] **Step 3: Implement EditorialFooter**

Requirements:
- Not full primary-blue bar (classic). Use `bg-surface-alt` + top border.
- Large wordmark / siteName; optional tagline if available from branding.
- Sparse nav from `footerNavItems`.
- Copyright string from `useBranding().footer.copyright` only (already string).
- ICP if present.
- `ProductPoweredBy` if exported from theme-host; if not exported, omit (do not deep-import host component).

- [ ] **Step 4: Attach chrome in factory**

```ts
layoutChrome: {
  Header: EditorialHeader,
  Footer: EditorialFooter,
},
```

- [ ] **Step 5: Manual visual check (dev)**

After later registration (Task 3), verify header/footer render. For this task alone, type-check package:

```bash
pnpm --filter @inkless/theme-editorial-firm type-check
```

- [ ] **Step 6: Commit**

```bash
git add packages/theme-editorial-firm frontend/tailwind.config.ts
git commit -m "feat(theme): editorial-firm header and footer chrome"
```

---

### Task 3: Host registration + backend pages + InstalledTheme seed

**Files:**
- Create: `frontend/src/plugins/themes/editorial-firm/index.ts`
- Create: `frontend/src/plugins/themes/editorial-firm/contractAlignment.test.ts`
- Modify: `frontend/src/plugins/ThemeManagerContext.tsx`
- Modify: `frontend/src/plugins/builtinThemePages.test.ts`
- Modify: `backend/internal/builtinthemes/constants.go`
- Modify: `backend/internal/builtinthemes/pages.json`
- Modify: `backend/internal/seed/seed.go`

**Interfaces:**
- Produces: theme selectable in admin after seed; theme pages meta aligned FE/BE.

- [ ] **Step 1: Host entry**

```ts
// frontend/src/plugins/themes/editorial-firm/index.ts
import { createEditorialFirmTheme } from "@inkless/theme-editorial-firm";

export {
  EDITORIAL_FIRM_THEME_ID,
  EDITORIAL_FIRM_CONTRACT_VERSION,
  createEditorialFirmTheme,
  editorialFirmTokens,
} from "@inkless/theme-editorial-firm";

export const editorialFirmTheme = createEditorialFirmTheme();
```

- [ ] **Step 2: Register built-in**

In `ThemeManagerContext.tsx`:

```ts
import { editorialFirmTheme } from "./themes/editorial-firm";
// ...
themeManager.registerBuiltIn(editorialFirmTheme);
```

- [ ] **Step 3: Backend constant + pages.json**

```go
// constants.go
EditorialFirm = "editorial-firm"
```

Append to `pages.json`:

```json
"editorial-firm": [
  {
    "slug": "home",
    "contentKey": "home",
    "renderMode": "dynamic",
    "sortOrder": 0,
    "title": { "zh": "首页", "en": "Home" },
    "navConfig": { "showInHeader": true, "showInFooter": true }
  },
  {
    "slug": "about",
    "contentKey": "about",
    "renderMode": "dynamic",
    "sortOrder": 1,
    "title": { "zh": "关于", "en": "About" },
    "navConfig": { "showInHeader": true, "showInFooter": true }
  },
  {
    "slug": "services",
    "contentKey": "services",
    "renderMode": "dynamic",
    "sortOrder": 2,
    "title": { "zh": "服务", "en": "Services" },
    "navConfig": { "showInHeader": true, "showInFooter": true }
  },
  {
    "slug": "contact",
    "contentKey": "contact",
    "renderMode": "dynamic",
    "sortOrder": 3,
    "title": { "zh": "联系", "en": "Contact" },
    "navConfig": { "showInHeader": true, "showInFooter": true }
  }
]
```

- [ ] **Step 4: Seed InstalledTheme (inactive)**

In `SeedInstalledThemes`, after product-first / before or after minimal-starter:

```go
if err := s.ensureInstalledTheme(ctx, &model.InstalledTheme{
  ThemeID:     builtinthemes.EditorialFirm,
  Name:        "Editorial Firm",
  NameZh:      "编辑机构",
  Description: "杂志气质机构官网：首页、关于、服务、联系",
  Author:      brand.ProductName,
  Version:     "0.1.0",
  Source:      "built-in",
  IsActive:    false,
  Preview:     "linear-gradient(135deg, #111111 0%, #C45C26 100%)",
}); err != nil {
  return err
}
```

- [ ] **Step 5: Tests**

`contractAlignment.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import {
  EDITORIAL_FIRM_CONTRACT_VERSION,
  EDITORIAL_FIRM_THEME_ID,
} from "@inkless/theme-editorial-firm";
import { THEME_CONTRACT_VERSION } from "@/theme-host/contract";
import { BUILTIN_THEME_IDS } from "@/plugins/builtinThemes";
import { editorialFirmTheme } from "./index";

describe("editorial-firm contract alignment", () => {
  it("manifest id matches builtin constant", () => {
    expect(EDITORIAL_FIRM_THEME_ID).toBe(BUILTIN_THEME_IDS.EDITORIAL_FIRM);
    expect(editorialFirmTheme.manifest.id).toBe(BUILTIN_THEME_IDS.EDITORIAL_FIRM);
  });

  it("targets host contract v1", () => {
    expect(EDITORIAL_FIRM_CONTRACT_VERSION).toBe(THEME_CONTRACT_VERSION);
    expect(editorialFirmTheme.contractVersion).toBe(THEME_CONTRACT_VERSION);
  });

  it("has four dynamic pages", () => {
    expect(editorialFirmTheme.pages).toHaveLength(4);
    for (const page of editorialFirmTheme.pages) {
      expect(page.renderMode).toBe("dynamic");
    }
    expect(editorialFirmTheme.pages.map((p) => p.contentKey)).toEqual([
      "home",
      "about",
      "services",
      "contact",
    ]);
  });

  it("exposes layout chrome", () => {
    expect(editorialFirmTheme.layoutChrome?.Header).toBeTruthy();
    expect(editorialFirmTheme.layoutChrome?.Footer).toBeTruthy();
  });
});
```

Extend `builtinThemePages.test.ts` with editorial-firm FE/BE meta parity (copy corporate-classic test pattern).

- [ ] **Step 6: Run tests**

```bash
cd frontend && pnpm test -- src/plugins/themes/editorial-firm/contractAlignment.test.ts src/plugins/builtinThemePages.test.ts
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add frontend backend/internal/builtinthemes backend/internal/seed/seed.go
git commit -m "feat(theme): register editorial-firm built-in and seed pages meta"
```

---

### Task 4: Section library (P1) — components + metas

**Files:**
- Create: eight section components under `packages/theme-editorial-firm/src/sections/`
- Create: `packages/theme-editorial-firm/src/sections/index.ts`
- Create: `packages/theme-editorial-firm/src/sections/schemas.ts`
- Modify: `packages/theme-editorial-firm/src/index.ts` (`sections`, `sectionMetas`)
- Modify: `frontend/src/theme/sectionSchemas.ts` (merge schemas for admin)

**Interfaces:**
- Produces section types:
  - `ef-hero-editorial`
  - `ef-pull-quote`
  - `ef-feature-split`
  - `ef-service-index`
  - `ef-mosaic`
  - `ef-cta-band`
  - `ef-contact-split`
  - `ef-rich-text`
- Each component: `export default function (props: SectionProps<...>)` compatible with host registry (`data` already locale-resolved by SectionRenderer).

- [ ] **Step 1: Shared layout helpers inside package**

Small internal helper (not exported to host):

```ts
// sections/shell.tsx
export function EfShell({
  children,
  className = "",
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={`max-w-layout mx-auto px-4 md:px-content ${className}`}>
      {children}
    </div>
  );
}
```

Note: SectionRenderer may already wrap with padding/background from `settings`. Prefer **minimal outer padding** inside components when settings.padding is used; use `settings` prop when provided.

- [ ] **Step 2: Implement each section (visual rules from spec)**

| Component | Behavior notes |
|-----------|----------------|
| `EfHeroEditorial` | `layout === "split"` → grid 1fr 1fr; else full-bleed image under/behind title. `font-heading` for title (text-4xl–6xl). CTA is `<a>` if `ctaHref`. |
| `EfPullQuote` | max-w-3xl mx-auto, text-2xl–4xl leading-snug, attribution muted. |
| `EfFeatureSplit` | `imageSide` left/right; image + title + body; caption text-sm muted. |
| `EfServiceIndex` | Numbered `01`… list with border-b rows; optional href per item. |
| `EfMosaic` | CSS grid 2–3 cols; tiles with object-cover aspect. |
| `EfCtaBand` | Full width band `bg-primary text-on-primary` or surface-alt; single button/link. |
| `EfContactSplit` | Two columns: intro + phone/email/address; form fields if `showForm !== false`. Submit: `POST` existing public contact endpoint if present in codebase (search `form_submission` / `/public/contact`); on failure show mailto fallback using `email` prop. |
| `EfRichText` | `prose` or `max-w-prose`; render `body` as plain text with `\n` → paragraphs (no raw HTML unless host already sanitizes elsewhere — **plain text only in v1**). |

All string fields: treat missing as `""`; never render raw objects.

- [ ] **Step 3: Export registry maps**

```ts
// sections/index.ts
export const editorialFirmSections = {
  "ef-hero-editorial": EfHeroEditorial,
  "ef-pull-quote": EfPullQuote,
  "ef-feature-split": EfFeatureSplit,
  "ef-service-index": EfServiceIndex,
  "ef-mosaic": EfMosaic,
  "ef-cta-band": EfCtaBand,
  "ef-contact-split": EfContactSplit,
  "ef-rich-text": EfRichText,
};

export const editorialFirmSectionMetas = [
  { type: "ef-hero-editorial", label: "Editorial Hero", labelZh: "编辑主视觉" },
  // ... all eight
];
```

Wire into `createEditorialFirmTheme`:

```ts
sections: editorialFirmSections,
sectionMetas: editorialFirmSectionMetas,
```

- [ ] **Step 4: Field schemas for admin**

In `schemas.ts`, define `editorialFirmSectionSchemas: Record<string, FieldSchema[]>` using the host `FieldType` vocabulary (`bilingual`, `bilingual-textarea`, `media`, `select`, `array`, `boolean`, `text`).

Merge into host `frontend/src/theme/sectionSchemas.ts`:

```ts
import { editorialFirmSectionSchemas } from "@inkless/theme-editorial-firm";
// or duplicate keys if package export path is awkward — prefer export from package index

export const sectionSchemas: Record<string, FieldSchema[]> = {
  // existing...
  ...editorialFirmSectionSchemas,
};
```

If importing FieldSchema type from host into package is painful, define schemas only in `frontend/src/theme/sectionSchemas.ts` under `ef-*` keys (acceptable for v1; keep component props aligned).

- [ ] **Step 5: Unit smoke tests (optional but recommended)**

One vitest in package or frontend: render `EfPullQuote` with `{ quote: "Hello", attribution: "A" }` and assert text content. Use `@testing-library/react` already in frontend.

- [ ] **Step 6: Type-check + lint**

```bash
pnpm --filter @inkless/theme-editorial-firm type-check
pnpm -C frontend type-check
```

- [ ] **Step 7: Commit**

```bash
git add packages/theme-editorial-firm frontend/src/theme/sectionSchemas.ts
git commit -m "feat(theme): editorial-firm ef-* section library"
```

---

### Task 5: Default seed page configs (data only)

**Files:**
- Create: `packages/theme-editorial-firm/src/seed/pageConfigs.ts`
- Create: `backend/internal/builtinthemes/editorial_firm_seeds.json` (mirror of same data for Go embed)
- Modify: `backend/internal/builtinthemes/embed.go` if needed to export seed bytes

**Interfaces:**
- Produces: pure data maps `{ home, about, services, contact }` each `{ sections: SectionData[] }` with stable string ids (`ef-home-hero`, …).

- [ ] **Step 1: Write TS seed matching design §8**

Example home skeleton:

```ts
export const editorialFirmPageConfigs = {
  home: {
    sections: [
      {
        id: "ef-home-hero",
        type: "ef-hero-editorial",
        data: {
          kicker: { zh: "工作室", en: "Studio" },
          title: { zh: "以编辑之眼，塑造机构叙事", en: "Institutional stories, editorial craft" },
          deck: {
            zh: "为品牌与专业服务机构提供清晰、有分量的线上表达。",
            en: "Clear, substantial web presence for brands and professional firms.",
          },
          layout: "full",
          ctaLabel: { zh: "了解服务", en: "View services" },
          ctaHref: "/services",
          image: "/images/hero-bg.png",
        },
      },
      // feature-split, service-index (3 items), pull-quote, cta-band → /contact
    ],
  },
  // about, services, contact ...
} as const;
```

Use neutral copy only. Prefer existing public image path `/images/hero-bg.png` or empty image string.

- [ ] **Step 2: Mirror JSON for backend**

Same structure in `editorial_firm_seeds.json` for Go to apply on activate.

- [ ] **Step 3: Export from package index for tests**

```ts
export { editorialFirmPageConfigs } from "./seed/pageConfigs";
```

Test: home has ≥4 sections; every section `type` starts with `ef-`.

- [ ] **Step 4: Commit**

```bash
git add packages/theme-editorial-firm backend/internal/builtinthemes
git commit -m "feat(theme): editorial-firm default page section seeds"
```

---

### Task 6: Apply seeds on theme activate (empty UnifiedPage only)

**Files:**
- Modify: `backend/internal/builtinthemes/embed.go` (embed seeds JSON)
- Create or modify: service helper e.g. `backend/internal/service/theme_page_seed_config.go`
- Modify: `backend/internal/handler/installed_theme/handler.go` activate path
- Test: `theme_page_seed_config_test.go` or handler test

**Behavior (strict):**
1. After successful `SeedThemePages` for `editorial-firm`.
2. For each slug in `home|about|services|contact`:
   - Load UnifiedPage by slug.
   - If **not found**, create published page with seed config, titles from pages.json.
   - If found and `PublishedConfig` has **no** `sections` (missing, null, or empty array), apply seed config to draft+published.
   - If found and already has sections → **do not overwrite**.

This protects sites that already edited pages.

- [ ] **Step 1: Embed + parse seeds**

```go
//go:embed editorial_firm_seeds.json
var EditorialFirmSeedsJSON []byte
```

- [ ] **Step 2: Implement `ApplyEditorialFirmPageSeeds(ctx, unifiedPageRepo)`**

Inject `unifiedPageRepo` into installed_theme handler if not already available (check constructor; wire in `main.go` if needed).

- [ ] **Step 3: Call from activate handler**

```go
if target.ThemeID == builtinthemes.EditorialFirm {
  if err := h.applyEditorialFirmSeeds(c.Request.Context()); err != nil {
    // warning only, do not fail activation
  }
}
```

- [ ] **Step 4: Go test**

Table-driven: empty page gets seed; page with existing sections unchanged.

```bash
cd backend && go test ./internal/service/ -count=1 -run EditorialFirm
```

- [ ] **Step 5: Commit**

```bash
git add backend
git commit -m "feat(theme): seed editorial-firm unified pages on activate when empty"
```

---

### Task 7: Polish, a11y, empty state, quality gate

**Files:**
- Possibly: `frontend/src/theme/DynamicPage.tsx` (friendlier empty state when sections length 0)
- Chrome/section small responsive fixes
- `packages/theme-editorial-firm/README.md` complete author guide

- [ ] **Step 1: DynamicPage empty state**

If `config.sections` is empty array, show muted centered message (bilingual via i18n if keys exist, else English/Chinese hardcode minimal): “页面暂无内容 / This page has no content yet” — avoid blank white content area.

- [ ] **Step 2: A11y pass**

- Header nav: `nav` landmark, mobile button `aria-expanded`.
- Images: `alt` from caption/label or empty decorative `alt=""`.
- Contrast: default tokens already dark on paper; verify CTA band text uses `on-primary`.

- [ ] **Step 3: Contact form**

Confirm endpoint with:

```bash
rg -n "form_submission|public/contact|ContactForm" backend frontend/src --glob '*.{go,ts,tsx}' | head -40
```

Wire `EfContactSplit` to the same API ContactFormSection uses; if none, mailto only.

- [ ] **Step 4: Full verification**

```bash
pnpm -C frontend lint
pnpm -C frontend type-check
pnpm -C frontend test:run
cd backend && go test ./internal/builtinthemes/... ./internal/service/ -count=1
```

Expected: all pass; corporate-classic tests still pass.

- [ ] **Step 5: Manual checklist (local dev)**

1. Fresh or existing DB: ensure InstalledTheme row exists (re-seed or insert).
2. Admin → activate Editorial Firm.
3. Public `/` shows hero + sections (not white screen).
4. Nav: Home / About / Services / Contact.
5. Switch back to Corporate Classic → seven-page consulting site still works.

- [ ] **Step 6: Final commit**

```bash
git add -A
git commit -m "feat(theme): polish editorial-firm empty state and a11y"
```

---

### Task 8: Docs close-out

**Files:**
- Modify: `packages/theme-editorial-firm/README.md`
- Modify: `docs/theme-contract.md` — one short bullet under themes list / ownership if there is a theme inventory section

- [ ] **Step 1: README sections**

- Theme id, contract, four pages
- How sections register (`ef-*`)
- How to customize seed
- Explicit: does not replace corporate-classic

- [ ] **Step 2: Commit**

```bash
git commit -am "docs: editorial-firm theme author guide"
```

---

## Spec coverage checklist

| Spec requirement | Task |
|------------------|------|
| New theme id `editorial-firm`, package path | T1 |
| Tokens ink-editorial + noir-gallery | T1 |
| Chrome principles | T2 |
| Host register + pages.json + inactive seed | T3 |
| Four dynamic pages IA | T1, T3 |
| Eight `ef-*` sections + metas | T4 |
| Admin schemas for ef-* | T4 |
| Default seed compositions | T5 |
| Activate applies empty configs only | T6 |
| Isolation / no fallback change | T1/T3 (no constant change) |
| Empty resilience / a11y / tests | T7 |
| README / contract note | T8 |

## Placeholder / consistency self-review

- No TBD steps remaining; contact endpoint discovered at implement time via explicit `rg` step.
- Theme id string is consistently `editorial-firm`.
- Section prefix consistently `ef-`.
- Services slug is `services` (not `core-services`).

---

## Execution handoff

Plan complete and saved to `docs/superpowers/plans/2026-07-22-editorial-firm-theme.md`.

**Two execution options:**

1. **Subagent-Driven (recommended)** — fresh subagent per task, review between tasks, fast iteration  
2. **Inline Execution** — execute tasks in this session with checkpoints  

Which approach?
