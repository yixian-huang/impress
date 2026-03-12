import { useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useSearch } from "@/hooks/useSearch";
import SearchBox from "@/components/feature/SearchBox";

/** Sanitize FTS5 snippet: escape all HTML except <mark> tags */
function sanitizeSnippet(html: string): string {
  const parts = html.split(/(<\/?mark>)/gi);
  return parts
    .map((part) => {
      if (part.toLowerCase() === "<mark>" || part.toLowerCase() === "</mark>") {
        return part.toLowerCase();
      }
      return part.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
    })
    .join("");
}

export default function SearchPage() {
  const [searchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const page = Number(searchParams.get("page") ?? "1");
  const { results, loading, search } = useSearch();
  const { t } = useTranslation();

  useEffect(() => {
    if (query) search(query, "", page);
  }, [query, page, search]);

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <SearchBox className="mb-8" />
      {loading && (
        <p className="text-gray-500">{t("search.loading", "Searching...")}</p>
      )}
      {results && (
        <>
          <p className="text-sm text-gray-500 mb-4">
            {t("search.results_count", {
              count: results.total,
              defaultValue: "{{count}} results found",
            })}
          </p>
          <div className="space-y-6">
            {results.results.map((r) => (
              <Link key={`${r.type}-${r.id}`} to={r.url} className="block group">
                <h3 className="text-lg font-medium text-blue-700 group-hover:underline">
                  {r.title}
                </h3>
                <p
                  className="text-sm text-gray-600 mt-1"
                  dangerouslySetInnerHTML={{
                    __html: sanitizeSnippet(r.snippet),
                  }}
                />
                <span className="text-xs text-gray-400">{r.url}</span>
              </Link>
            ))}
          </div>
          {results.total > results.pageSize && (
            <div className="flex justify-center gap-2 mt-8">
              {Array.from(
                { length: Math.ceil(results.total / results.pageSize) },
                (_, i) => (
                  <Link
                    key={i}
                    to={`/search?q=${encodeURIComponent(query)}&page=${i + 1}`}
                    className={`px-3 py-1 rounded ${
                      i + 1 === results.page
                        ? "bg-blue-600 text-white"
                        : "bg-gray-100 hover:bg-gray-200"
                    }`}
                  >
                    {i + 1}
                  </Link>
                )
              )}
            </div>
          )}
        </>
      )}
      {results && results.total === 0 && (
        <p className="text-gray-500">{t("search.no_results", "No results found")}</p>
      )}
    </div>
  );
}
