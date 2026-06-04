/** Role of the article body font stack (maps to theme `font-heading` serif or `font-sans`). */
export type BodyFontRole = "serif" | "sans";

export type FontPresetRole = "serif" | "sans" | "mono";

export interface CustomFontRef {
  url: string;
  family: string;
  weight?: number;
  style?: "normal" | "italic";
}

/** Per-article override stored in `article.metadata.typography`. */
export interface ArticleTypographyOverride {
  /** When true, use `bodyFontRole` below instead of theme default. */
  enabled?: boolean;
  bodyFontRole?: BodyFontRole;
}

export interface FontPreset {
  id: string;
  role: FontPresetRole;
  name: string;
  nameZh: string;
  stack: string;
  /** Optional Google Fonts CSS URL (self-host recommended for production). */
  googleCssUrl?: string;
}

export interface ArticleTypographyConfig {
  bodyFontRole: BodyFontRole;
  bodyFontStack: string;
  /** Article title / in-body headings — always serif stack by default. */
  titleFontStack: string;
  uiFontStack: string;
  monoFontStack: string;
  bodySize: string;
  bodyLineHeight: number;
  customFonts: CustomFontRef[];
}

export interface ThemeFontSources {
  sansPresetId?: string;
  headingPresetId?: string;
  monoPresetId?: string;
  headingUpload?: CustomFontRef;
  sansUpload?: CustomFontRef;
}
