export interface TocHeading {
  id: string;
  level: 2 | 3;
  text: string;
}

/** A = hidden; B = collapsible inline; C = inline + desktop sticky sidebar */
export type TocLayout = "none" | "inline" | "full";

const MIN_HEADINGS = 2;
const SIDEBAR_MIN_HEADINGS = 4;
const SIDEBAR_MIN_CHARS = 8000;

function slugify(text: string, index: number): string {
  const base = text
    .trim()
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]+/gu, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 48);
  return base ? `${base}-${index}` : `section-${index}`;
}

export function resolveTocLayout(headingCount: number, plainTextLength: number): TocLayout {
  if (headingCount < MIN_HEADINGS) return "none";
  if (headingCount >= SIDEBAR_MIN_HEADINGS || plainTextLength >= SIDEBAR_MIN_CHARS) {
    return "full";
  }
  return "inline";
}

/**
 * Parse h2/h3 from article HTML, assign stable ids, return processed HTML for rendering.
 */
export function buildArticleToc(html: string): {
  headings: TocHeading[];
  htmlWithIds: string;
  layout: TocLayout;
  plainTextLength: number;
} {
  if (!html?.trim() || typeof document === "undefined") {
    return { headings: [], htmlWithIds: html, layout: "none", plainTextLength: 0 };
  }

  const doc = new DOMParser().parseFromString(html, "text/html");
  const plainTextLength = (doc.body.textContent ?? "").replace(/\s+/g, "").length;
  const headings: TocHeading[] = [];
  const usedIds = new Set<string>();

  doc.body.querySelectorAll("h2, h3").forEach((el, index) => {
    const level = el.tagName === "H2" ? 2 : 3;
    const text = (el.textContent ?? "").trim();
    if (!text) return;

    let id = el.getAttribute("id")?.trim() || slugify(text, index + 1);
    while (usedIds.has(id)) {
      id = `${id}-${index + 1}`;
    }
    usedIds.add(id);
    el.setAttribute("id", id);
    headings.push({ id, level, text });
  });

  const layout = resolveTocLayout(headings.length, plainTextLength);

  return {
    headings,
    htmlWithIds: doc.body.innerHTML,
    layout,
    plainTextLength,
  };
}
