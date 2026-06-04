import type { ThemeTokens } from "@/theme/tokens";
import { DEFAULT_MONO_PRESET_ID, getFontPreset } from "./presets";
import type {
  ArticleTypographyConfig,
  ArticleTypographyOverride,
  BodyFontRole,
  CustomFontRef,
} from "./types";

const DEFAULT_BODY_SIZE = "1.0625rem";
const DEFAULT_BODY_LINE_HEIGHT = 1.8;

function mergeCustomFont(base: string, upload?: CustomFontRef): { stack: string; custom: CustomFontRef[] } {
  if (!upload?.family || !upload.url) {
    return { stack: base, custom: [] };
  }
  const quoted = upload.family.includes(" ") ? `"${upload.family}"` : upload.family;
  return {
    stack: `${quoted}, ${base}`,
    custom: [upload],
  };
}

export interface ResolveTypographyInput {
  tokens: ThemeTokens;
  /** Installed-theme config flattened keys, e.g. `article.bodyFontRole`. */
  themeSettings?: Record<string, unknown>;
  articleOverride?: ArticleTypographyOverride | null;
}

/**
 * Resolve article typography from design tokens + `article.bodyFontRole` only.
 * Font stacks, size, and line height come from tokens; per-article override may change body role only.
 */
export function resolveArticleTypography(input: ResolveTypographyInput): ArticleTypographyConfig {
  const { tokens, themeSettings = {}, articleOverride } = input;
  const sources = tokens.fontSources ?? {};

  const themeBodyRole =
    (themeSettings["article.bodyFontRole"] as BodyFontRole | undefined) ?? "serif";

  const bodySize = tokens.typography?.article?.bodySize ?? DEFAULT_BODY_SIZE;
  const bodyLineHeight = tokens.typography?.article?.bodyLineHeight ?? DEFAULT_BODY_LINE_HEIGHT;

  const overrideActive = articleOverride?.enabled === true;
  const bodyFontRole = overrideActive
    ? (articleOverride?.bodyFontRole ?? themeBodyRole)
    : themeBodyRole;

  const serifResolved = mergeCustomFont(tokens.fonts.heading, sources.headingUpload);
  const sansResolved = mergeCustomFont(tokens.fonts.sans, sources.sansUpload);
  const monoStack =
    tokens.fonts.mono ?? getFontPreset(DEFAULT_MONO_PRESET_ID)?.stack ?? "ui-monospace, monospace";

  const bodyMerge =
    bodyFontRole === "serif"
      ? serifResolved
      : sansResolved;

  const customFonts = [...serifResolved.custom, ...sansResolved.custom].filter(
    (f, i, arr) => arr.findIndex((x) => x.url === f.url && x.family === f.family) === i,
  );

  return {
    bodyFontRole,
    bodyFontStack: bodyMerge.stack,
    titleFontStack: serifResolved.stack,
    uiFontStack: sansResolved.stack,
    monoFontStack: monoStack,
    bodySize,
    bodyLineHeight,
    customFonts,
  };
}

export function parseArticleTypographyOverride(metadata: Record<string, unknown> | undefined): ArticleTypographyOverride | null {
  if (!metadata?.typography || typeof metadata.typography !== "object") {
    return null;
  }
  return metadata.typography as ArticleTypographyOverride;
}

export function typographyToCssVars(config: ArticleTypographyConfig): Record<string, string> {
  return {
    "--article-font-body": config.bodyFontStack,
    "--article-font-title": config.titleFontStack,
    "--article-font-ui": config.uiFontStack,
    "--article-font-mono": config.monoFontStack,
    "--article-size-body": config.bodySize,
    "--article-leading-body": String(config.bodyLineHeight),
  };
}
