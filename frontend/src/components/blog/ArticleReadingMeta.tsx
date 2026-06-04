import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useArticleReadingMeta } from "@/hooks/useArticleReadingMeta";
import { articleReadingStatsFromHtml } from "@/utils/articleReadingStats";

interface ArticleReadingMetaProps {
  html: string;
  className?: string;
}

/** Word count + estimated reading time (gated by features + theme settings). */
export default function ArticleReadingMeta({ html, className = "" }: ArticleReadingMetaProps) {
  const { t, i18n } = useTranslation("common");
  const { enabled, wordsPerMinute } = useArticleReadingMeta();
  const stats = useMemo(
    () => articleReadingStatsFromHtml(html, wordsPerMinute),
    [html, wordsPerMinute],
  );

  if (!enabled || stats.wordCount === 0) return null;

  return (
    <span className={`tabular-nums ${className}`.trim()}>
      {t("blog.readingMeta", {
        words: stats.wordCount.toLocaleString(i18n.language),
        minutes: stats.readingMinutes,
      })}
    </span>
  );
}
