import { useState, useEffect, useCallback } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { getPublicArticles } from "@/api/articles";
import type { Article } from "@/api/articles";
import BlogLayout from "@/theme/layouts/BlogLayout";
import SeoHead from "@/components/SeoHead";
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { useSEODefaults } from "@/hooks/useSEODefaults";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import { pickLocaleValue } from "@/lib/locale";
import { SITE_CONFIG_GLOBAL_DEFAULT } from "@/types/siteConfig";
import {
  articleTitle,
  articleBody,
  articleExcerpt,
  formatArticleDate,
} from "@/utils/articleLocale";

const HOME_RECENT_COUNT = 6;

export default function BlogHomePage() {
  const navigate = useNavigate();
  const { t } = useTranslation("common");
  const { config } = useGlobalConfig();
  const { buildTitle, defaultDescription, defaultOgImage } = useSEODefaults();
  const { localeMode, defaultLocale, currentLocale } = useLocaleMode();

  const siteConfig = config.siteConfig ?? SITE_CONFIG_GLOBAL_DEFAULT;
  const siteName = pickLocaleValue({
    value: siteConfig.identity.name,
    mode: localeMode,
    defaultLocale,
    currentLocale,
  });
  const authorName = siteConfig.author?.name?.trim() || siteName;
  const bio = pickLocaleValue({
    value: siteConfig.author?.bio,
    mode: localeMode,
    defaultLocale,
    currentLocale,
  });
  const tagline = pickLocaleValue({
    value: siteConfig.identity.tagline,
    mode: localeMode,
    defaultLocale,
    currentLocale,
  });
  const intro = bio || tagline || defaultDescription;

  const [articles, setArticles] = useState<Article[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadRecent = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getPublicArticles(1, HOME_RECENT_COUNT);
      setArticles(data.items || []);
      setTotal(data.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load articles");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadRecent();
  }, [loadRecent]);

  const pageTitle = buildTitle(siteName || t("blog.homeTitle", "Home"));

  return (
    <BlogLayout>
      <SeoHead
        title={pageTitle}
        description={intro}
        ogTitle={siteName}
        ogDescription={intro}
        ogImage={siteConfig.brand.ogImage || defaultOgImage}
        ogType="website"
        canonicalUrl="/"
      />
      <div className="max-w-3xl mx-auto px-4 py-12 md:py-16">
        <header className="mb-12">
          {siteConfig.author?.avatar && (
            <img
              src={siteConfig.author.avatar}
              alt={authorName}
              className="w-20 h-20 rounded-full object-cover mb-4"
            />
          )}
          <h1 className="text-3xl md:text-4xl font-bold text-gray-900 tracking-tight">
            {authorName}
          </h1>
          {tagline && bio && (
            <p className="mt-2 text-lg text-gray-600">{tagline}</p>
          )}
          {intro && (
            <p className="mt-4 text-gray-700 leading-relaxed whitespace-pre-wrap">{intro}</p>
          )}
        </header>

        <section>
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-gray-900">
              {t("blog.recentPosts", "Recent posts")}
            </h2>
            {total > HOME_RECENT_COUNT && (
              <Link to="/blog" className="text-sm text-blue-600 hover:text-blue-800">
                {t("blog.viewAll", "View all")} →
              </Link>
            )}
          </div>

          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-800">
              {error}
            </div>
          )}

          {loading ? (
            <div className="py-12 text-center text-gray-500">{t("loading", "Loading...")}</div>
          ) : articles.length === 0 ? (
            <p className="text-gray-500 py-8">{t("blog.noPosts", "No posts yet.")}</p>
          ) : (
            <ul className="divide-y divide-gray-200">
              {articles.map((article) => {
                const title = articleTitle(article, localeMode, defaultLocale, currentLocale);
                const body = articleBody(article, localeMode, defaultLocale, currentLocale);
                return (
                  <li key={article.id} className="py-6 first:pt-0">
                    <button
                      type="button"
                      onClick={() => navigate(`/blog/${article.slug}`)}
                      className="text-left w-full group"
                    >
                      <time className="text-sm text-gray-500">
                        {formatArticleDate(article.publishedAt || article.createdAt, currentLocale)}
                      </time>
                      <h3 className="mt-1 text-lg font-semibold text-gray-900 group-hover:text-blue-700">
                        {title}
                      </h3>
                      <p className="mt-2 text-gray-600 line-clamp-2">
                        {articleExcerpt(body)}
                      </p>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}

          {total > 0 && total <= HOME_RECENT_COUNT && (
            <p className="mt-6">
              <Link to="/blog" className="text-sm text-blue-600 hover:text-blue-800">
                {t("blog.archive", "Archive")} →
              </Link>
            </p>
          )}
        </section>
      </div>
    </BlogLayout>
  );
}
