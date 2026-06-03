import type { Article } from "@/api/articles";
import {
  articleTitle,
  articleBody,
  articleExcerpt,
  formatArticleDate,
} from "@/utils/articleLocale";
import type { LocaleMode, Locale } from "@/lib/locale";

interface ArticleListProps {
  articles: Article[];
  localeMode: LocaleMode;
  defaultLocale: Locale;
  currentLocale: Locale;
  onSelect: (slug: string) => void;
  loading?: boolean;
  loadingLabel?: string;
  emptyLabel?: string;
}

function ArticleListSkeleton() {
  return (
    <ul className="divide-y divide-border" aria-hidden="true">
      {[0, 1, 2].map((key) => (
        <li key={key} className="py-6 first:pt-0 animate-pulse">
          <div className="h-4 w-24 bg-surface-alt rounded" />
          <div className="mt-3 h-5 w-3/4 bg-surface-alt rounded" />
          <div className="mt-2 h-4 w-full bg-surface-alt rounded" />
        </li>
      ))}
    </ul>
  );
}

export default function ArticleList({
  articles,
  localeMode,
  defaultLocale,
  currentLocale,
  onSelect,
  loading = false,
  loadingLabel = "Loading...",
  emptyLabel = "No posts yet.",
}: ArticleListProps) {
  if (loading) {
    return (
      <div>
        <p className="sr-only">{loadingLabel}</p>
        <ArticleListSkeleton />
      </div>
    );
  }

  if (articles.length === 0) {
    return <p className="text-on-surface-muted py-8">{emptyLabel}</p>;
  }

  return (
    <ul className="divide-y divide-border">
      {articles.map((article) => {
        const title = articleTitle(article, localeMode, defaultLocale, currentLocale);
        const body = articleBody(article, localeMode, defaultLocale, currentLocale);
        return (
          <li key={article.id} className="py-6 first:pt-0">
            <button
              type="button"
              onClick={() => onSelect(article.slug)}
              className="text-left w-full group"
            >
              <time className="text-sm text-on-surface-muted">
                {formatArticleDate(article.publishedAt || article.createdAt, currentLocale)}
              </time>
              <h3 className="mt-1 text-lg font-semibold text-on-surface group-hover:text-primary transition-colors">
                {title}
              </h3>
              <p className="mt-2 text-on-surface-muted line-clamp-2">
                {articleExcerpt(body)}
              </p>
            </button>
          </li>
        );
      })}
    </ul>
  );
}
