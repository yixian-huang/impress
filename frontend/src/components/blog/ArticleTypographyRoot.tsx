import type { CSSProperties, ReactNode } from "react";
import { useLoadArticleFonts, useArticleTypography } from "@/hooks/useArticleTypography";
import { typographyToCssVars } from "@/theme/typography";

interface ArticleTypographyRootProps {
  children: ReactNode;
  className?: string;
  /** reading = public article; editor = admin preview */
  mode?: "reading" | "editor" | "default";
  articleMetadata?: Record<string, unknown> | null;
  style?: CSSProperties;
}

/**
 * Applies shared article typography CSS variables for public pages and editor preview.
 * `font-heading` in code = serif/body stack; UI chrome uses `--article-font-ui`.
 */
export default function ArticleTypographyRoot({
  children,
  className = "",
  mode = "reading",
  articleMetadata,
  style,
}: ArticleTypographyRootProps) {
  const config = useArticleTypography(articleMetadata);
  useLoadArticleFonts(config);

  const modeClass =
    mode === "reading"
      ? "article-typography article-reading"
      : mode === "editor"
        ? "article-typography article-editor-preview"
        : "article-typography";

  return (
    <div
      className={[modeClass, className].filter(Boolean).join(" ")}
      style={{ ...typographyToCssVars(config), ...style }}
      data-body-font-role={config.bodyFontRole}
    >
      {children}
    </div>
  );
}
