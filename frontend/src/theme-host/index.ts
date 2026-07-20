/**
 * Stable host surface for Inkless themes (built-in packages + external UMD).
 *
 * Themes must import host APIs only from `@inkless/theme-host` (or this module
 * when resolved via Vite alias). Do not deep-import `@/…` from theme packages.
 *
 * Runtime: also published on `window.__INKLESS_SHARED__.host` for UMD bundles.
 */

// --- Types ---
export type {
  ThemePlugin,
  ThemePageDefinition,
  ThemeLayoutChrome,
  HeaderChromeProps,
  FooterChromeProps,
  ThemeSettingGroup,
  TokenPreset,
} from "@/plugins/types";
export type { ThemeTokens } from "@/theme/tokens";
export type { LayoutConfig, HeaderConfig, FooterConfig } from "@/theme/layouts/types";
export type { HeaderBrandMode } from "@/types/siteConfig";
export type { BrandingView } from "@/hooks/useBranding";

// --- Layout defaults / chrome primitives ---
export { BLOG_DEFAULT_LAYOUT } from "@/theme/layouts/defaults";
export {
  BaseSiteHeader,
  BrandMark,
  HeaderUtilities,
  useHeaderSettings,
} from "@/theme/layouts/chrome";

// --- Hooks ---
export { useBranding } from "@/hooks/useBranding";
export {
  useContentMaxWidth,
  useIsReadingLayout,
  useIsThemeHomePath,
} from "@/plugins/hooks";
export { useGlobalConfig } from "@/contexts/GlobalConfigContext";
export { useSEODefaults } from "@/hooks/useSEODefaults";
export { useLocaleMode } from "@/hooks/useLocaleMode";

// --- Public UI primitives ---
export { default as SeoHead } from "@/components/SeoHead";
export { default as BlogPageShell } from "@/components/blog/BlogPageShell";
export { default as AuthorIntro } from "@/components/blog/AuthorIntro";
export { default as ArticleList } from "@/components/blog/ArticleList";
export { default as AuthorSocialLinks } from "@/components/blog/AuthorSocialLinks";
export { default as ArticleAdjacentNav } from "@/components/blog/ArticleAdjacentNav";

// --- Data / i18n helpers ---
export { getPublicArticles } from "@/api/articles";
export { pickLocaleValue } from "@/lib/locale";
export { SITE_CONFIG_GLOBAL_DEFAULT } from "@/types/siteConfig";
