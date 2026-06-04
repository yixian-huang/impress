/** Strip HTML and count CJK + Latin words (aligned with admin editor sidebar). */
export function countWordsFromText(text: string): number {
  const cjk = (text.match(/[\u4e00-\u9fff\u3400-\u4dbf\uf900-\ufaff]/g) || []).length;
  const latin = text
    .replace(/[\u4e00-\u9fff\u3400-\u4dbf\uf900-\ufaff]/g, " ")
    .trim()
    .split(/\s+/)
    .filter(Boolean).length;
  return cjk + latin;
}

export function stripHtmlToText(html: string): string {
  if (typeof document !== "undefined") {
    const doc = new DOMParser().parseFromString(html, "text/html");
    return doc.body.textContent?.replace(/\s+/g, " ").trim() ?? "";
  }
  return html.replace(/<[^>]+>/g, " ").replace(/\s+/g, " ").trim();
}

export interface ArticleReadingStats {
  charCount: number;
  wordCount: number;
  /** Estimated minutes at ~280 CJK-or-word units per minute */
  readingMinutes: number;
}

export function articleReadingStatsFromHtml(
  html: string,
  wordsPerMinute = 280,
): ArticleReadingStats {
  const text = stripHtmlToText(html);
  const charCount = text.length;
  const wordCount = countWordsFromText(text);
  const wpm = wordsPerMinute > 0 ? wordsPerMinute : 280;
  const readingMinutes = Math.max(1, Math.ceil(wordCount / wpm));
  return { charCount, wordCount, readingMinutes };
}
