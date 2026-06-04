import { useState, useEffect, useCallback, useRef } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { getPublicArticle } from "@/api/articles";
import type { Article } from "@/api/articles";
import SeoHead from "@/components/SeoHead";
import ArticlePostWithToc from "@/components/blog/ArticlePostWithToc";
import ArticlePostHeader from "@/components/blog/ArticlePostHeader";
import ArticlePostTaxonomy from "@/components/blog/ArticlePostTaxonomy";
import BlogPageShell from "@/components/blog/BlogPageShell";
import ArticleTypographyRoot from "@/components/blog/ArticleTypographyRoot";
import { CommentSlot } from "@/modules/comment";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import {
  articleTitle,
  articleBody,
  articleMetaDescription,
} from "@/utils/articleLocale";

export default function BlogDetailPage() {
  const { slug } = useParams<{ slug: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation("common");
  const { localeMode, defaultLocale, currentLocale } = useLocaleMode();
  const [article, setArticle] = useState<Article | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);
  const contentRef = useRef<HTMLElement>(null);

  useEffect(() => {
    if (!slug) return;

    const load = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await getPublicArticle(slug);
        setArticle(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load article");
      } finally {
        setLoading(false);
      }
    };

    load();
  }, [slug]);

  const handleContentClick = useCallback((e: React.MouseEvent) => {
    const target = e.target as HTMLElement;
    if (target.tagName === "IMG") {
      const src = (target as HTMLImageElement).src;
      if (src) setLightboxSrc(src);
    }
  }, []);

  const title = article
    ? articleTitle(article, localeMode, defaultLocale, currentLocale)
    : "";
  const body = article
    ? articleBody(article, localeMode, defaultLocale, currentLocale)
    : "";
  const metaDesc = article
    ? articleMetaDescription(article, localeMode, defaultLocale, currentLocale)
    : "";

  useEffect(() => {
    if (loading || !article || !body) return;
    const hash = window.location.hash.replace(/^#/, "");
    if (!hash) return;
    const timer = window.setTimeout(() => {
      document.getElementById(hash)?.scrollIntoView({ behavior: "smooth", block: "start" });
    }, 100);
    return () => window.clearTimeout(timer);
  }, [loading, article, body]);

  return (
    <>
      {article && (
        <SeoHead
          title={title}
          description={metaDesc}
          ogTitle={article.zhSeoTitle || article.enSeoTitle || title}
          ogDescription={metaDesc}
          ogImage={article.ogImage || article.coverImage || ""}
          ogType="article"
          canonicalUrl={`/blog/${article.slug}`}
        />
      )}
      <BlogPageShell>
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <div className="text-on-surface-muted">{t("status.loading")}</div>
          </div>
        ) : error || !article ? (
          <div className="text-center">
            <p className="text-red-600 mb-4">{error || t("blog.notFound")}</p>
            <button
              type="button"
              onClick={() => navigate("/blog")}
              className="text-primary hover:text-accent transition-colors"
            >
              {t("blog.backToArchive")}
            </button>
          </div>
        ) : (
          <>
            <ArticleTypographyRoot mode="reading" articleMetadata={article.metadata} className="article-public-view">
              <ArticlePostHeader
                title={title}
                bodyHtml={body}
                article={article}
                currentLocale={currentLocale}
              />

              <ArticlePostTaxonomy article={article} />

              <ArticlePostWithToc html={body} contentRef={contentRef} onClick={handleContentClick} />
            </ArticleTypographyRoot>

            <CommentSlot
              contentType="article"
              contentId={article.id}
              contentAllowed={article.allowComments !== false}
            />
          </>
        )}
      </BlogPageShell>

      {lightboxSrc && (
        <div
          className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/70 cursor-zoom-out"
          onClick={() => setLightboxSrc(null)}
          onKeyDown={(e) => e.key === "Escape" && setLightboxSrc(null)}
          role="presentation"
        >
          <img
            src={lightboxSrc}
            alt=""
            className="max-w-[90vw] max-h-[90vh] object-contain rounded-lg shadow-2xl"
          />
        </div>
      )}
    </>
  );
}
