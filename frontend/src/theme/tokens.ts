export interface ThemeTypographyArticle {
  bodySize?: string;
  bodyLineHeight?: number;
}

export interface ThemeFontSources {
  sansPresetId?: string;
  headingPresetId?: string;
  monoPresetId?: string;
  headingUpload?: { url: string; family: string; weight?: number; style?: "normal" | "italic" };
  sansUpload?: { url: string; family: string; weight?: number; style?: "normal" | "italic" };
}

export interface ThemeTokens {
  colors: {
    primary: string;
    primaryDark: string;
    accent: string;
    accentHover: string;
    surface: string;
    surfaceAlt: string;
    onPrimary: string;
    onSurface: string;
    onSurfaceMuted: string;
    border: string;
  };
  fonts: {
    sans: string;
    heading: string;
    mono?: string;
  };
  fontSources?: ThemeFontSources;
  typography?: {
    article?: ThemeTypographyArticle;
  };
  layout: {
    maxWidth: string;
    borderRadius: string;
    contentPadding: string;
    sectionSpacing: string;
    contentGap: string;
  };
}

export const defaultTokens: ThemeTokens = {
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
    mono: 'ui-monospace, "SF Mono", Menlo, monospace',
  },
  typography: {
    article: {
      bodySize: "1.0625rem",
      bodyLineHeight: 1.8,
    },
  },
  layout: {
    maxWidth: "1400px",
    borderRadius: "0.5rem",
    contentPadding: "1.5rem",
    sectionSpacing: "5rem",
    contentGap: "2rem",
  },
};
