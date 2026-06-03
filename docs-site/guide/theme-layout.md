# Theme layout & chrome

Impress follows the same split as Ghost and headless CMS products:

| Layer | Responsibility |
|-------|----------------|
| **Site Config + Menus** | Data: site name, logo URL, author, navigation links |
| **Theme plugin** | Presentation: Header/Footer components, `defaultLayout`, theme settings, home page mapping |
| **SiteLayout** | Shell: reads active theme and renders chrome around page content |

**Active theme is the single source of truth** for site presentation (corporate vs blog-first layout, chrome, and `/` content). Features only gates public routes and blog toggles (RSS, comments)—not site mode.

## Active theme drives chrome

Each built-in theme registers layout chrome in its plugin definition:

- [`corporate-classic`](../../frontend/src/plugins/themes/corporate-classic/index.ts) — `CorporateHeader` / `CorporateFooter`, wide layout
- [`blog-first`](../../frontend/src/plugins/themes/blog-first/index.ts) — `BlogHeader` / `BlogFooter`, narrow reading layout

[`SiteLayout`](../../frontend/src/theme/layouts/PublicLayout.tsx) (also exported as `PublicLayout`) resolves:

```text
activeTheme.layoutChrome.Header/Footer
activeTheme.defaultLayout  (header/footer style, layout type)
```

Pages should not hardcode header/footer config unless overriding the theme default.

## Shared chrome utilities

Under [`frontend/src/theme/layouts/chrome/`](../../frontend/src/theme/layouts/chrome/):

- `useSiteNavigation()` — Menus → theme pages → legacy global nav, with Features gating
- `useBranding()` — logo, site name, author from Site Config
- `useHeaderSettings()` — merges theme `settingSchema` defaults with Site Config **Header** tab
- `BrandMark` — text / logo / avatar / none brand area

## Configuring blog-first header

1. **Theme settings** (Admin → Theme → Settings): default brand mode, RSS, socials
2. **Site Config → Header**: override brand mode and utility toggles
3. **Site Config → Author / Brand**: name, avatar, logo URL
4. **Menus**: primary navigation links
5. **Features**: enable `/blog`, RSS feed

Theme defaults apply when Site Config Header fields are empty.

## Adding a new theme

1. Create `plugins/themes/my-theme/chrome/MyHeader.tsx` and `MyFooter.tsx`
2. Register in theme plugin:

```ts
layoutChrome: { Header: MyHeader, Footer: MyFooter },
defaultLayout: { type: "default", header: { style: "sticky" }, footer: { style: "minimal" } },
```

3. Optionally add `settingSchema` for theme-specific toggles (Ghost-style custom settings)

No changes to `SiteLayout` are required if chrome components accept `HeaderChromeProps` / `FooterChromeProps`.
