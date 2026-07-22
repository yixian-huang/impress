/**
 * Ambient types for standalone theme development / CI / UMD extract.
 * Runtime: host provides window.InklessThemeHost / @inkless/theme-host.
 *
 * Monorepo package type-check may still map to src/__typecheck__/theme-host.ts;
 * this shim is the extract-ready surface (same role as blog-first).
 */
declare module "@inkless/theme-host" {
  import type { ComponentType, CSSProperties, ReactNode } from "react";

  export const THEME_CONTRACT_VERSION: string;
  export const THEME_CONTRACT_SUPPORTED: readonly string[];
  export const SITE_CONFIG_GLOBAL_DEFAULT: Record<string, unknown>;

  export type HeaderBrandMode = "text" | "logo" | "avatar" | "none";

  export type BrandingView = {
    siteName: string;
    tagline: string;
    logo: { light: string; dark?: string };
    favicon: string;
    primaryColor: string;
    author: {
      name: string;
      avatar?: string;
      bio: string;
      socials: Array<{ kind: string; url: string; label?: string }>;
    };
    footer: { copyright: string; icp?: string; extraLinks?: unknown[] };
    localeMode: string;
    defaultLocale: string;
    currentLocale: string;
  };

  export type ThemeTokens = {
    colors?: Record<string, string>;
    fonts?: Record<string, string>;
    layout?: Record<string, string>;
    [key: string]: unknown;
  };

  export type ThemePlugin = {
    manifest: Record<string, unknown>;
    contractVersion?: string;
    defaultTokens: ThemeTokens;
    pages: unknown[];
    [key: string]: unknown;
  };

  export type HeaderChromeProps = {
    config?: { style?: string; logo?: string; [key: string]: unknown };
  };
  export type FooterChromeProps = {
    config?: { style?: string; copyright?: string; [key: string]: unknown };
  };
  export type ThemePageDefinition = unknown;
  export type ThemeLayoutChrome = unknown;
  export type ThemeSettingGroup = unknown;
  export type TokenPreset = unknown;
  export type LayoutConfig = unknown;
  export type HeaderConfig = unknown;
  export type FooterConfig = unknown;

  export type ThemeNavItem = {
    label: string;
    path: string;
    sortOrder?: number;
    target?: string;
    children?: ThemeNavItem[];
  };

  export function useBranding(): BrandingView;
  export function useThemePages(): {
    pages: unknown[];
    unifiedPages: unknown[];
    headerNavItems: ThemeNavItem[];
    footerNavItems: ThemeNavItem[];
    menuNavItems: ThemeNavItem[];
    isLoading: boolean;
  };
  export function useGlobalConfig(): {
    config: Record<string, unknown>;
    features?: unknown;
    locale?: string;
  };
  export function useHeaderScroll(enabled: boolean): boolean;
  export function useContentMaxWidth(): string;
  export function useIsReadingLayout(): boolean;
  export function useIsThemeHomePath(): boolean;
  export function useHeaderSettings(): {
    brandMode: HeaderBrandMode;
    showRssLink: boolean;
    showSocials: boolean;
  };
  export function pickLocaleValue(input: unknown): string;

  export const BaseSiteHeader: ComponentType<{
    config?: unknown;
    variant?: string;
    brand?: ReactNode;
    utilities?: ReactNode;
    containerClassName?: string;
    containerStyle?: CSSProperties;
    headerClassName?: string;
    navPaddingClassName?: string;
    languagePlacement?: string;
    showMobileLanguagePanel?: boolean;
    scrolled?: boolean;
    sticky?: boolean;
    [key: string]: unknown;
  }>;
  export const ProductPoweredBy: ComponentType<{ className?: string }>;
  export const BrandMark: ComponentType<any>;
  export const HeaderUtilities: ComponentType<any>;
}
