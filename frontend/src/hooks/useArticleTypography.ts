import { useEffect, useMemo, useContext } from "react";
import { useThemeSettings } from "@/plugins/hooks";
import { ThemeContext } from "@/theme/ThemeContext";
import { defaultTokens } from "@/theme/tokens";
import {
  loadCustomFonts,
  parseArticleTypographyOverride,
  resolveArticleTypography,
  type ArticleTypographyConfig,
  type ArticleTypographyOverride,
} from "@/theme/typography";

export function useArticleTypography(articleMetadata?: Record<string, unknown> | null): ArticleTypographyConfig {
  const ctx = useContext(ThemeContext);
  const tokens = ctx?.tokens ?? defaultTokens;
  const themeSettings = useThemeSettings();

  const override = useMemo(
    () => parseArticleTypographyOverride(articleMetadata ?? undefined),
    [articleMetadata],
  );

  return useMemo(
    () =>
      resolveArticleTypography({
        tokens,
        themeSettings,
        articleOverride: override,
      }),
    [tokens, themeSettings, override],
  );
}

/** Load uploaded @font-face assets when typography config changes. */
export function useLoadArticleFonts(config: ArticleTypographyConfig): void {
  useEffect(() => {
    loadCustomFonts(config.customFonts);
  }, [config.customFonts]);
}

export type { ArticleTypographyConfig, ArticleTypographyOverride };
