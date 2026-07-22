/**
 * Package type-check surface for `@inkless/theme-host`.
 *
 * The real host facade re-exports runtime modules under `@/` and is resolved
 * by the frontend Vite alias. For isolated package `tsc`, map the import to
 * this thin type-only re-export so we do not type-check the entire SPA graph.
 *
 * Runtime / host bundling still uses `frontend/src/theme-host/index.ts`.
 */
import type {
  ComponentType,
  CSSProperties,
  ReactElement,
  ReactNode,
} from "react";

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
export type {
  LayoutConfig,
  HeaderConfig,
  FooterConfig,
} from "@/theme/layouts/types";

/** Minimal branding view used by chrome components during package type-check. */
export interface BrandingView {
  siteName: string;
  tagline: string;
  logo: { light: string; dark?: string };
  favicon: string;
  primaryColor: string;
  author: {
    name: string;
    avatar?: string;
    bio: string;
    socials: unknown[];
  };
  footer: {
    copyright: string;
    icp?: string;
    extraLinks: unknown[];
  };
  localeMode: string;
  defaultLocale: string;
  currentLocale: string;
}

export interface ThemeNavItem {
  label: string;
  path: string;
  sortOrder?: number;
  target?: "_self" | "_parent" | "_blank" | "_top";
  children?: ThemeNavItem[];
}

export interface ThemePagesView {
  pages: unknown[];
  unifiedPages: unknown[];
  headerNavItems: ThemeNavItem[];
  footerNavItems: ThemeNavItem[];
  menuNavItems: ThemeNavItem[];
  isLoading: boolean;
}

export interface GlobalConfigView {
  config: {
    footer?: Record<string, unknown>;
    nav?: { items?: Array<{ label?: string; href?: string }> };
  };
  features: unknown;
}

export interface BaseSiteHeaderProps {
  config?: unknown;
  variant: "corporate" | "blog";
  brand: ReactNode;
  utilities?: ReactNode;
  containerClassName?: string;
  containerStyle?: CSSProperties;
  headerClassName?: string;
  navPaddingClassName?: string;
  languagePlacement?: "top-bar" | "inline" | "none";
  showMobileLanguagePanel?: boolean;
  scrolled?: boolean;
  sticky?: boolean;
}

/** Runtime stubs — values exist only so package tsc accepts value imports. */
export function useBranding(): BrandingView {
  return {
    siteName: "",
    tagline: "",
    logo: { light: "" },
    favicon: "",
    primaryColor: "",
    author: { name: "", bio: "", socials: [] },
    footer: { copyright: "", extraLinks: [] },
    localeMode: "bilingual",
    defaultLocale: "zh",
    currentLocale: "zh",
  };
}

export function useThemePages(): ThemePagesView {
  return {
    pages: [],
    unifiedPages: [],
    headerNavItems: [],
    footerNavItems: [],
    menuNavItems: [],
    isLoading: false,
  };
}

export function useGlobalConfig(): GlobalConfigView {
  return { config: {}, features: {} };
}

export function useHeaderScroll(_enabled: boolean): boolean {
  return false;
}

export const BaseSiteHeader: ComponentType<BaseSiteHeaderProps> = () =>
  null as unknown as ReactElement;

export const ProductPoweredBy: ComponentType<{ className?: string }> = () =>
  null as unknown as ReactElement;
