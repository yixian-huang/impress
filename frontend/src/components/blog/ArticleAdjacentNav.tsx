import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { getPublicArticles, type Article } from "@/api/articles";
import { useIsReadingLayout } from "@/plugins/hooks";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import { articleTitle } from "@/utils/articleLocale";

interface Adjacent {
  prev: Article | null;
  next: Article | null;
}

const ADJACENT_PAGE_SIZE = 100;

/**
 * Previous / next posts by public list order (newest first).
 * prev = older (list index + 1), next = newer (list index - 1).
 */
export default function ArticleAdjacentNav({ currentSlug }: { currentSlug: string }) {
  const { t } = useTranslation("common");
  const isReading = useIsReadingLayout();
  const { localeMode, defaultLocale, currentLocale } = useLocaleMode();
  const [adj, setAdj] = useState<Adjacent>({ prev: null, next: null });

  useEffect(() => {
    let cancelled = false;
    getPublicArticles(1, ADJACENT_PAGE_SIZE)
      .then((data) => {
        if (cancelled) return;
        const items = data.items || [];
        const idx = items.findIndex((a) => a.slug === currentSlug);
        if (idx < 0) {
          setAdj({ prev: null, next: null });
          return;
        }
        setAdj({
          next: idx > 0 ? items[idx - 1] : null,
          prev: idx < items.length - 1 ? items[idx + 1] : null,
        });
      })
      .catch(() => {
        if (!cancelled) setAdj({ prev: null, next: null });
      });
    return () => {
      cancelled = true;
    };
  }, [currentSlug]);

  if (!adj.prev && !adj.next) return null;

  const labelFor = (article: Article) =>
    articleTitle(article, localeMode, defaultLocale, currentLocale);

  return (
    <nav
      className={
        isReading
          ? "mt-10 pt-8 border-t border-border/80 article-page-ui font-sans"
          : "mt-10 pt-8 border-t border-border"
      }
      aria-label={t("blog.adjacentNav")}
    >
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
        <div className="min-w-0">
          {adj.prev ? (
            <Link
              to={`/blog/${adj.prev.slug}`}
              className="group block text-left hover:opacity-90 transition-opacity"
            >
              <span className="block text-[11px] uppercase tracking-[0.14em] text-on-surface-muted mb-1.5">
                ← {t("blog.prevPost")}
              </span>
              <span className="block text-sm sm:text-[0.95rem] text-on-surface group-hover:text-primary transition-colors line-clamp-2 leading-snug">
                {labelFor(adj.prev)}
              </span>
            </Link>
          ) : (
            <span className="block text-[11px] uppercase tracking-[0.14em] text-on-surface-muted/40">
              ← {t("blog.prevPost")}
            </span>
          )}
        </div>
        <div className="min-w-0 sm:text-right">
          {adj.next ? (
            <Link
              to={`/blog/${adj.next.slug}`}
              className="group block sm:text-right hover:opacity-90 transition-opacity"
            >
              <span className="block text-[11px] uppercase tracking-[0.14em] text-on-surface-muted mb-1.5">
                {t("blog.nextPost")} →
              </span>
              <span className="block text-sm sm:text-[0.95rem] text-on-surface group-hover:text-primary transition-colors line-clamp-2 leading-snug">
                {labelFor(adj.next)}
              </span>
            </Link>
          ) : (
            <span className="block text-[11px] uppercase tracking-[0.14em] text-on-surface-muted/40">
              {t("blog.nextPost")} →
            </span>
          )}
        </div>
      </div>
    </nav>
  );
}
