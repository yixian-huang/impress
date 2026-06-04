import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { getPublicCategories } from "@/api/articles";
import type { Category } from "@/api/articles";
import SeoHead from "@/components/SeoHead";
import BlogPageShell from "@/components/blog/BlogPageShell";
import BlogPageHeader from "@/components/blog/BlogPageHeader";
import { useSEODefaults } from "@/hooks/useSEODefaults";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import { pickLocalizedName } from "@/components/blog/pickLocalizedName";
import { useIsReadingLayout } from "@/plugins/hooks";

export default function CategoriesPage() {
  const { t } = useTranslation("common");
  const { buildTitle, defaultDescription } = useSEODefaults();
  const { currentLocale } = useLocaleMode();
  const isReading = useIsReadingLayout();
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const pageTitle = t("blog.categoriesPageTitle");
  const pageDesc = t("blog.categoriesPageDesc");

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const data = await getPublicCategories();
        setCategories(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load categories");
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  return (
    <>
      <SeoHead
        title={buildTitle(pageTitle)}
        description={pageDesc || defaultDescription}
        ogTitle={pageTitle}
        ogDescription={pageDesc || defaultDescription}
        ogType="website"
        canonicalUrl="/categories"
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
        ) : categories.length === 0 ? (
          <p className="text-on-surface-muted py-10 text-center">{t("status.empty")}</p>
        ) : (
          <ul className="divide-y divide-border/80">
            {categories.map((category) => {
              const name = pickLocalizedName(category.zhName, category.enName, currentLocale);
              const desc =
                currentLocale === "en" && category.enDescription
                  ? category.enDescription
                  : category.zhDescription || category.enDescription;
              return (
                <li key={category.id}>
                  <Link
                    to={`/categories/${category.slug}`}
                    className={`block group ${isReading ? "py-6 first:pt-0" : "py-5"}`}
                  >
                    <h2 className="text-xl font-heading font-normal text-on-surface group-hover:text-primary transition-colors">
                      <span className="group-hover:underline decoration-border underline-offset-4">
                        {name}
                      </span>
                    </h2>
                    {desc && (
                      <p className="mt-2 text-sm text-on-surface-muted leading-relaxed line-clamp-2">
                        {desc}
                      </p>
                    )}
                  </Link>
                </li>
              );
            })}
          </ul>
        )}
      </BlogPageShell>
    </>
  );
}
