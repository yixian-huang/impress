import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { getPublicTags } from "@/api/articles";
import type { Tag } from "@/api/articles";
import SeoHead from "@/components/SeoHead";
import BlogPageShell from "@/components/blog/BlogPageShell";
import BlogPageHeader from "@/components/blog/BlogPageHeader";
import { useSEODefaults } from "@/hooks/useSEODefaults";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import { pickLocalizedName } from "@/components/blog/pickLocalizedName";

export default function TagsPage() {
  const { t } = useTranslation("common");
  const { buildTitle, defaultDescription } = useSEODefaults();
  const { currentLocale } = useLocaleMode();
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const pageTitle = t("blog.tagsPageTitle");
  const pageDesc = t("blog.tagsPageDesc");

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const data = await getPublicTags();
        setTags(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load tags");
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const chipClass =
    "inline-flex items-center px-2 py-1 rounded-sm text-sm font-sans text-on-surface-muted hover:text-primary hover:bg-surface-alt/80 transition-colors";

  return (
    <>
      <SeoHead
        title={buildTitle(pageTitle)}
        description={pageDesc || defaultDescription}
        ogTitle={pageTitle}
        ogDescription={pageDesc || defaultDescription}
        ogType="website"
        canonicalUrl="/tags"
      />
      <BlogPageShell>
        <BlogPageHeader
          title={pageTitle}
          description={pageDesc}
          backTo={{ href: "/blog", label: t("blog.backToArchive") }}
        />

        {loading ? (
          <p className="text-on-surface-muted py-10 text-center">{t("status.loading")}</p>
        ) : error ? (
          <div className="p-4 border border-red-200 rounded-sm text-red-800 bg-red-50">{error}</div>
        ) : tags.length === 0 ? (
          <p className="text-on-surface-muted py-10 text-center">{t("status.empty")}</p>
        ) : (
          <div className="flex flex-wrap gap-2 article-page-ui font-sans">
            {tags.map((tag) => (
              <Link key={tag.id} to={`/tags/${tag.slug}`} className={chipClass}>
                #{pickLocalizedName(tag.zhName, tag.enName, currentLocale)}
              </Link>
            ))}
          </div>
        )}
      </BlogPageShell>
    </>
  );
}
