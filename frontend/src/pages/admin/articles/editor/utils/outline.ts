/** Shared outline item for richtext (pos) and markdown (line, 1-based). */

export type OutlineItem = {
  level: number;
  text: string;
  /** TipTap document position (start of heading node) */
  pos?: number;
  /** Markdown 1-based line number */
  line?: number;
};

/** Parse ATX headings (# … ######) from Markdown source. */
export function parseMarkdownOutline(source: string | undefined | null): OutlineItem[] {
  if (!source) return [];
  const items: OutlineItem[] = [];
  const lines = source.split(/\r?\n/);
  // Skip fenced code blocks so # inside code is not treated as heading
  let inFence = false;
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const fence = line.match(/^(`{3,}|~{3,})/);
    if (fence) {
      inFence = !inFence;
      continue;
    }
    if (inFence) continue;
    const m = line.match(/^(#{1,6})\s+(.+?)\s*#*\s*$/);
    if (!m) continue;
    const level = m[1].length;
    const text = m[2].replace(/\s+#+\s*$/, "").trim();
    items.push({ level, text, line: i + 1 });
  }
  return items;
}

/** Extract headings from HTML (fallback when TipTap is unmounted). */
export function parseHtmlOutline(html: string | undefined | null): OutlineItem[] {
  if (!html) return [];
  const items: OutlineItem[] = [];
  const re = /<h([1-6])\b[^>]*>([\s\S]*?)<\/h\1>/gi;
  let m: RegExpExecArray | null;
  while ((m = re.exec(html)) !== null) {
    const level = Number(m[1]);
    const text = m[2].replace(/<[^>]+>/g, "").replace(/&nbsp;/g, " ").trim();
    items.push({ level, text });
  }
  return items;
}
