import { Link, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useIsReadingLayout } from "@/plugins/hooks";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import { useBlogTaxonomy } from "@/hooks/useBlogTaxonomy";
import { pickLocalizedName } from "@/components/blog/pickLocalizedName";
import TaxonomyChip from "@/components/blog/TaxonomyChip";

interface BlogArchiveNavProps {
  activeCategory?: string;
  activeTag?: string;
  onFilterCategory: (slug: string) => void;
  onFilterTag: (slug: string) => void;
}

function SectionHeader({ label, href, linkLabel }: { label: string; href: string; linkLabel: string }) {
  return (
    <div className="flex items-baseline justify-between gap-4 mb-2">
      <span className="text-sm text-on-surface-muted">{label}</span>
      <Link
        to={href}
        className="shrink-0 text-sm text-on-surface-muted hover:text-primary transition-colors"
      >
        {linkLabel}
      </Link>
    </div>
  );
}

export default function BlogArchiveNav({
  activeCategory,
  activeTag,
  onFilterCategory,
  onFilterTag,
}: BlogArchiveNavProps) {
  const { t } = useTranslation("common");
  const [searchParams, setSearchParams] = useSearchParams();
  const isReading = useIsReadingLayout();
  const { currentLocale } = useLocaleMode();
  const { categories, tags } = useBlogTaxonomy();

  if (categories.length === 0 && tags.length === 0) {
    return null;
  }

  if (!isReading) {
    return null;
  }

  const hasActiveFilter = Boolean(activeCategory || activeTag);

  const clearFilters = () => {
    const params = new URLSearchParams(searchParams);
    params.delete("category");
    params.delete("tag");
    params.delete("page");
    setSearchParams(params);
  };

  const activeCategoryName = activeCategory
    ? pickLocalizedName(
        categories.find((c) => c.slug === activeCategory)?.zhName ?? "",
        categories.find((c) => c.slug === activeCategory)?.enName ?? "",
        currentLocale,
      ) || activeCategory
    : null;

  const activeTagName = activeTag
    ? pickLocalizedName(
        tags.find((tg) => tg.slug === activeTag)?.zhName ?? "",
        tags.find((tg) => tg.slug === activeTag)?.enName ?? "",
        currentLocale,
      ) || activeTag
    : null;

  return (
    <nav
      className="mb-8 font-sans article-page-ui space-y-5"
      aria-label={t("blog.browseByTopic")}
    >
      {hasActiveFilter && (
        <div className="flex flex-wrap items-center gap-x-3 gap-y-2 text-sm">
          <span className="text-on-surface-muted">{t("blog.filters")}</span>
          {activeCategory && (
            <TaxonomyChip
              label={activeCategoryName!}
              active
              onClick={() => onFilterCategory(activeCategory)}
            />
          )}
          {activeTag && (
            <TaxonomyChip
              label={activeTagName!}
              active
              prefix="#"
              onClick={() => onFilterTag(activeTag)}
            />
          )}
          <button
            type="button"
            onClick={clearFilters}
            className="text-on-surface-muted hover:text-primary transition-colors"
          >
            {t("blog.clearFilter")}
          </button>
        </div>
      )}

      {categories.length > 0 && (
        <section>
          <SectionHeader
            label={t("blog.categories")}
            href="/categories"
            linkLabel={t("blog.allCategories")}
          />
          <div className="flex flex-wrap gap-1">
            {categories.map((cat) => (
              <TaxonomyChip
                key={cat.id}
                label={pickLocalizedName(cat.zhName, cat.enName, currentLocale)}
                active={activeCategory === cat.slug}
                onClick={() => onFilterCategory(cat.slug)}
              />
            ))}
          </div>
        </section>
      )}

      {tags.length > 0 && (
        <section>
          <SectionHeader label={t("blog.tags")} href="/tags" linkLabel={t("blog.allTags")} />
          <div className="flex flex-wrap gap-1">
            {tags.map((tag) => (
              <TaxonomyChip
                key={tag.id}
                label={pickLocalizedName(tag.zhName, tag.enName, currentLocale)}
                active={activeTag === tag.slug}
                prefix="#"
                onClick={() => onFilterTag(tag.slug)}
              />
            ))}
          </div>
        </section>
      )}
    </nav>
  );
}
