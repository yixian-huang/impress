import { pinyin } from "pinyin-pro";

/**
 * Convert a free-text title into a URL-safe Latin slug.
 * - English / Latin titles are lowercased and hyphenated
 * - Chinese (and mixed) titles are romanized via pinyin (no tones)
 * - Non [a-z0-9] characters become hyphens
 */
export function titleToLatinSlug(title: string): string {
  const raw = (title || "").trim();
  if (!raw) return "";

  let romanized = raw;
  try {
    // CJK → pinyin syllables; consecutive non-Chinese runs kept intact
    const parts = pinyin(raw, {
      toneType: "none",
      type: "array",
      nonZh: "consecutive",
      v: true,
    }) as string[];
    romanized = parts.join(" ");
  } catch {
    romanized = raw;
  }

  return romanized
    .toLowerCase()
    .normalize("NFKD")
    // Strip combining diacritics (é → e) for other Latin scripts
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 80);
}

/**
 * Build a default letter-path slug for an article.
 * Prefers English title when it yields a usable Latin slug; otherwise
 * romanizes the Chinese (or primary) title via pinyin.
 */
export function slugifyTitle(zhTitle: string, enTitle?: string): string {
  const candidates = [enTitle, zhTitle]
    .map((t) => (t || "").trim())
    .filter(Boolean);

  for (const candidate of candidates) {
    const slug = titleToLatinSlug(candidate);
    if (slug) return slug;
  }

  return `article-${Date.now()}`;
}
