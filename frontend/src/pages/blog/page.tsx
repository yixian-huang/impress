import { useState, useEffect, useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { getPublicArticles } from "@/api/articles";
import type { Article } from "@/api/articles";
import BlogLayout from "@/theme/layouts/BlogLayout";
import SeoHead from "@/components/SeoHead";
import { useSEODefaults } from "@/hooks/useSEODefaults";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import {
  articleTitle,
  articleBody,
  articleExcerpt,
  formatArticleDate,
} from "@/utils/articleLocale";

export default function BlogPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { t } = useTranslation("common");
  const { buildTitle, defaultDescription } = useSEODefaults();
  const { localeMode, defaultLocale, currentLocale } = useLocaleMode();

  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 9;

  const categoryFilter = searchParams.get("category") || undefined;
  const tagFilter = searchParams.get("tag") || undefined;

  const loadArticles = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getPublicArticles(page, pageSize, categoryFilter, tagFilter);
      setArticles(data.items || []);
      setTotal(data.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load articles");
    } finally {
      setLoading(false);
    }
  }, [page, categoryFilter, tagFilter]);

  useEffect(() => {
    loadArticles();
  }, [loadArticles]);

  const totalPages = Math.ceil(total / pageSize);

  const handleCategoryClick = (slug: string) => {
    const params = new URLSearchParams(searchParams);
    if (params.get("category") === slug) {
      params.delete("category");
    } else {
      params.set("category", slug);
    }
    params.delete("page");
    setSearchParams(params);
    setPage(1);
  };

  const handleTagClick = (slug: string) => {
    const params = new URLSearchParams(searchParams);
    if (params.get("tag") === slug) {
      params.delete("tag");
    } else {
      params.set("tag", slug);
    }
    params.delete("page");
    setSearchParams(params);
    setPage(1);
  };

  const listTitle = t("blog.archiveTitle", "Blog");
  const listDesc = t("blog.archiveDescription", "All posts");

  return (
    <BlogLayout>
      <SeoHead
        title={buildTitle(listTitle)}
        description={listDesc || defaultDescription}
        ogTitle={listTitle}
        ogDescription={listDesc || defaultDescription}
        ogType="website"
        canonicalUrl="/blog"
      />
      <div className="max-w-6xl mx-auto px-4 py-12">
        <header className="mb-10">
          <h1 className="text-3xl font-bold text-gray-900">{listTitle}</h1>
          <p className="mt-2 text-gray-600">{listDesc}</p>
        </header>

        {(categoryFilter || tagFilter) && (
          <div className="mb-6 flex items-center gap-2 flex-wrap">
            <span className="text-sm text-gray-500">{t("blog.filters", "Filters")}:</span>
            {categoryFilter && (
              <button
                type="button"
                onClick={() => handleCategoryClick(categoryFilter)}
                className="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm hover:bg-blue-200"
              >
                {categoryFilter}
                <span className="ml-1">&times;</span>
              </button>
            )}
            {tagFilter && (
              <button
                type="button"
                onClick={() => handleTagClick(tagFilter)}
                className="inline-flex items-center gap-1 px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm hover:bg-green-200"
              >
                {tagFilter}
                <span className="ml-1">&times;</span>
              </button>
            )}
          </div>
        )}

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-800">
            {error}
          </div>
        )}

        {loading ? (
          <div className="flex items-center justify-center h-64">
            <div className="text-gray-600">{t("loading", "Loading...")}</div>
          </div>
        ) : articles.length === 0 ? (
          <div className="flex items-center justify-center h-64">
            <p className="text-gray-500">{t("blog.noPosts", "No posts yet.")}</p>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {articles.map((article) => {
                const title = articleTitle(article, localeMode, defaultLocale, currentLocale);
                const body = articleBody(article, localeMode, defaultLocale, currentLocale);
                return (
                  <article
                    key={article.id}
                    className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow cursor-pointer"
                    onClick={() => navigate(`/blog/${article.slug}`)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") navigate(`/blog/${article.slug}`);
                    }}
                    role="button"
                    tabIndex={0}
                  >
                    {article.coverImage && (
                      <div className="aspect-video bg-gray-100 overflow-hidden">
                        <img
                          src={article.coverImage}
                          alt={title}
                          className="w-full h-full object-cover"
                          loading="lazy"
                        />
                      </div>
                    )}
                    <div className="p-5">
                      <div className="flex items-center gap-2 mb-3 flex-wrap">
                        {article.category && (
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleCategoryClick(article.category!.slug);
                            }}
                            className="text-xs px-2 py-0.5 bg-blue-50 text-blue-700 rounded hover:bg-blue-100"
                          >
                            {article.category.zhName || article.category.enName}
                          </button>
                        )}
                        {article.tags?.map((tag) => (
                          <button
                            key={tag.id}
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleTagClick(tag.slug);
                            }}
                            className="text-xs px-2 py-0.5 bg-gray-100 text-gray-600 rounded hover:bg-gray-200"
                          >
                            {tag.zhName || tag.enName}
                          </button>
                        ))}
                      </div>
                      <h2 className="text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
                        {title}
                      </h2>
                      <p className="text-sm text-gray-600 mb-3 line-clamp-3">
                        {articleExcerpt(body)}
                      </p>
                      <div className="text-xs text-gray-400">
                        {formatArticleDate(article.publishedAt || article.createdAt, currentLocale)}
                      </div>
                    </div>
                  </article>
                );
              })}
            </div>

            {totalPages > 1 && (
              <div className="mt-8 flex items-center justify-center gap-2">
                <button
                  type="button"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="px-4 py-2 text-sm border rounded-lg hover:bg-gray-50 disabled:opacity-50"
                >
                  {t("pagination.prev", "Previous")}
                </button>
                <span className="text-sm text-gray-600 px-4">
                  {page} / {totalPages}
                </span>
                <button
                  type="button"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                  className="px-4 py-2 text-sm border rounded-lg hover:bg-gray-50 disabled:opacity-50"
                >
                  {t("pagination.next", "Next")}
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </BlogLayout>
  );
}
