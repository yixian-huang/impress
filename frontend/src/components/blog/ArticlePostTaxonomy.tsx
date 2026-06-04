import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { Article } from "@/api/articles";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import { pickLocalizedName } from "@/components/blog/pickLocalizedName";

interface ArticlePostTaxonomyProps {
  article: Article;
}

const linkClass =
  "text-sm text-on-surface-muted hover:text-primary transition-colors";

export default function ArticlePostTaxonomy({ article }: ArticlePostTaxonomyProps) {
  const { t } = useTranslation("common");
  const { currentLocale } = useLocaleMode();

  const hasCategory = Boolean(article.category);
  const hasTags = Boolean(article.tags?.length);

  if (!hasCategory && !hasTags) {
    return null;
  }

  return (
    <section
      className="article-page-ui font-sans mb-8 flex flex-wrap items-center gap-x-4 gap-y-2 text-sm"
      aria-label={t("blog.taxonomyTitle")}
    >
      {hasCategory && article.category && (
        <span className="inline-flex items-center gap-2">
          <span className="text-on-surface-muted">{t("blog.categories")}</span>
          <Link to={`/blog?category=${article.category.slug}`} className={linkClass}>
            {pickLocalizedName(
              article.category.zhName,
              article.category.enName,
              currentLocale,
            )}
          </Link>
        </span>
      )}
      {hasTags && article.tags && article.tags.length > 0 && (
        <span className="inline-flex flex-wrap items-center gap-2">
          <span className="text-on-surface-muted">{t("blog.tags")}</span>
          {article.tags.map((tag) => (
            <Link key={tag.id} to={`/blog?tag=${tag.slug}`} className={linkClass}>
              #{pickLocalizedName(tag.zhName, tag.enName, currentLocale)}
            </Link>
          ))}
        </span>
      )}
    </section>
  );
}
