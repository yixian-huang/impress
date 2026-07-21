import type { ComponentType } from "react";
import type { ThemePlugin, ThemePageDefinition, ThemeTokens } from "@inkless/theme-host";
import CorporateHeader from "./chrome/CorporateHeader";
import CorporateFooter from "./chrome/CorporateFooter";
import StatsCounterSection from "./sections/StatsCounterSection";

/** Theme id — keep in sync with host `BUILTIN_THEME_IDS.CORPORATE_CLASSIC` and DB. */
export const CORPORATE_CLASSIC_THEME_ID = "corporate-classic";

/**
 * Host contract this package targets.
 * Keep in lockstep with host THEME_CONTRACT_VERSION and inkless.theme.json.
 */
export const CORPORATE_CLASSIC_CONTRACT_VERSION = "1";

/** Wide corporate layout (consulting / company site). */
export const CORPORATE_DEFAULT_LAYOUT = {
  type: "default" as const,
  contentProfile: "wide" as const,
  header: { style: "sticky" as const },
  footer: { style: "full" as const },
};

/** Classic blue-green tokens (historical corporate defaults). */
export const corporateClassicTokens: ThemeTokens = {
  colors: {
    primary: "#1a5f8f",
    primaryDark: "#26548b",
    accent: "#8bc34a",
    accentHover: "#7cb342",
    surface: "#ffffff",
    surfaceAlt: "#f9fafb",
    onPrimary: "#ffffff",
    onSurface: "#111827",
    onSurfaceMuted: "#374151",
    border: "#e5e7eb",
  },
  fonts: {
    sans: "system-ui, -apple-system, sans-serif",
    heading: "system-ui, -apple-system, sans-serif",
  },
  layout: {
    maxWidth: "1200px",
    borderRadius: "0.5rem",
    contentPadding: "1.5rem",
    sectionSpacing: "5rem",
    contentGap: "2rem",
  },
};

/** Content keys for the seven corporate pages (backend pages.json SSOT). */
export const CORPORATE_PAGE_CONTENT_KEYS = [
  "home",
  "about",
  "advantages",
  "core-services",
  "cases",
  "experts",
  "contact",
] as const;

export type CorporatePageContentKey = (typeof CORPORATE_PAGE_CONTENT_KEYS)[number];

export type CorporatePageLoaders = Partial<
  Record<CorporatePageContentKey, () => Promise<{ default: ComponentType }>>
>;

const PAGE_NAV: Record<
  CorporatePageContentKey,
  { slug: string; label: string; labelZh: string; order: number }
> = {
  home: { slug: "home", label: "Home", labelZh: "首页", order: 0 },
  about: { slug: "about", label: "About", labelZh: "关于我们", order: 1 },
  advantages: { slug: "advantages", label: "Advantages", labelZh: "我们的优势", order: 2 },
  "core-services": { slug: "core-services", label: "Services", labelZh: "核心服务", order: 3 },
  cases: { slug: "cases", label: "Cases", labelZh: "案例展示", order: 4 },
  experts: { slug: "experts", label: "Experts", labelZh: "专家团队", order: 5 },
  contact: { slug: "contact", label: "Contact", labelZh: "联系我们", order: 6 },
};

function buildPages(loaders: CorporatePageLoaders = {}): ThemePageDefinition[] {
  return CORPORATE_PAGE_CONTENT_KEYS.map((contentKey) => {
    const meta = PAGE_NAV[contentKey];
    const lazy = loaders[contentKey];
    return {
      slug: meta.slug,
      renderMode: "hardcoded" as const,
      contentKey,
      lazyComponent: lazy,
      nav: {
        label: meta.label,
        labelZh: meta.labelZh,
        order: meta.order,
        showInHeader: true,
        showInFooter: true,
      },
    };
  });
}

/**
 * Build the corporate-classic theme plugin.
 *
 * Hardcoded page **components** still live in the host (`frontend/src/pages/*`).
 * Pass `loaders` so the theme can lazy-load them without deep-importing `@/…`.
 */
export function createCorporateClassicTheme(options?: {
  loaders?: CorporatePageLoaders;
}): ThemePlugin {
  const loaders = options?.loaders ?? {};

  return {
    manifest: {
      id: CORPORATE_CLASSIC_THEME_ID,
      name: "Corporate Classic",
      nameZh: "企业经典",
      description:
        "Professional corporate website with homepage, about, advantages, services, cases, experts, and contact pages",
      descriptionZh: "专业企业官网，含首页、关于、优势、服务、案例、专家、联系",
      author: "Inkless CMS",
      version: "1.0.0",
      type: "theme",
      preview: "linear-gradient(135deg, #1a5f8f 0%, #8bc34a 100%)",
      tags: ["corporate", "bilingual", "consulting"],
    },
    contractVersion: CORPORATE_CLASSIC_CONTRACT_VERSION,
    defaultTokens: corporateClassicTokens,
    settingSchema: [
      {
        group: "homepage",
        label: "Homepage",
        labelZh: "首页设置",
        fields: [
          {
            name: "heroStyle",
            type: "select",
            label: "Hero Style",
            labelZh: "主视觉样式",
            defaultValue: "image",
            options: [
              { label: "背景图", value: "image" },
              { label: "纯色渐变", value: "gradient" },
            ],
          },
          {
            name: "showLatestArticles",
            type: "boolean",
            label: "Show Latest Articles",
            labelZh: "显示最新文章",
            defaultValue: true,
          },
          {
            name: "latestArticlesCount",
            type: "number",
            label: "Article Count",
            labelZh: "文章数量",
            defaultValue: 3,
          },
        ],
      },
      {
        group: "footer",
        label: "Footer",
        labelZh: "页脚设置",
        fields: [
          {
            name: "showSocialLinks",
            type: "boolean",
            label: "Show Social Links",
            labelZh: "显示社交链接",
            defaultValue: false,
          },
          {
            name: "icp",
            type: "text",
            label: "ICP Number",
            labelZh: "ICP 备案号",
            defaultValue: "",
          },
        ],
      },
    ],
    tokenPresets: [
      {
        id: "default",
        name: "经典蓝绿",
        nameZh: "经典蓝绿",
        preview: "linear-gradient(135deg, #1a5f8f 0%, #8bc34a 100%)",
        tokens: corporateClassicTokens,
      },
      {
        id: "modern-dark",
        name: "现代暗色",
        nameZh: "现代暗色",
        preview: "linear-gradient(135deg, #6366f1 0%, #22d3ee 100%)",
        tokens: {
          colors: {
            primary: "#6366f1",
            primaryDark: "#4f46e5",
            accent: "#22d3ee",
            accentHover: "#06b6d4",
            surface: "#0f172a",
            surfaceAlt: "#1e293b",
            onPrimary: "#ffffff",
            onSurface: "#e2e8f0",
            onSurfaceMuted: "#94a3b8",
            border: "#334155",
          },
          fonts: {
            sans: "Inter, system-ui, -apple-system, sans-serif",
            heading: "Inter, system-ui, -apple-system, sans-serif",
          },
          layout: {
            maxWidth: "1200px",
            borderRadius: "0.75rem",
            contentPadding: "2rem",
            sectionSpacing: "5rem",
            contentGap: "2rem",
          },
        },
      },
      {
        id: "warm-earth",
        name: "暖色大地",
        nameZh: "暖色大地",
        preview: "linear-gradient(135deg, #92400e 0%, #d97706 100%)",
        tokens: {
          colors: {
            primary: "#92400e",
            primaryDark: "#78350f",
            accent: "#d97706",
            accentHover: "#b45309",
            surface: "#fffbeb",
            surfaceAlt: "#fef3c7",
            onPrimary: "#ffffff",
            onSurface: "#451a03",
            onSurfaceMuted: "#78350f",
            border: "#fde68a",
          },
          fonts: {
            sans: "Georgia, 'Times New Roman', serif",
            heading: "Georgia, 'Times New Roman', serif",
          },
          layout: {
            maxWidth: "1200px",
            borderRadius: "0.25rem",
            contentPadding: "1.25rem",
            sectionSpacing: "4rem",
            contentGap: "1.5rem",
          },
        },
      },
    ],
    sections: {
      "stats-counter": StatsCounterSection,
    },
    sectionMetas: [
      { type: "stats-counter", label: "Stats Counter", labelZh: "数据统计" },
    ],
    pages: buildPages(loaders),
    defaultLayout: CORPORATE_DEFAULT_LAYOUT,
    layoutChrome: {
      Header: CorporateHeader,
      Footer: CorporateFooter,
    },
  };
}

/**
 * Theme shell without host page loaders (UMD / external path).
 * Prefer `createCorporateClassicTheme({ loaders })` for built-in host use.
 */
export const corporateClassicTheme: ThemePlugin = createCorporateClassicTheme();

export {
  CorporateHeader,
  CorporateFooter,
  StatsCounterSection,
};
