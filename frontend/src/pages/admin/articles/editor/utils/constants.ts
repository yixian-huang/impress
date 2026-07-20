export const ALL_LANGS = [
  { key: "zh", label: "中文" },
  { key: "en", label: "English" },
] as const;

export type LangKey = (typeof ALL_LANGS)[number]["key"];

export const AUTOSAVE_DEBOUNCE_MS = 4000;
export const WORD_STATS_DEBOUNCE_MS = 250;
export const TOAST_MS = 3000;

export function statusLabelOf(
  status: "draft" | "published" | "scheduled",
  dirty?: boolean,
): string {
  const base =
    status === "published" ? "已发布" : status === "scheduled" ? "定时" : "草稿";
  return dirty ? `${base} · 含未保存` : base;
}

export function hasMeaningfulHtml(html: string | undefined | null): boolean {
  if (!html || html === "<p></p>") return false;
  return html.replace(/<[^>]+>/g, "").trim().length > 0;
}
