import { pickLocaleValue, type Locale, type LocaleMode } from "@/lib/locale";
import type { Article } from "@/api/articles";

export function articleTitle(
  article: Article,
  mode: LocaleMode,
  defaultLocale: Locale,
  currentLocale: Locale,
): string {
  return (
    pickLocaleValue({ value: { zh: article.zhTitle, en: article.enTitle }, mode, defaultLocale, currentLocale }) ||
    article.zhTitle ||
    article.enTitle ||
    ""
  );
}

export function articleBody(
  article: Article,
  mode: LocaleMode,
  defaultLocale: Locale,
  currentLocale: Locale,
): string {
  return (
    pickLocaleValue({ value: { zh: article.zhBody, en: article.enBody }, mode, defaultLocale, currentLocale }) ||
    article.zhBody ||
    article.enBody ||
    ""
  );
}

export function articleMetaDescription(
  article: Article,
  mode: LocaleMode,
  defaultLocale: Locale,
  currentLocale: Locale,
): string {
  return (
    pickLocaleValue({
      value: { zh: article.zhMetaDescription, en: article.enMetaDescription },
      mode,
      defaultLocale,
      currentLocale,
    }) || ""
  );
}

export function articleExcerpt(body: string, maxLen = 120): string {
  if (!body) return "";
  const text = body.replace(/<[^>]*>/g, "").trim();
  return text.length > maxLen ? text.slice(0, maxLen) + "..." : text;
}

export function formatArticleDate(dateStr: string | null, locale: Locale): string {
  if (!dateStr) return "";
  const tag = locale === "en" ? "en-US" : "zh-CN";
  return new Date(dateStr).toLocaleDateString(tag, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}
