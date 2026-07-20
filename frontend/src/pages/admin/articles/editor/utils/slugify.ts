/** Build a URL-safe slug from a title (CJK preserved). */
export function slugifyTitle(title: string): string {
  return (
    title
      .trim()
      .toLowerCase()
      .replace(/[\s_]+/g, "-")
      .replace(/[^\w\u4e00-\u9fff-]+/g, "")
      .replace(/-+/g, "-")
      .replace(/^-|-$/g, "")
      .slice(0, 80) || `article-${Date.now()}`
  );
}
